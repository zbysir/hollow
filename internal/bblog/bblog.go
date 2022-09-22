package bblog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	lru "github.com/hashicorp/golang-lru"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/easyfs"
	"github.com/zbysir/hollow/internal/pkg/git"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/httpsrv"
	"github.com/zbysir/hollow/internal/pkg/log"
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
		c, ok := goja.AssertFunction(t)
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

type Bblog struct {
	x *jsx.Jsx
	// 项目文件和主题文件可以独立，比如支持 project 在本地编写，主题通过 http 加载（到本地），用来做主题预览
	projectFs billy.Filesystem // 项目文件
	themeFs   billy.Filesystem // 多个主题，顶级是主题名字文件夹
	log       *zap.SugaredLogger
	cache     *lru.Cache // 缓存耗时操作与多次调用的数据，如获取 config、blog 文件夹，加速多个页面渲染相同数据的情况。
}

type Option struct {
	Fs      billy.Filesystem
	ThemeFs billy.Filesystem
}

type StdFileSystem struct {
}

func (f StdFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func NewBblog(o Option) (*Bblog, error) {
	if o.Fs == nil {
		//o.Fs = os.DirFS(".")
		o.Fs = osfs.New(".")
	}
	if o.ThemeFs == nil {
		o.ThemeFs = o.Fs
	}

	var err error
	x, err := jsx.NewJsx(jsx.Option{
		SourceCache: jsx.NewFileCache("./.cache"),
		SourceFs:    gobilly.NewStdFs(o.ThemeFs),
		Debug:       false,
	})
	if err != nil {
		return nil, err
	}

	cache, err := lru.New(100)
	if err != nil {
		return nil, err
	}
	b := &Bblog{
		x:         x,
		projectFs: o.Fs,
		log:       log.StdLogger,
		themeFs:   o.ThemeFs,
		cache:     cache,
	}

	x.RegisterModule("@bysir/hollow", map[string]interface{}{
		"getArticles":      b.getArticles,
		"getConfig":        b.getConfig,
		"getArticleDetail": b.getArticleDetail,
		"md":               b.md,
	})

	return b, nil
}

type ExecOption struct {
	Log *zap.SugaredLogger

	// 开发环境每次都会读取最新的文件，而生成环境会缓存
	IsDev bool
}

// Build 生成静态源文件
func (b *Bblog) Build(distPath string, o ExecOption) error {
	return b.BuildToFs(osfs.New(distPath), o)
}

func (b *Bblog) BuildToFs(dst billy.Filesystem, o ExecOption) error {
	start := time.Now()
	conf, err := b.LoadConfig(true)
	if err != nil {
		return err
	}

	configFile := filepath.Join(conf.Theme, "config")
	themeModule, err := b.loadTheme(configFile)
	if err != nil {
		return err
	}
	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	for i, p := range themeModule.Pages {
		var v, err = p.GetComponent()
		if err != nil {
			return err
		}
		body := v.Render()
		name := p.GetPath()
		distFile := "index.html"
		if name != "" && name != "index" {
			distFile = filepath.Join(name, "index.html")
		}
		f, err := dst.Create(distFile)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(body))
		if err != nil {
			return err
		}

		l.Infof("create pages [%d]: %v ", i, distFile)
	}

	for _, a := range themeModule.Assets {
		d := filepath.Dir(configFile)

		srcDir := filepath.Join(d, a)
		err = copyDir(srcDir, "", gobilly.NewStdFs(b.themeFs), dst)
		if err != nil {
			return err
		}
		l.Infof("copy theme assets: %v ", a)
	}

	for _, a := range conf.Assets {
		err = copyDir(a, "", gobilly.NewStdFs(b.projectFs), dst)
		if err != nil {
			return err
		}
		l.Infof("copy assets: %v ", a)
	}

	l.Infof("Done in %v", time.Now().Sub(start))
	return nil
}

