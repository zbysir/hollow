package hollow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/gookit/color"
	"github.com/gorilla/websocket"
	lru "github.com/hashicorp/golang-lru/v2"
	gojsx "github.com/zbysir/gojsx"
	hollowdev "github.com/zbysir/hollow/front/hollow-dev"
	"github.com/zbysir/hollow/internal/pkg/asynctask"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/easyfs"
	"github.com/zbysir/hollow/internal/pkg/execcmd"
	"github.com/zbysir/hollow/internal/pkg/git"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/http_file_server"
	"github.com/zbysir/hollow/internal/pkg/httpsrv"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/timetrack"
	"github.com/zbysir/hollow/internal/pkg/ws"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"html"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Page map[string]interface{}
type Pages []Page

func (p Page) GetPath() string {
	pa := ""
	switch t := p["path"].(type) {
	case goja.Value:
		pa = t.Export().(string)
	case string:
		pa = p["path"].(string)
	default:
		pa = ""
	}

	// 删除前后的 /
	pa = strings.Trim(pa, "/")
	return pa
}

func tryToVDom(i interface{}) gojsx.VDom {
	switch t := i.(type) {
	case map[string]interface{}:
		return t
	}

	return gojsx.VDom{}
}

func (p Page) GetComponent() (gojsx.VDom, error) {
	var v gojsx.VDom
	switch t := p["component"].(type) {
	case *goja.Object:
		c, ok := gojsx.AssertFunction(t)
		if ok {
			// for: component: () => Index(props)
			val, err := c(goja.Null())
			if err != nil {
				return v, err
			}
			v = tryToVDom(val.Export())
		} else {
			// for: component: Index(props)
			v = tryToVDom(t.Export())
		}
		return v, nil
		//case map[string]interface{}:
		//	// for: component: Index(props)
		//	v = tryToVDom(t)
		//	return v, nil
	}

	return v, fmt.Errorf("uncased value type: %T", p["component"])
}

func (p Page) Render() (string, error) {
	if p["component"] != nil {
		vd, err := p.GetComponent()
		if err != nil {
			return "", err
		}

		replaceAttrDot(vd)

		return vd.Render(), nil
	} else if p["body"] != nil {
		return exportGojaValueToString(p["body"]), nil
	}

	return "", fmt.Errorf("can't render page: %+v", p)
}

type Hollow struct {
	jsx         *gojsx.Jsx
	log         *zap.SugaredLogger
	asyncTask   *asynctask.Manager
	wsHub       *ws.WsHub
	debug       bool
	sourceStdFs fs.FS

	// gojsx 使用了 fs 地址作为缓存 key，这里也缓存上 ThemeLoader 固定 fs 对象地址。
	themeLoaderCache *lru.Cache[string, ThemeLoader]

	Option
}

type RenderContext struct {
	cache *lru.Cache[string, interface{}] // 缓存耗时操作与多次调用的数据，如获取 config、blog 文件夹，加速多个页面渲染相同数据的情况。
	timer *timetrack.TimeTracker
	debug bool
	data  sync.Map
	lock  sync.Mutex
}

func (b *RenderContext) timerStart(span string) func() {
	if !b.debug {
		return func() {}
	}
	return b.timer.Start(span)
}

func (b *RenderContext) GetDataAll(f func(key, value any) bool) {
	b.data.Range(f)
}
func (b *RenderContext) GetData(key string) []interface{} {
	i, ok := b.data.Load(key)
	if ok {
		return i.([]interface{})
	}

	return nil
}

func (b *RenderContext) Save(k string, v interface{}) {
	b.lock.Lock()
	defer b.lock.Unlock()
	exist, ok := b.data.Load(k)
	if ok {
		b.data.Store(k, append(exist.([]interface{}), v))
	} else {
		b.data.Store(k, []interface{}{v})
	}
}

func NewRenderContext() *RenderContext {
	c, _ := lru.New[string, interface{}](100)
	return &RenderContext{
		cache: c,
		timer: &timetrack.TimeTracker{},
		debug: false,
		data:  sync.Map{},
	}
}

type Option struct {
	SourceFs   billy.Filesystem
	FixedTheme string           // 重新指定主题，可用于预览
	CacheFs    billy.Filesystem // 缓存文件系统，默认为 memory
}

type StdFileSystem struct {
}

