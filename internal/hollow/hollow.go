package hollow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/gookit/color"
	"github.com/gorilla/websocket"
	lru "github.com/hashicorp/golang-lru"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/front/hollow-dev"
	"github.com/zbysir/hollow/internal/pkg/asynctask"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/easyfs"
	"github.com/zbysir/hollow/internal/pkg/execcmd"
	"github.com/zbysir/hollow/internal/pkg/git"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/http_file_server"
	"github.com/zbysir/hollow/internal/pkg/httpsrv"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/ws"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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

func tryToVDom(i interface{}) jsx.VDom {
	switch t := i.(type) {
	case map[string]interface{}:
		return t
	}

	return jsx.VDom{}
}

func (p Page) GetComponent() (jsx.VDom, error) {
	var v jsx.VDom
	switch t := p["component"].(type) {
	case *goja.Object:
		c, ok := jsx.AssertFunction(t)
		if ok {
			// for: component: () => Index(props)
			val, err := c(nil)
			if err != nil {
				return v, err
			}
			v = tryToVDom(val.Export())
		} else {
			// for: component: Index(props)
			v = tryToVDom(t.Export())
		}
		return v, nil
	}

	return v, fmt.Errorf("uncased value type: %T", p["component"])
}

func (p Page) Render() (string, error) {
	if p["component"] != nil {
		vd, err := p.GetComponent()
		if err != nil {
			return "", err
		}
		return vd.Render(), nil
	} else if p["body"] != nil {
		return exportGojaValueToString(p["body"]), nil
	}

	return "", fmt.Errorf("can't render page: %+v", p)
}

type Hollow struct {
	jsx        *jsx.Jsx
	sourceFs   billy.Filesystem // 文件
	fixedTheme string           // 重新指定主题，可用于预览
	log        *zap.SugaredLogger
	cache      *lru.Cache // 缓存耗时操作与多次调用的数据，如获取 config、blog 文件夹，加速多个页面渲染相同数据的情况。
	asyncTask  *asynctask.Manager
	wsHub      *ws.WsHub
}

type Option struct {
	SourceFs   billy.Filesystem
	FixedTheme string // 重新指定主题，可用于预览
}

type StdFileSystem struct {
}

func (f StdFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

type MemJsxCache struct {
	m *sync.Map
}

func NewMemJsxCache() *MemJsxCache {
	return &MemJsxCache{m: &sync.Map{}}
}

func (m *MemJsxCache) Get(key string) (f *jsx.Source, exist bool, err error) {
	i, ok := m.m.Load(key)
	if !ok {
		return nil, false, nil
	}

	return i.(*jsx.Source), true, nil
}

func (m *MemJsxCache) Set(key string, f *jsx.Source) (err error) {
	m.m.Store(key, f)
	return nil
}

func NewHollow(o Option) (*Hollow, error) {
	if o.SourceFs == nil {
		o.SourceFs = osfs.New(".")
	}

	var err error
	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: NewMemJsxCache(),
		Debug:       false,
	})
	if err != nil {
		return nil, err
	}

	cache, err := lru.New(100)
	if err != nil {
		return nil, err
	}
	b := &Hollow{
		jsx:        jx,
		sourceFs:   o.SourceFs,
		fixedTheme: o.FixedTheme,
		log:        log.Logger(),
		cache:      cache,
		asyncTask:  asynctask.NewManager(),
		wsHub:      ws.NewHub(),
	}
	b.asyncTask.AddListener(func(task *asynctask.Task, event *asynctask.Event) {
		if event.IsDone {
			b.wsHub.Close(task.Key)
		} else {
			b.wsHub.Send(task.Key, []byte(event.Log))
		}
	})

	jx.RegisterModule("@bysir/hollow", map[string]interface{}{
		"getContents":      b.getContents,
		"getConfig":        b.getConfig,
		"getContentDetail": b.getContentDetail,
		"md":               b.md,
	})

	return b, nil
}

type ExecOption struct {
	Log *zap.SugaredLogger

	IsDev bool // 开发环境每次都会读取最新的文件，而生成环境会缓存
}

// Build 生成静态源文件
func (b *Hollow) Build(ctx context.Context, distPath string, o ExecOption) error {
	return b.BuildToFs(ctx, osfs.New(distPath), o)
}