func (b *Bblog) BuildAndPublish(dst billy.Filesystem, o ExecOption) error {
	err := b.BuildToFs(dst, o)
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
	g, err := git.NewGit(conf.Deploy.Token, dst, l)
	if err != nil {
		return err
	}

	branch := conf.Deploy.Branch
	if branch == "" {
		branch = "docs"
	}
	err = g.Push(conf.Deploy.Repo, branch, "by hollow", true)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bblog) PushProject(o ExecOption) error {
	conf, err := b.LoadConfig(true)
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	g, err := git.NewGit(conf.Source.Token, b.projectFs, l)
	if err != nil {
		return err
	}

	log.Infof("config %+v", conf.Source)
	err = g.Push(conf.Source.Repo, conf.Source.Branch, "-", true)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bblog) PullProject(o ExecOption) error {
	conf, err := b.LoadConfig(true)
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	g, err := git.NewGit(conf.Deploy.Token, b.projectFs, l)
	if err != nil {
		return err
	}

	err = g.Pull(conf.Source.Repo, conf.Source.Branch, true)
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

	log.Infof("try open %v", fullName)
	f, err := d.fs.Open(fullName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type Config struct {
	Theme  string  `json:"theme" yaml:"theme"`
	Deploy GitRepo `json:"deploy" yaml:"deploy"`
	Source GitRepo `json:"source" yaml:"source"`
	//Git         ConfigGit   `json:"git" yaml:"git"`
	Oss         ConfigOss   `json:"oss" yaml:"oss"`
	Assets      Assets      `json:"assets" yaml:"assets"`
	ThemeConfig interface{} `json:"theme_config" yaml:"theme_config"`
}

type GitRepo struct {
	Token  string `json:"token" yaml:"token"`
	Repo   string `json:"repo" yaml:"repo"`
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

func (b *Bblog) LoadConfig(env bool) (conf *Config, err error) {
	f, err := easyfs.GetFile(gobilly.NewStdFs(b.projectFs), "config.yml")
	if err != nil {
		return nil, err
	}

	body := f.Body
	if env {
		body = os.Expand(body, os.Getenv)
	} else {
		body = os.Expand(body, func(s string) string {
			k := os.Getenv(s)
			return strings.Repeat("*", len(k))
		})
	}

	conf = &Config{}
	err = yaml.Unmarshal([]byte(body), conf)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig error: %w", err)
	}

	return
}

func (b *Bblog) ServiceHandle(o ExecOption) func(writer http.ResponseWriter, request *http.Request) {
	var assetsHandler http.Handler

	var themeModule ThemeModule
	prepare := func() error {
		b.cache.Purge()
		var conf *Config

		conf, err := b.LoadConfig(true)
		if err != nil {
			return err
		}
		themeDir := conf.Theme

		configFile := filepath.Join(themeDir, "config")
		themeModule, err = b.loadTheme(configFile)
		if err != nil {
			return err
		}
		configDir := filepath.Dir(configFile)
		var dirs MuitDir
		for _, i := range themeModule.Assets {
			// 得到相对 themeFs root 的路径，e.g. dark/publish
			dir := filepath.Join(configDir, i)
			sub, err := fs.Sub(gobilly.NewStdFs(b.themeFs), dir)
			if err != nil {
				return fmt.Errorf("sub fs '%v' error: %w", i, err)
			}

			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}
		for _, i := range conf.Assets {
			dir := i
			sub, err := fs.Sub(gobilly.NewStdFs(b.projectFs), dir)
			if err != nil {
				return fmt.Errorf("sub fs '%v' error: %w", dir, err)
			}
			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}

		assetsHandler = http.FileServer(dirs)
		return nil
	}
	if !o.IsDev {
		err := prepare()
		if err != nil {
			return nil
		}
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		if o.IsDev {
			b.x.RefreshRegistry(nil)
			err := prepare()
			if err != nil {
				writer.Write([]byte(err.Error()))
				return
			}
		}

		reqPath := strings.Trim(request.URL.Path, "/")
		for _, p := range themeModule.Pages {
			if reqPath == p.GetPath() {
				component, err := p.GetComponent()
				if err != nil {
					writer.Write([]byte(err.Error()))
					return
				}
				x := component.Render()
				writer.Write([]byte(x))
				return
			}
		}
		assetsHandler.ServeHTTP(writer, request)
	}
}

// Service 运行一个渲染程序
func (b *Bblog) Service(ctx context.Context, o ExecOption, addr string) error {
	s, err := httpsrv.NewService(addr)
	if err != nil {
		return err
	}
	s.Handler("/", b.ServiceHandle(o))

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

var supportExt = map[string]bool{
	".md":   true,
	".html": true,
}

type Blog struct {
	Name       string                 `json:"name"`
	GetContent func() string          `json:"getContent"`
	Meta       map[string]interface{} `json:"meta"`
	Ext        string                 `json:"ext"`
	Content    string                 `json:"content"`
}

type BlogLoader interface {
	Load(fs fs.FS, filePath string, withContent bool) (Blog, bool, error)
}

type MDBlogLoader struct {
	assets Assets
}

func (m *MDBlogLoader) Load(f fs.FS, filePath string, withContent bool) (Blog, bool, error) {
	dir, name := filepath.Split(filePath)
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}

	ext := filepath.Ext(filePath)
	if supportExt[ext] {
		name = strings.TrimSuffix(name, ext)
	} else {
		return Blog{}, false, nil
	}

	// 读取 metadata
	body, err := fs.ReadFile(f, filePath)
	if err != nil {
		return Blog{}, false, err
	}

	var meta = map[string]interface{}{}
	if bytes.HasPrefix(body, []byte("---\n")) {
		bbs := bytes.SplitN(body, []byte("---"), 3)
		if len(bbs) > 2 {
			metaByte := bbs[1]
			err = yaml.Unmarshal(metaByte, &meta)
			if err != nil {
				return Blog{}, false, fmt.Errorf("parse file metadata error: %w", err)
			}

			body = bbs[2]
		}
	}

	// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式。在 js 中好处理
	for k, v := range meta {
		switch t := v.(type) {
		case time.Time:
			meta[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
		}
	}

	md := newMdRender(func(s string) string {
		p := s
		if filepath.IsAbs(s) {
		} else {
			p = filepath.Join(dir, s)
		}

		// 移除 assets 文件夹前缀
		for _, a := range m.assets {
			if strings.HasPrefix(p, "/"+a) {
				p = strings.TrimPrefix(p, "/"+a)
				break
			}
		}
		return p
	})

	content := ""
	if withContent {
		content = string(md.Render(body))
	}
	return Blog{
		Name: name,
		GetContent: func() string {
			if withContent {
				return content
			}
			return string(md.Render(body))
		},
		Meta:    meta,
		Ext:     ext,
		Content: content,
	}, true, nil
}

type getBlogOption struct {
	Sort func(a, b interface{}) bool `json:"sort"`
}

func (g getBlogOption) cacheKey() string {
	return ""
}

type BlogList struct {
	Total int    `json:"total"`
	List  []Blog `json:"list"`
}

func (b *Bblog) getLoader(ext string) (l BlogLoader, ok bool) {
	c, err := b.LoadConfig(true)
	if err != nil {
		return nil, false
	}
	switch ext {
	case ".md":
		return &MDBlogLoader{
			assets: c.Assets,
		}, true
	}
	return nil, false
}

// getArticles 返回 dir 目录下的所有博客
func (b *Bblog) getArticles(dir string, opt getBlogOption) BlogList {
	var blogs []Blog
	var total int
	cacheKey := fmt.Sprintf("blog:%v%v", dir, opt)
	x, ok := b.cache.Get(cacheKey)
	if ok {
		blogs = x.([]Blog)
	} else {
		err := fs.WalkDir(gobilly.NewStdFs(b.projectFs), dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			loader, ok := b.getLoader(ext)

			if !ok {
				return nil
			}
			total++

			blog, ok, err := loader.Load(gobilly.NewStdFs(b.projectFs), path, false)
			if err != nil {
				log.Errorf("load blog (%v) file: %v", path, err)
			}
			if !ok {
				return nil
			}
			// read meta
			metaFileName := path + ".yaml"
			bs, err := fs.ReadFile(gobilly.NewStdFs(b.projectFs), metaFileName)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return fmt.Errorf("read meta file error: %w", err)
				}
				err = nil
			} else {
				var m = map[string]interface{}{}
				err = yaml.Unmarshal(bs, &m)
				if err != nil {
					return fmt.Errorf("unmarshal meta file error: %w", err)
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

			blogs = append(blogs, blog)

			return nil
		})
		if err != nil {
			log.Errorf("get source '%s' error: %v", dir, err)
			return BlogList{}
		}

		b.cache.Add(cacheKey, blogs)
	}

	if opt.Sort != nil {
		sort.Slice(blogs, func(i, j int) bool {
			return opt.Sort(blogs[i], blogs[j])
		})
	}

	return BlogList{
		Total: 0,
		List:  blogs,
	}
}

// getArticleDetail 返回一个文件
func (b *Bblog) getArticleDetail(path string) Blog {
	ext := filepath.Ext(path)
	loader, ok := b.getLoader(ext)
	if !ok {
		return Blog{}
	}
	blog, ok, _ := loader.Load(gobilly.NewStdFs(b.projectFs), path, true)
	if !ok {
		return Blog{}
	}

	return blog
}

// getConfig config.yml 下的 theme_config 字段
func (b *Bblog) getConfig() interface{} {
	c, err := b.LoadConfig(true)
	if err != nil {
		panic(err)
	}
	return c.ThemeConfig
}

type MdOptions struct {
	Unwrap bool `json:"unwrap"`
}

// getConfig config.yml 下的 params 字段
func (b *Bblog) md(str string, options MdOptions) string {
	s := string(renderMd([]byte(str)))
	if options.Unwrap {
		s = strings.TrimPrefix(s, "<p>")
		s = strings.TrimSuffix(s, "</p>")
	}
	return s
}

type ThemeModule struct {
	raw    map[string]interface{}
	Pages  Pages
	Assets Assets
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

func (b *Bblog) loadTheme(configFile string) (ThemeModule, error) {
	envBs, _ := json.Marshal(nil)
	processCode := fmt.Sprintf("var process = {env: %s}", envBs)

	// 添加 ./ 告知 module 加载项目文件而不是 node_module
	configFile = "./" + filepath.Clean(configFile)
	v, err := b.x.RunJs("root.js", []byte(fmt.Sprintf(`%s;require("%v").default`, processCode, configFile)), false)
	if err != nil {
		return ThemeModule{}, err
	}

	// 直接 export 会导致 function 无法捕获 panic，不好实现
	raw := exportGojaValue(v).(map[string]interface{})

	pages := raw["pages"].([]interface{})
	ps := make(Pages, len(pages))
	for i, p := range pages {
		ps[i] = p.(map[string]interface{})
	}
	as := raw["assets"].([]interface{})
	assets := make(Assets, len(as))
	for k, v := range as {
		assets[k] = exportGojaValueToString(v)
	}

	return ThemeModule{raw: raw, Pages: ps, Assets: assets}, nil
}