func (f StdFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func NewHollow(o Option) (*Hollow, error) {
	if o.SourceFs == nil {
		o.SourceFs = osfs.New(".")
	}
	if o.CacheFs == nil {
		o.CacheFs = memfs.New()
	}

	stdFs := gobilly.NewStdFs(o.SourceFs)

	var err error
	jsx, err := gojsx.NewJsx(gojsx.Option{
		Debug: false,
		Fs:    stdFs,
	})
	if err != nil {
		return nil, err
	}

	c, _ := lru.New[string, ThemeLoader](100)
	b := &Hollow{
		jsx:              jsx,
		Option:           o,
		log:              log.Logger(),
		asyncTask:        asynctask.NewManager(),
		wsHub:            ws.NewHub(),
		debug:            os.Getenv("DEBUG") != "",
		sourceStdFs:      stdFs,
		themeLoaderCache: c,
	}

	b.asyncTask.AddListener(func(task *asynctask.Task, event *asynctask.Event) {
		if event.IsDone {
			b.wsHub.Close(task.Key)
		} else {
			b.wsHub.Send(task.Key, []byte(event.Log))
		}
	})

	return b, nil
}

type ExecOption struct {
	Log *zap.SugaredLogger

	IsDev bool // 开发环境每次都会读取最新的文件，而生成环境会缓存
}

// Build 生成静态源文件
func (b *Hollow) Build(ctx *RenderContext, distPath string, o ExecOption) error {
	return b.BuildToFs(ctx, osfs.New(distPath), o)
}

func (b *Hollow) loadTheme(ctx *RenderContext, url string, refresh bool, enableAsync bool) (ThemeExport, fs.FS, *asynctask.Task, error) {
	themeLoader, err := b.getThemeLoader(url)
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}

	//log.Infof("-------- loadTheme -------")
	themeModule, themeFs, task, err := themeLoader.Load(ctx, refresh, enableAsync, gojsx.WithNativeModule("@bysir/hollow", map[string]interface{}{
		"getContents":      b.getContents(ctx),
		"getConfig":        b.getConfig(ctx),
		"getContentDetail": b.getContentDetail(ctx),
		"md":               b.md(ctx),
		"mdx":              b.mdx(ctx),
	}))
	if err != nil {
		return ThemeExport{}, nil, nil, fmt.Errorf("load theme '%s' error: %w", url, err)
	}

	return themeModule, themeFs, task, nil
}

func (b *Hollow) BuildToFs(ctx *RenderContext, dst billy.Filesystem, o ExecOption) error {
	start := time.Now()
	conf, err := b.LoadConfig(ctx)
	if err != nil {
		return fmt.Errorf("LoadConfig error: %w", err)
	}
	themeUrl := b.prepareThemeUrl(conf.Hollow.Theme, b.FixedTheme)

	themeModule, themeFs, _, err := b.loadTheme(ctx, themeUrl, true, false)
	if err != nil {
		return err
	}
	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	l = l.Named("[Build]\t")

	for i, p := range themeModule.Pages {
		name := p.GetPath()
		body, err := p.Render()
		if err != nil {
			return fmt.Errorf("render page '%v' error: %w", name, err)
		}
		var distFile string

		// 如果有扩展名，则说明是文件
		if filepath.Ext(name) != "" {
			distFile = name
		} else {
			// 否则存入文件夹
			distFile = filepath.Join(name, "index.html")
		}

		f, err := dst.Create(distFile)
		if err != nil {
			return fmt.Errorf("create file '%v' error: %w", distFile, err)
		}
		_, err = f.Write([]byte(body))
		if err != nil {
			return err
		}

		l.Infof("Create file [%03d]: %v", i, distFile)
	}

	for _, a := range themeModule.Assets {
		err = copyDir(a, "", themeFs, dst)
		if err != nil {
			return err
		}
		l.Infof("Copy theme assets: %v ", a)
	}

	for _, a := range conf.Hollow.Assets {
		err = copyDir(a, "", b.sourceStdFs, dst)
		if err != nil {
			return err
		}
		l.Infof("Copy assets: %v ", a)
	}

	// copy assets in content
	a := ctx.GetData("assets")
	for _, a := range a {
		for _, a := range a.(Assets) {
			if err = copyFile(a, filepath.Join("__source", a), b.sourceStdFs, dst); err != nil {
				log.Warnf("exportFile error: %s", err)
			}
		}
	}

	l.Infof("Done in %v", time.Now().Sub(start))
	return nil
}

func (b *Hollow) BuildAndPublish(ctx *RenderContext, dst billy.Filesystem, o ExecOption) error {
	err := b.BuildToFs(ctx, dst, o)
	if err != nil {
		return err
	}

	conf, err := b.LoadConfig(ctx)
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	l = l.Named("[Git]\t")

	g, err := git.NewGit(conf.Hollow.Deploy.Token, dst, l)
	if err != nil {
		return err
	}

	branch := conf.Hollow.Deploy.Branch
	if branch == "" {
		branch = "docs"
	}
	err = g.Push(conf.Hollow.Deploy.Remote, branch, "by hollow", true)
	if err != nil {
		return err
	}
	return nil
}