func (b *Hollow) BuildToFs(ctx context.Context, dst billy.Filesystem, o ExecOption) error {
	start := time.Now()
	conf, err := b.LoadConfig(true)
	if err != nil {
		return err
	}
	themeUrl := b.prepareThemeUrl(conf.Theme, b.fixedTheme)
	themeLoader, err := b.GetThemeLoader(themeUrl)
	if err != nil {
		return err
	}
	var themeFs fs.FS
	themeModule, themeFs, _, err := themeLoader.Load(ctx, b.jsx, true, false)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	l = l.Named("[Build]\t")

	for i, p := range themeModule.Pages {
		body, err := p.Render()
		if err != nil {
			return err
		}
		name := p.GetPath()
		var distFile string

		// 如果有扩展名，则说明是文件
		if filepath.Ext(name) != "" {
			distFile = name
		} else {
			// 存入文件夹
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

	for _, a := range conf.Assets {
		err = copyDir(a, "", gobilly.NewStdFs(b.sourceFs), dst)
		if err != nil {
			return err
		}
		l.Infof("Copy assets: %v ", a)
	}

	l.Infof("Done in %v", time.Now().Sub(start))
	return nil
}

func (b *Hollow) BuildAndPublish(ctx context.Context, dst billy.Filesystem, o ExecOption) error {
	err := b.BuildToFs(ctx, dst, o)
	if err != nil {
		return err
	}

	conf, err := b.LoadConfig(true)
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	l = l.Named("[Git]\t")

	g, err := git.NewGit(conf.Deploy.Token, dst, l)
	if err != nil {
		return err
	}

	branch := conf.Deploy.Branch
	if branch == "" {
		branch = "docs"
	}
	err = g.Push(conf.Deploy.Remote, branch, "by hollow", true)
	if err != nil {
		return err
	}
	return nil
}

func (b *Hollow) PushProject(o ExecOption) error {
	conf, err := b.LoadConfig(true)
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	g, err := git.NewGit(conf.Source.Token, b.sourceFs, l)
	if err != nil {
		return err
	}

	log.Infof("config %+v", conf.Source)
	err = g.Push(conf.Source.Remote, conf.Source.Branch, "-", true)
	if err != nil {
		return err
	}

	return nil
}

func (b *Hollow) PullProject(o ExecOption) error {
	conf, err := b.LoadConfig(true)
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	g, err := git.NewGit(conf.Deploy.Token, b.sourceFs, l)
	if err != nil {
		return err
	}

	err = g.Pull(conf.Source.Remote, conf.Source.Branch, true)
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

type Config struct {
	Theme       string      `json:"theme" yaml:"theme"`
	Deploy      GitRepo     `json:"deploy" yaml:"deploy"`
	Source      GitRepo     `json:"source" yaml:"source"`
	Oss         ConfigOss   `json:"oss" yaml:"oss"`
	Assets      Assets      `json:"assets" yaml:"assets"`
	ThemeConfig interface{} `json:"theme_config" yaml:"theme_config"`
}

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

// LoadConfig 加载 source 下的 config 文件
func (b *Hollow) LoadConfig(expandEnv bool) (conf Config, err error) {
	f, err := easyfs.GetFile(gobilly.NewStdFs(b.sourceFs), "config.yml")
	if err != nil {
		return
	}

	body := f.Body
	if expandEnv {
		body = os.Expand(body, os.Getenv)
	} else {
		body = os.Expand(body, func(s string) string {
			k := os.Getenv(s)
			return strings.Repeat("*", len(k))
		})
	}

	err = yaml.Unmarshal([]byte(body), &conf)
	if err != nil {
		err = fmt.Errorf("LoadConfig error: %w", err)
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

// GetThemeLoader 返回主题加载器，支持以下协议的地址。
// https:// : git
// file:// : relative or absolute path, e.g. file://usr/bysir/xx , file://./bysir/xx
// source:// : relative path of source fs
func (b *Hollow) GetThemeLoader(url string) (ThemeLoader, error) {
	switch {
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"):
		return NewGitThemeLoader(b.asyncTask, url), nil
	case strings.HasPrefix(url, "file://"):
		pa := strings.TrimPrefix(url, "file://")
		if strings.HasPrefix(pa, ".") {
			// 相对路径
		} else {
			// 绝对路径
			pa = "/" + pa
		}
		return NewFsThemeLoader(gobilly.NewStdFs(osfs.New(pa))), nil
	case strings.HasPrefix(url, "source://"):
		pa := strings.TrimPrefix(url, "source://")
		subFs, err := b.sourceFs.Chroot(pa)
		if err != nil {
			return nil, err
		}
		return NewFsThemeLoader(gobilly.NewStdFs(subFs)), nil
	default:
		return nil, fmt.Errorf("unsupported protocol, url: %v", url)
	}
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

	prepare := func(opt *PrepareOpt) (asyncTaskKey string, err error) {
		b.cache.Purge()
		var projectConf Config

		projectConf, err = b.LoadConfig(true)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {

			} else {
				return "", err
			}
		}

		refresh := false
		if opt != nil && opt.NoCache {
			refresh = true
		}

		themeUrl := b.prepareThemeUrl(projectConf.Theme, b.fixedTheme)
		themeLoader, err := b.GetThemeLoader(themeUrl)
		if err != nil {
			return "", err
		}
		var themeFs fs.FS
		var task *asynctask.Task
		themeModule, themeFs, task, err = themeLoader.Load(context.Background(), b.jsx, refresh, true)
		if err != nil {
			return "", err
		}
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
		for _, dir := range projectConf.Assets {
			sub, err := fs.Sub(gobilly.NewStdFs(b.sourceFs), dir)
			if err != nil {
				return "", fmt.Errorf("sub fs '%v' error: %w", dir, err)
			}
			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}

		assetsHandler = http_file_server.FileServer(dirs)
		return "", nil
	}

	// 不是 dev 环境只会加载一次主题，而不是每次刷新页面都加载
	if !o.IsDev {
		task, err := prepare(nil)
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

		if o.IsDev {
			opt := PrepareOpt{
				NoCache: request.Header.Get("Cache-Control") == "no-cache",
			}
			task, err := prepare(&opt)
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
	config, err := b.LookupConfig()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	if config.ThemePath != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, err := b.checkPackageJsonScript(config.ThemePath, "dev")
			if err != nil {
				panic(err)
			}
			if !ok {
				return
			}

			name := "Theme "
			col := color.Green
			log.Infof(col.Render(fmt.Sprintf("Running dev service for [%s] [%s]", name, config.ThemePath)))
			err = b.runDevServer(ctx, name, col, config.ThemePath)
			if err != nil {
				log.Errorf(col.Render(fmt.Sprintf("RunDevServer [%v] error: %v", name, err)))
			}
		}()
	}

	if config.SourcePath != "" && config.ThemePath != config.SourcePath {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, err := b.checkPackageJsonScript(config.SourcePath, "dev")
			if err != nil {
				panic(err)
			}
			if !ok {
				return
			}

			name := "Source"
			col := color.Magenta
			log.Infof(col.Render(fmt.Sprintf("Running dev service for [%v] [%v]", name, config.ThemePath)))

			err = b.runDevServer(ctx, name, col, config.SourcePath)
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
	r.NoRoute(func(c *gin.Context) {
		b.ServiceHandle(o)(c.Writer, c.Request)
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
		children := v.Children.Flat(includeDir)
		if len(children) == 0 || includeDir {
			s = append(s, v)
		}
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

func (b *Hollow) getContentLoader(ext string) (l ContentLoader, ok bool) {
	c, err := b.LoadConfig(true)
	if err != nil {
		log.Warnf("LoadConfig for getContentLoader error: %v", err)
	}
	switch ext {
	case ".md", ".mdx":
		return NewMDLoader(c.Assets, b.jsx), true
	}
	return nil, false
}

func MapDir(fsys fs.FS, root string, fn func(path string, d fs.DirEntry) (ContentTree, error)) (ContentTrees, error) {
	info, err := fs.Stat(fsys, root)
	if err != nil {

	} else {
		at, err := mapDir(fsys, root, &statDirEntry{info}, fn)
		return at, err
	}

	return nil, err
}

type statDirEntry struct {
	info fs.FileInfo
}

func (d *statDirEntry) Name() string               { return d.info.Name() }
func (d *statDirEntry) IsDir() bool                { return d.info.IsDir() }
func (d *statDirEntry) Type() fs.FileMode          { return d.info.Mode().Type() }
func (d *statDirEntry) Info() (fs.FileInfo, error) { return d.info, nil }

// walkDir recursively descends path, calling walkDirFn.
func mapDir(fsys fs.FS, name string, d fs.DirEntry, walkDirFn func(path string, d fs.DirEntry) (ContentTree, error)) ([]ContentTree, error) {
	dirs, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, err
	}
	var ats []ContentTree
	for _, d1 := range dirs {
		name1 := path.Join(name, d1.Name())

		a, err := walkDirFn(name1, d1)
		if err != nil {
			return nil, err
		}
		if a.Name == "" {
			continue
		}
		var children ContentTrees
		if d1.IsDir() {
			children, err = mapDir(fsys, name1, d1, walkDirFn)
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

// getContents 返回 dir 目录下的所有内容
func (b *Hollow) getContents(dir string, opt getBlogOption) BlogList {
	var blogs ContentTrees
	//var total int
	cacheKey := fmt.Sprintf("blog:%v%v", dir, opt)
	x, ok := b.cache.Get(cacheKey)
	if ok {
		blogs = x.(ContentTrees)
	} else {
		ts, err := MapDir(gobilly.NewStdFs(b.sourceFs), dir, func(path string, d fs.DirEntry) (ContentTree, error) {
			//if err != nil {
			//	return ContentTree{}, err
			//}
			if d.IsDir() {
				// read dir meta
				var mate = map[string]interface{}{}
				metaFileName := filepath.Join(path, "meta.yaml")
				bs, err := fs.ReadFile(gobilly.NewStdFs(b.sourceFs), metaFileName)
				if err != nil {
					if !errors.Is(err, fs.ErrNotExist) {
						return ContentTree{}, fmt.Errorf("read meta file error: %w", err)
					}
					err = nil
				} else {
					err = yaml.Unmarshal(bs, &mate)
					if err != nil {
						return ContentTree{}, fmt.Errorf("unmarshal meta file error: %w", err)
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
				}}, nil
			}

			ext := filepath.Ext(path)
			loader, ok := b.getContentLoader(ext)

			if !ok {
				return ContentTree{}, nil
			}

			blog, err := loader.Load(gobilly.NewStdFs(b.sourceFs), path, false)
			if err != nil {
				log.Errorf("load blog (%v) file: %v", path, err)
				return ContentTree{}, nil
			}
			// read meta
			metaFileName := path + ".yaml"
			bs, err := fs.ReadFile(gobilly.NewStdFs(b.sourceFs), metaFileName)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return ContentTree{}, fmt.Errorf("read meta file error: %w", err)
				}
				err = nil
			} else {
				var m = map[string]interface{}{}
				err = yaml.Unmarshal(bs, &m)
				if err != nil {
					return ContentTree{}, fmt.Errorf("unmarshal meta file error: %w", err)
				}

				for k, v := range m {
					blog.Meta[k] = v
				}
			}
			// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式
			for k, v := range blog.Meta {
				switch t := v.(type) {
				case time.Time:
					blog.Meta[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
				}
			}

			return ContentTree{Content: blog}, nil
		})
		if err != nil {
			log.Warnf("getContents error: %v", err)
			return BlogList{}
		}

		if !opt.Tree {
			ts = ts.Flat(false)
		}

		blogs = ts
		b.cache.Add(cacheKey, blogs)
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

// getContentDetail 返回一个内容
func (b *Hollow) getContentDetail(path string) Content {
	ext := filepath.Ext(path)
	loader, ok := b.getContentLoader(ext)
	if !ok {
		log.Warnf("can't loader '%v' file", ext)
		return Content{}
	}
	blog, err := loader.Load(gobilly.NewStdFs(b.sourceFs), path, true)
	if err != nil {
		log.Warnf("can't loader file, error: %v", err)
		return Content{}
	}
	if !ok {
		return Content{}
	}

	return blog
}

// getConfig config.yml 下的 theme_config 字段
func (b *Hollow) getConfig() interface{} {
	c, err := b.LoadConfig(true)
	if err != nil {
		log.Warnf("LoadConfig for js error: %v", err)
	}
	return c.ThemeConfig
}

type MdOptions struct {
	Unwrap bool `json:"unwrap"`
}

// getConfig config.yml 下的 params 字段
func (b *Hollow) md(str string, options MdOptions) string {
	md := NewGoldMdRender(GoldMdRenderOptions{
		jsx:              b.jsx,
		fs:               gobilly.NewStdFs(b.sourceFs),
		assetsUrlProcess: nil,
	})
	r, err := md.Render([]byte(str))
	if err != nil {
		return err.Error()
	}
	s := string(r.Body)
	if options.Unwrap {
		s = strings.TrimPrefix(s, "<p>")
		s = strings.TrimSuffix(s, "</p>")
	}
	return s
}

type LookupConfig struct {
	ThemePath  string // 主题绝对路径
	SourcePath string // 源文件绝对路径
}

// LookupConfig 返回配置信息
func (b *Hollow) LookupConfig() (LookupConfig, error) {
	sourcePath := b.sourceFs.Root()

	conf, err := b.LoadConfig(true)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
		} else {
			return LookupConfig{}, err
		}
	}

	themeUrl := b.prepareThemeUrl(conf.Theme, b.fixedTheme)

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
