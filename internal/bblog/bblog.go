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
	"github.com/zbysir/blog/internal/pkg/easyfs"
	"github.com/zbysir/blog/internal/pkg/git"
	"github.com/zbysir/blog/internal/pkg/log"
	jsx "github.com/zbysir/gojsx"
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
	// 项目文件和主题文件可以独立，比如支持 project 在本地编写，主题通过 http 加载，用来做主题预览
	projectFs fs.FS // 项目文件
	themeFs   fs.FS // 多个主题，顶级是主题名字文件夹
	log       *zap.SugaredLogger
}

type Option struct {
	Fs      fs.FS
	ThemeFs fs.FS
}

type StdFileSystem struct {
}

func (f StdFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func NewBblog(o Option) (*Bblog, error) {
	if o.Fs == nil {
		//o.Fs = os.DirFS(".")
		o.Fs = StdFileSystem{}
	}
	if o.ThemeFs == nil {
		o.ThemeFs = o.Fs
	}

	var err error
	x, err := jsx.NewJsx(jsx.Option{
		SourceCache: jsx.NewFileCache("./.cache"),
		SourceFs:    o.ThemeFs,
		Debug:       true,
	})
	if err != nil {
		panic(err)
	}

	b := &Bblog{
		x:         x,
		projectFs: o.Fs,
		log:       log.StdLogger,
		themeFs:   o.ThemeFs,
	}

	x.RegisterModule("@bysir/hollow", map[string]interface{}{
		"getBlogs":      b.getBlogs,
		"getConfig":     b.getConfig,
		"getBlogDetail": b.getBlogDetail,
		"md":            b.md,
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
	conf, err := b.loadConfig()
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

	for _, p := range themeModule.Pages {
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

		l.Infof("create pages: %v ", distFile)
	}

	for _, a := range themeModule.Assets {
		d := filepath.Dir(configFile)

		srcDir := filepath.Join(d, a)
		err = copyDir(srcDir, "", b.themeFs, dst)
		if err != nil {
			return err
		}
		l.Infof("copy theme assets: %v ", a)
	}

	for _, a := range conf.Assets {
		err = copyDir(a, "", b.projectFs, dst)
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

	conf, err := b.loadConfig()
	if err != nil {
		return err
	}

	l := b.log
	if o.Log != nil {
		l = o.Log
	}
	g := git.NewGit(conf.Git.Token, l)

	branch := conf.Git.Branch
	if branch == "" {
		branch = "docs"
	}
	err = g.Push(dst, conf.Git.Repo, "bblog", branch, true)
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
	Theme       string      `json:"theme" yaml:"theme"`
	Git         ConfigGit   `json:"git" yaml:"git"`
	Assets      Assets      `json:"assets" yaml:"assets"`
	ThemeConfig interface{} `json:"theme_config" yaml:"theme_config"`
}

type ConfigGit struct {
	Repo   string `json:"repo" yaml:"repo"`
	Token  string `json:"token" yaml:"token"`
	Branch string `yaml:"branch"`
}

func (b *Bblog) loadConfig() (conf *Config, err error) {
	f, err := easyfs.GetFile(b.projectFs, "config.yml")
	if err != nil {
		return nil, err
	}

	body := f.Body
	body = os.ExpandEnv(body)

	conf = &Config{}
	err = yaml.Unmarshal([]byte(body), conf)
	if err != nil {
		return nil, fmt.Errorf("loadConfig error: %w", err)
	}

	return
}

// Service 运行一个渲染程序
func (b *Bblog) Service(ctx context.Context, o ExecOption, addr string) error {
	s, err := NewService(addr)
	if err != nil {
		return err
	}
	var assetsHandler http.Handler

	var themeModule ThemeModule
	prepare := func() error {
		var conf *Config

		conf, err = b.loadConfig()
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
			sub, err := fs.Sub(b.themeFs, dir)
			if err != nil {
				return fmt.Errorf("sub fs '%v' error: %w", i, err)
			}

			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}
		for _, i := range conf.Assets {
			dir := i
			sub, err := fs.Sub(b.projectFs, dir)
			if err != nil {
				return fmt.Errorf("sub fs '%v' error: %w", dir, err)
			}
			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}
		//
		//dirs = append(dirs, &DirFs{
		//	appPrefix: filepath.Join(configDir),
		//	fs:        http.FS(b.themeFs),
		//})
		//dirs = append(dirs, &DirFs{
		//	appPrefix: "",
		//	stripPrefix: "",
		//	fs:        http.FS(b.projectFs),
		//})

		// for img import by md file
		//dirs = append(dirs, &DirFs{
		//	appPrefix: "",
		//	fs:     http.FS(b.projectFs),
		//})
		assetsHandler = http.FileServer(dirs)
		return nil
	}
	if !o.IsDev {
		err = prepare()
		if err != nil {
			return err
		}
	}

	s.Handler("/", func(writer http.ResponseWriter, request *http.Request) {
		if o.IsDev {
			b.x.RefreshRegistry(nil)
			err = prepare()
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

	})

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
		p := filepath.Join(dir, s)

		// 移除 assets 文件夹前缀
		for _, a := range m.assets {
			if strings.HasPrefix(p, a) {
				p = strings.TrimPrefix(p, a)
				break
			}
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
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
type BlogList struct {
	Total int    `json:"total"`
	List  []Blog `json:"list"`
}

func (b *Bblog) getLoader(ext string) (l BlogLoader, ok bool) {
	c, err := b.loadConfig()
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

// getBlogs 返回 dir 目录下的所有博客
func (b *Bblog) getBlogs(dir string, opt getBlogOption) BlogList {
	var blogs []Blog
	var total int

	err := fs.WalkDir(b.projectFs, dir, func(path string, d fs.DirEntry, err error) error {
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

		blog, ok, err := loader.Load(b.projectFs, path, false)
		if err != nil {
			log.Errorf("load blog (%v) file: %v", path, err)
		}
		if !ok {
			return nil
		}
		// read meta
		metaFileName := path + ".yaml"
		bs, err := fs.ReadFile(b.projectFs, metaFileName)
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

// getBlogDetail 返回一个文件
func (b *Bblog) getBlogDetail(path string) Blog {
	ext := filepath.Ext(path)
	loader, ok := b.getLoader(ext)
	if !ok {
		return Blog{}
	}
	blog, ok, _ := loader.Load(b.projectFs, path, true)
	if !ok {
		return Blog{}
	}

	return blog
}

// getConfig config.yml 下的 theme_config 字段
func (b *Bblog) getConfig() interface{} {
	c, err := b.loadConfig()
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