func (b *Hollow) PushProject(o ExecOption) error {
	conf, err := b.LoadConfig(NewRenderContext())
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	g, err := git.NewGit(conf.Hollow.Source.Token, b.SourceFs, l)
	if err != nil {
		return err
	}

	log.Infof("config %+v", conf.Hollow.Source)
	err = g.Push(conf.Hollow.Source.Remote, conf.Hollow.Source.Branch, "-", true)
	if err != nil {
		return err
	}

	return nil
}

func (b *Hollow) PullProject(o ExecOption) error {
	conf, err := b.LoadConfig(NewRenderContext())
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	g, err := git.NewGit(conf.Hollow.Deploy.Token, b.SourceFs, l)
	if err != nil {
		return err
	}

	err = g.Pull(conf.Hollow.Source.Remote, conf.Hollow.Source.Branch, true)
	if err != nil {
		return err
	}

	return nil
}

type DirFs struct {
	appPrefix   string // 访问 url /1.txt 将会 访问 /prefix/1.txt 文件
	stripPrefix string // 访问 url /prefix/1.txt 将会访问 /1.txt
	fs          http.FileSystem
}

func (d *DirFs) Open(name string) (http.File, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := d.appPrefix
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
	fullName = strings.TrimPrefix(fullName, d.stripPrefix)

	f, err := d.fs.Open(fullName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type HollowConfig struct {
	Theme  string    `json:"theme"`
	Deploy GitRepo   `json:"deploy"`
	Source GitRepo   `json:"source"`
	Oss    ConfigOss `json:"oss"`
	Assets Assets    `json:"assets"`
}

type Config struct {
	Hollow HollowConfig
	Theme  ThemeConfig
}

type ThemeConfig interface{}

type GitRepo struct {
	Token  string `json:"token" yaml:"token"`
	Remote string `json:"remote" yaml:"remote"`
	Branch string `json:"branch" yaml:"branch"`
}

type ConfigGit struct {
	Deploy GitRepo `json:"deploy" yaml:"deploy"`
	Source GitRepo `json:"source" yaml:"source"`
}

type ConfigOss struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	//Zone      string `yaml:"zone"`
	Bucket string `yaml:"bucket"`
	Prefix string `yaml:"prefix"`
}

func loadYamlConfig(body string, expandEnv bool) (con Config, err error) {
	if expandEnv {
		body = os.Expand(body, os.Getenv)
	} else {
		body = os.Expand(body, func(s string) string {
			k := os.Getenv(s)
			return strings.Repeat("*", len(k))
		})
	}

	type YamlConfig struct {
		Theme       string      `yaml:"theme"`
		Deploy      GitRepo     `yaml:"deploy"`
		Source      GitRepo     `yaml:"source"`
		Oss         ConfigOss   `yaml:"oss"`
		Assets      Assets      `yaml:"assets"`
		ThemeConfig interface{} `yaml:"theme_config"`
	}

	var yc YamlConfig

	err = yaml.Unmarshal([]byte(body), &yc)
	if err != nil {
		err = fmt.Errorf("LoadConfig error: %w", err)
		return
	}

	con = Config{
		Hollow: HollowConfig{
			Theme:  yc.Theme,
			Deploy: yc.Deploy,
			Source: yc.Source,
			Oss:    yc.Oss,
			Assets: yc.Assets,
		},
		Theme: yc.ThemeConfig,
	}

	return
}

// LoadConfig 加载 source 下的 config 文件
func (b *Hollow) LoadConfig(ctx *RenderContext) (conf Config, err error) {
	cacheKey := fmt.Sprintf("LoadConfig")
	x, ok := ctx.cache.Get(cacheKey)
	if ok {
		return x.(Config), nil
	}
	defer func() {
		if err == nil {
			ctx.cache.Add(cacheKey, conf)
		}
	}()

	var jsConfigExist = false
	for _, f := range []string{"config.ts", "config.js"} {
		_, err := b.SourceFs.Stat(f)
		if err == nil {
			jsConfigExist = true
		}
	}
	if jsConfigExist {
		var exports *gojsx.ModuleExport
		exports, err = b.jsx.ExecCode([]byte(fmt.Sprintf("module.exports = require('./config')")))
		if err != nil {
			return
		}

		bs, _ := json.Marshal(exports.Exports["hollow"])
		json.Unmarshal(bs, &conf.Hollow)

		bs, _ = json.Marshal(exports.Exports["theme"])
		json.Unmarshal(bs, &conf.Theme)

		return
	}

	f, err := easyfs.GetFile(b.sourceStdFs, "config.yml")
	if err != nil {
		return
	}
	if err == nil {
		conf, err = loadYamlConfig(f.Body, true)
		if err != nil {
			return
		}
		return conf, err
	} else if err == os.ErrNotExist {
		err = nil
	} else {
		return
	}

	return
}

type PrepareOpt struct {
	NoCache bool
}

func prepareThemeUrl(url string, defaultProtocol string) string {
	switch {
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"):
		return url
	case strings.HasPrefix(url, "file://"):
		return url
	case strings.HasPrefix(url, "source://"):
		return url
	default:
		return defaultProtocol + url
	}
}

// getThemeLoader 返回主题加载器，支持以下协议的地址。
// https:// : git
// file:// : relative or absolute path, e.g. file://usr/bysir/xx , file://./bysir/xx
// source:// : relative path of source fs
func (b *Hollow) getThemeLoader(url string) (ThemeLoader, error) {
	v, ok := b.themeLoaderCache.Get(url)
	if ok {
		return v, nil
	}

	log.Infof("new ThemeLoader %v", url)

	var tl ThemeLoader
	switch {
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"):
		tl = NewGitThemeLoader(b.asyncTask, url, b.CacheFs)
	case strings.HasPrefix(url, "file://"):
		pa := strings.TrimPrefix(url, "file://")
		if strings.HasPrefix(pa, ".") {
			// 相对路径
		} else {
			// 绝对路径
			pa = "/" + pa
		}
		tl = NewFsThemeLoader(gobilly.NewStdFs(osfs.New(pa)))
	case strings.HasPrefix(url, "source://"):
		pa := strings.TrimPrefix(url, "source://")
		subFs, err := b.SourceFs.Chroot(pa)
		if err != nil {
			return nil, err
		}
		tl = NewFsThemeLoader(gobilly.NewStdFs(subFs))
	default:
		return nil, fmt.Errorf("unsupported protocol, url: %v", url)
	}
	b.themeLoaderCache.Add(url, tl)
	return tl, nil
}

func handleAsyncTask(task string, name string, writer http.ResponseWriter, request *http.Request) {
	body := fmt.Sprintf(`
<html>
<head>
<link href="/_dev_/static/index.css" rel="stylesheet"/>
</head>
<script src="/_dev_/static/index.js"></script>
<script>
window.onload = function(){window.RenderTask({taskKey: '%s', name: '%s'})}
</script>
</html>
`, task, name)
	writer.Write([]byte(body))
}

func handleError(err error, writer http.ResponseWriter, request *http.Request) {
	bs, _ := json.Marshal(err.Error())
	body := fmt.Sprintf(`
<html>
<head>
<link href="/_dev_/static/index.css" rel="stylesheet"/>
</head>
<script src="/_dev_/static/index.js"></script>
<script>
window.onload = function(){ window.RenderError({msg: %s})}
</script>
</html>
`, bs)
	writer.Write([]byte(body))
	writer.WriteHeader(400)
}

func (b *Hollow) prepareThemeUrl(projectTheme string, fixedTheme string) string {
	themeUrl := prepareThemeUrl(projectTheme, "source://")
	if fixedTheme != "" {
		themeUrl = prepareThemeUrl(fixedTheme, "file://")
	}
	return themeUrl
}

func (b *Hollow) ServiceHandle(o ExecOption) func(writer http.ResponseWriter, request *http.Request) {
	var assetsHandler http.Handler
	var themeModule ThemeExport

	prepare := func(ctx *RenderContext, opt *PrepareOpt) (asyncTaskKey string, err error) {
		end := ctx.timerStart("config")

		var projectConf Config

		projectConf, err = b.LoadConfig(ctx)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {

			} else {
				return "", err
			}
		}
		end()

		refresh := false
		if opt != nil && opt.NoCache {
			refresh = true
		}

		themeUrl := b.prepareThemeUrl(projectConf.Hollow.Theme, b.FixedTheme)
		var themeFs fs.FS
		var task *asynctask.Task
		end = ctx.timerStart("theme")
		themeModule, themeFs, task, err = b.loadTheme(ctx, themeUrl, refresh, true)
		if err != nil {
			return "", err
		}
		end()
		if task != nil {
			return task.Key, nil
		}

		var dirs MuitDir
		for _, dir := range themeModule.Assets {
			sub, err := fs.Sub(themeFs, dir)
			if err != nil {
				return "", fmt.Errorf("sub fs '%v' error: %w", dir, err)
			}

			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}

		for _, dir := range projectConf.Hollow.Assets {
			sub, err := fs.Sub(gobilly.NewStdFs(b.SourceFs), dir)
			if err != nil {
				return "", fmt.Errorf("sub fs '%v' error: %w", dir, err)
			}
			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}

		// for file on source content
		dirs = append(dirs, &DirFs{
			stripPrefix: "__source",
			fs:          http.FS(gobilly.NewStdFs(b.SourceFs)),
		})

		assetsHandler = http_file_server.FileServer(dirs)
		return "", nil
	}

	// 不是 dev 环境只会加载一次主题，而不是每次刷新页面都加载
	if !o.IsDev {
		ctx := NewRenderContext()
		task, err := prepare(ctx, nil)
		if err != nil {
			return func(writer http.ResponseWriter, request *http.Request) {
				handleError(err, writer, request)
			}
		}

		if task != "" {
			return func(writer http.ResponseWriter, request *http.Request) {
				handleAsyncTask(task, "", writer, request)
			}
		}
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		reqPath := strings.Trim(request.URL.Path, "/")

		ctx := NewRenderContext()

		if o.IsDev {
			opt := PrepareOpt{
				NoCache: request.Header.Get("Cache-Control") == "no-cache",
			}
			task, err := prepare(ctx, &opt)
			if err != nil {
				handleError(err, writer, request)
				return
			}
			if task != "" {
				handleAsyncTask(task, "", writer, request)
				return
			}
		}
		for _, p := range themeModule.Pages {
			if reqPath == p.GetPath() {
				body, err := p.Render()
				if err != nil {
					handleError(err, writer, request)
					return
				}
				writer.WriteHeader(200)
				writer.Write([]byte(body))
				return
			}
		}
		assetsHandler.ServeHTTP(writer, request)
	}
}

var upgrader = websocket.Upgrader{
	// 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// DevService 运行前端开发逻辑，如 yarn dev
func (b *Hollow) DevService(ctx context.Context) error {
	conf, err := b.LookupConfig(NewRenderContext())
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	if conf.ThemePath != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, err := b.checkPackageJsonScript(conf.ThemePath, "dev")
			if err != nil {
				panic(err)
			}
			if !ok {
				return
			}

			name := "Theme "
			col := color.Green
			log.Infof(col.Render(fmt.Sprintf("Running dev service for [%s] [%s]", name, conf.ThemePath)))
			err = b.runDevServer(ctx, name, col, conf.ThemePath)
			if err != nil {
				log.Errorf(col.Render(fmt.Sprintf("RunDevServer [%v] error: %v", name, err)))
			}
		}()
	}

	if conf.SourcePath != "" && conf.ThemePath != conf.SourcePath {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, err := b.checkPackageJsonScript(conf.SourcePath, "dev")
			if err != nil {
				panic(err)
			}
			if !ok {
				return
			}

			name := "Source"
			col := color.Magenta
			log.Infof(col.Render(fmt.Sprintf("Running dev service for [%v] [%v]", name, conf.SourcePath)))

			err = b.runDevServer(ctx, name, col, conf.SourcePath)
			if err != nil {
				log.Errorf(col.Render(fmt.Sprintf("RunDevServer [%v] error: %v", name, err)))
			}
		}()
	}

	wg.Wait()

	return nil
}

func (b *Hollow) checkPackageJsonScript(dir string, script string) (ok bool, err error) {
	bs, err := os.ReadFile(path.Join(dir, "package.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		err = fmt.Errorf("read package.json error: %w", err)
		return
	}

	var j packageJson
	err = json.Unmarshal(bs, &j)
	if err != nil {
		err = fmt.Errorf("unmarshal package.json error: %w", err)
		return
	}

	_, ok = j.Scripts[script]
	return
}

type packageJson struct {
	Scripts map[string]string `json:"scripts"`
}

func (b *Hollow) runDevServer(ctx context.Context, name string, col color.Color, dir string) (err error) {
	logger := log.New(log.Options{
		IsDev:         false,
		To:            nil,
		DisableCaller: true,
		CallerSkip:    0,
		Name:          col.Render(fmt.Sprintf("[%v]", name)),
		DisableTime:   true,
		DisableLevel:  true,
	})
	err = execcmd.Run(ctx, dir, logger, "yarn", "dev")
	if err != nil {
		return
	}
	return nil
}

// Service 运行一个渲染程序
func (b *Hollow) Service(ctx context.Context, o ExecOption, addr string) error {
	s, err := httpsrv.NewService(addr)
	if err != nil {
		return err
	}
	if !config.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.GET("/_dev_/ws/:key", func(c *gin.Context) {
		key := c.Param("key")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Error(err)
			return
		}
		b.wsHub.Add(key, conn)
	})

	sub, _ := fs.Sub(hollowdev.Dist, "dist")
	r.StaticFS("/_dev_/static", http.FS(sub))
	//r.StaticFileFS("/_dev_/index.js", "dist/index.js", http.FS(editor.HollowDevFront))
	handle := b.ServiceHandle(o)
	r.NoRoute(func(c *gin.Context) {
		handle(c.Writer, c.Request)
	})

	s.Handler("/", r.Handler().ServeHTTP)

	return s.Start(ctx)
}

type MuitDir []http.FileSystem

func (m MuitDir) Open(name string) (http.File, error) {
	for _, i := range m {
		f, err := i.Open(name)
		if err != nil {
			continue
		}

		return f, nil
	}

	return nil, fs.ErrNotExist
}

type Content struct {
	Name       string                         `json:"name"`
	GetContent func(opt GetContentOpt) string `json:"getContent"`
	Meta       map[string]interface{}         `json:"meta"`
	Ext        string                         `json:"ext"`
	Content    string                         `json:"content"`
	IsDir      bool                           `json:"is_dir"`
	Toc        interface{}                    `json:"toc"`

	Assets Assets `json:"-"` // 文章中使用到的图片路径，base on content，需要复制到 statics
}

type ContentTree struct {
	Content
	Children ContentTrees `json:"children"`
}

type ContentTrees []ContentTree

func (cs ContentTrees) Sort(f func(a, b interface{}) bool) {
	sort.Slice(cs, func(i, j int) bool {
		return f(cs[i], cs[j])
	})

	for _, v := range cs {
		v.Children.Sort(f)
	}
}

func (cs ContentTrees) Filter(f func(a interface{}) bool) ContentTrees {
	if f == nil {
		return cs
	}
	var s ContentTrees
	for _, v := range cs {
		if f(v) {
			s = append(s, v)
		}
	}

	return s
}

func (cs ContentTrees) Flat(includeDir bool) ContentTrees {
	var s ContentTrees

	for _, v := range cs {
		if v.IsDir {
			if includeDir {
				s = append(s, v)
			}
		} else {
			s = append(s, v)
		}
		children := v.Children.Flat(includeDir)
		s = append(s, children...)
	}

	return s
}

type getBlogOption struct {
	Sort   func(a, b interface{}) bool `json:"sort"`
	Tree   bool                        `json:"tree"` // 传递 tree = true 则返回树结构
	Size   int                         `json:"size"`
	Page   int                         `json:"page"`
	Filter func(a interface{}) bool    `json:"filter"`
}

func (g getBlogOption) cacheKey() string {
	return ""
}

type BlogList struct {
	Total int           `json:"total"`
	List  []ContentTree `json:"list"`
}

func (b *Hollow) getContentLoader(ctx *RenderContext, ext string) (l ContentLoader, ok bool) {
	c, err := b.LoadConfig(ctx)
	if err != nil {
		// log.Warnf("LoadConfig for getContentLoader error: %v", err)
	}
	switch ext {
	case ".md", ".mdx":
		return NewMDLoader(c.Hollow.Assets, b.jsx, map[string]map[string]interface{}{
			"@bysir/hollow": {
				"getContents":      b.getContents(ctx),
				"getConfig":        b.getConfig(ctx),
				"getContentDetail": b.getContentDetail(ctx),
				"md":               b.md(ctx),
				"mdx":              b.mdx(ctx),
			},
		}), true
	}
	return nil, false
}

func MapDir(fsys fs.FS, root string, fn func(path string, d fs.DirEntry) (ContentTree, bool, error)) (ContentTrees, error) {
	at, err := mapDir(fsys, root, fn)
	if err != nil {
		return nil, err
	}
	return at, err
}

// walkDir recursively descends path, calling walkDirFn.
func mapDir(fsys fs.FS, name string, walkDirFn func(path string, d fs.DirEntry) (ContentTree, bool, error)) ([]ContentTree, error) {
	items, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, err
	}
	var ats []ContentTree
	for _, item := range items {
		nextName := path.Join(name, item.Name())

		a, ok, err := walkDirFn(nextName, item)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		var children ContentTrees
		if item.IsDir() {
			children, err = mapDir(fsys, nextName, walkDirFn)
			if err != nil {
				return nil, err
			}
		}

		ats = append(ats, ContentTree{
			Content:  a.Content,
			Children: children,
		})
	}
	return ats, nil
}

func trapBOM(fileBytes []byte) []byte {
	trimmedBytes := bytes.Trim(fileBytes, "\xef\xbb\xbf")
	return trimmedBytes
}

// tryReadMeta 尝试读取 meta
func (b *Hollow) tryReadMeta(file string) map[string]interface{} {
	// 读取 metadata
	body, err := fs.ReadFile(gobilly.NewStdFs(b.SourceFs), file)
	if err != nil {
		return nil
	}

	body = trapBOM(body)

	if bytes.HasPrefix(body, []byte("---\n")) {
		bbs := bytes.SplitN(body, []byte("---"), 3)
		if len(bbs) > 2 {
			metaByte := bbs[1]
			var meta = map[string]interface{}{}
			err = yaml.Unmarshal(metaByte, &meta)
			if err != nil {
				return nil
			}

			return meta
		}
	}

	return nil
}

//func MapDirGo()

// getContents 返回 dir 目录下的所有内容
func (b *Hollow) getContents(ctx *RenderContext) func(dir string, opt getBlogOption) BlogList {
	return func(dir string, opt getBlogOption) BlogList {
		//ctx = NewRenderContext()
		end := ctx.timerStart("getContents")
		defer end()
		var blogs ContentTrees
		//var total int
		cacheKey := fmt.Sprintf("getContents:%v%v", dir, opt)

		x, ok := ctx.cache.Get(cacheKey)
		if ok {
			blogs = x.(ContentTrees)
		} else {
			contents := sync.Map{}
			var wg sync.WaitGroup
			c := make(chan bool, 20)
			MapDir(gobilly.NewStdFs(b.SourceFs), dir, func(path string, d fs.DirEntry) (ContentTree, bool, error) {
				if !d.IsDir() {
					wg.Add(1)

					c <- true
					defer func() {
						<-c
						wg.Done()
					}()

					ext := filepath.Ext(path)
					loader, ok := b.getContentLoader(ctx, ext)
					if !ok {
						return ContentTree{}, true, nil
					}

					blog, err := loader.Load(path, false)

					if err != nil {
						err = fmt.Errorf("load blog '%v' error: %w", path, err)
						log.Warnf("%v", err)
						blog = b.newErrorContent(path, err)
					}

					if len(blog.Assets) > 0 {
						//log.Warnf("assets: %v", blog.Assets)
						ctx.Save("assets", blog.Assets)
					}

					contents.Store(path, blog)

					return ContentTree{}, true, nil

				}

				return ContentTree{}, true, nil
			})
			wg.Wait()

			ts, err := MapDir(gobilly.NewStdFs(b.SourceFs), dir, func(path string, d fs.DirEntry) (ContentTree, bool, error) {
				if d.IsDir() {
					// read dir meta
					var mate = map[string]interface{}{}
					metaFileName := filepath.Join(path, "meta.yaml")
					bs, err := fs.ReadFile(gobilly.NewStdFs(b.SourceFs), metaFileName)
					if err != nil {
						if !errors.Is(err, fs.ErrNotExist) {
							return ContentTree{}, false, fmt.Errorf("read meta file error: %w", err)
						}
						err = nil
					} else {
						err = yaml.Unmarshal(bs, &mate)
						if err != nil {
							return ContentTree{}, false, fmt.Errorf("unmarshal meta file error: %w", err)
						}
					}
					// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式
					for k, v := range mate {
						switch t := v.(type) {
						case time.Time:
							mate[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
						}
					}
					return ContentTree{Content: Content{
						Name: d.Name(),
						GetContent: func(opt GetContentOpt) string {
							return ""
						},
						Meta:    mate,
						Ext:     "",
						Content: "",
						IsDir:   true,
					}}, true, nil
				}

				i, ok := contents.Load(path)
				if ok {
					return ContentTree{Content: i.(Content)}, true, nil
				}

				return ContentTree{}, false, nil
			})
			if err != nil {
				log.Warnf("getContents error: %v", err)
				return BlogList{}
			}

			if !opt.Tree {
				ts = ts.Flat(false)
			}

			blogs = ts
			//ctx.cache.Add(cacheKey, blogs)
		}

		if opt.Sort != nil {
			blogs.Sort(opt.Sort)
		}

		blogs = blogs.Filter(opt.Filter)

		return BlogList{
			Total: 0,
			List:  blogs,
		}
	}
}

func (b *Hollow) newHtmlErrorMsg(err error) string {
	return fmt.Sprintf("<pre><code>%v</code></pre>", html.EscapeString(err.Error()))
}

func (b *Hollow) newErrorContent(file string, err error) Content {
	errHtml := b.newHtmlErrorMsg(err)
	meta := b.tryReadMeta(file)
	for k, v := range meta {
		switch t := v.(type) {
		case time.Time:
			// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式，前端才能处理
			meta[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
		}
	}

	return Content{
		Name: "Load error: " + file,
		GetContent: func(opt GetContentOpt) string {
			return errHtml
		},
		Content: errHtml,
		Ext:     filepath.Ext(file),
		Meta:    meta,
	}
}

// getContentDetail 返回一个内容
func (b *Hollow) getContentDetail(ctx *RenderContext) func(path string) Content {
	return func(path string) Content {
		ext := filepath.Ext(path)
		loader, ok := b.getContentLoader(ctx, ext)
		if !ok {
			err := fmt.Errorf("unsupported load '%v' file", ext)
			log.Warnf("%v", err)
			return b.newErrorContent(path, err)
		}
		blog, err := loader.Load(path, true)
		if err != nil {
			err = fmt.Errorf("load file '%v' error: %w", path, err)
			log.Warnf("%v", err)
			return b.newErrorContent(path, err)
		}

		return blog
	}
}

// getConfig config.yml 下的 theme_config 字段
func (b *Hollow) getConfig(ctx *RenderContext) func() interface{} {
	return func() interface{} {
		c, err := b.LoadConfig(ctx)
		if err != nil {
			log.Warnf("LoadConfig for js error: %v", err)
		}
		return c.Theme
	}
}

type MdOptions struct {
	Unwrap bool `json:"unwrap"`
}

func (b *Hollow) md(ctx *RenderContext) func(str string, options MdOptions) string {
	return func(str string, options MdOptions) string {
		ex, err := b.jsx.ExecCode([]byte(str), gojsx.WithFileName("root.md"), gojsx.WithAutoExecJsx(nil))
		if err != nil {
			return err.Error()
		}
		v := ex.Default.(gojsx.VDom)
		s := v.Render()
		// 支持处理只有一个 p 的情况，无法处理 <p> 1 </p> <h1> h1 </h1> <p> 2 </p>
		if options.Unwrap && strings.Count(s, "<p>") == 1 {
			s = strings.TrimPrefix(s, "<p>")
			s = strings.TrimSuffix(s, "</p>")
		}
		return s
	}
}

func (b *Hollow) mdx(ctx *RenderContext) func(str string, options MdOptions) string {
	return func(str string, options MdOptions) string {
		ex, err := b.jsx.ExecCode([]byte(str), gojsx.WithFileName("root.mdx"), gojsx.WithAutoExecJsx(nil))
		if err != nil {
			return err.Error()
		}
		v := ex.Default.(gojsx.VDom)
		s := v.Render()
		if options.Unwrap && strings.Count(s, "<p>") == 1 {
			s = strings.TrimPrefix(s, "<p>")
			s = strings.TrimSuffix(s, "</p>")
		}
		return s
	}

}

type LookupConfig struct {
	ThemePath  string // 主题绝对路径
	SourcePath string // 源文件绝对路径
}

// LookupConfig 返回配置信息
func (b *Hollow) LookupConfig(ctx *RenderContext) (LookupConfig, error) {
	sourcePath := b.SourceFs.Root()

	conf, err := b.LoadConfig(ctx)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
		} else {
			return LookupConfig{}, err
		}
	}

	themeUrl := b.prepareThemeUrl(conf.Hollow.Theme, b.FixedTheme)

	if strings.HasPrefix(themeUrl, "file://") {
		themeUrl = strings.TrimPrefix(themeUrl, "file://")
		if strings.HasPrefix(themeUrl, ".") {
			// 相对路径
		} else {
			// 绝对路径
			themeUrl = "/" + themeUrl
		}
	} else if strings.HasPrefix(themeUrl, "source://") {
		themeUrl = path.Join(sourcePath, strings.TrimPrefix(themeUrl, "source://"))
	} else {
		themeUrl = ""
	}
	switch {
	case strings.HasPrefix(themeUrl, "file://"):
		themeUrl = strings.TrimPrefix(themeUrl, "file://")
	}

	return LookupConfig{
		ThemePath:  themeUrl,
		SourcePath: sourcePath,
	}, nil
}

type Assets []string

func exportGojaValueToString(i interface{}) string {
	switch t := i.(type) {
	case goja.Value:
		return t.String()
	case string:
		return t
	}

	return fmt.Sprintf("%T can't to string", i)
}

// 和 goja 自己的 export 不一样的是，不会尝试导出单个变量为 golang 基础类型，而是保留 goja.Value，只是展开 Object
func exportGojaValue(i interface{}) interface{} {
	switch t := i.(type) {
	case *goja.Object:
		switch t.ExportType() {
		case reflect.TypeOf(map[string]interface{}{}):
			m := map[string]interface{}{}
			for _, k := range t.Keys() {
				m[k] = exportGojaValue(t.Get(k))
			}
			return m
		case reflect.TypeOf([]interface{}{}):
			arr := make([]interface{}, len(t.Keys()))
			for _, k := range t.Keys() {
				index, _ := strconv.ParseInt(k, 10, 64)
				arr[index] = exportGojaValue(t.Get(k))
			}
			return arr
		}
	}

	return i
}
