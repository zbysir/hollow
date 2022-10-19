package bblog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	lru "github.com/hashicorp/golang-lru"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/easyfs"
	"github.com/zbysir/hollow/internal/pkg/git"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/http_file_server"
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

type Bblog struct {
	x         *jsx.Jsx
	projectFs billy.Filesystem // 文件
	log       *zap.SugaredLogger
	cache     *lru.Cache // 缓存耗时操作与多次调用的数据，如获取 config、blog 文件夹，加速多个页面渲染相同数据的情况。
}

type Option struct {
	Fs billy.Filesystem
}

type StdFileSystem struct {
}

func (f StdFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func NewBblog(o Option) (*Bblog, error) {
	if o.Fs == nil {
		o.Fs = osfs.New(".")
	}

	var err error
	x, err := jsx.NewJsx(jsx.Option{
		SourceCache: jsx.NewFileCache("./.cache"),
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
		cache:     cache,
	}

	x.RegisterModule("react", map[string]interface{}{
		"useState":  func() []interface{} { return []interface{}{nil, nil} },
		"useEffect": func() {},
		"useRef":    func() {},
	})

	x.RegisterModule("fuse.js", map[string]interface{}{})

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

	themeLoader, err := b.GetThemeLoader(conf.Theme)
	if err != nil {
		return err
	}
	var themeFs fs.FS
	themeModule, themeFs, err := themeLoader.Load(b.x, conf.Theme, true)
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
		err = copyDir(a, "", gobilly.NewStdFs(b.projectFs), dst)
		if err != nil {
			return err
		}
		l.Infof("Copy assets: %v ", a)
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
	err = g.Push(conf.Source.Remote, conf.Source.Branch, "-", true)
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

	log.Infof("try open %v", fullName)
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

type PrepareOpt struct {
	NoCache bool
}

func (b *Bblog) GetThemeLoader(path string) (ThemeLoader, error) {
	switch {
	case strings.HasPrefix(path, "http://"), strings.HasPrefix(path, "https://"):
		return NewGitThemeLoader(), nil
	case strings.HasPrefix(path, "file://"):
		abs, err := resolveFileUrl(path)
		if err != nil {
			return nil, err
		}

		return NewLocalThemeLoader(gobilly.NewStdFs(osfs.New(abs)), "."), nil
	default:
		return NewLocalThemeLoader(gobilly.NewStdFs(b.projectFs), ""), nil
	}
}

func (b *Bblog) ServiceHandle(o ExecOption) func(writer http.ResponseWriter, request *http.Request) {
	var assetsHandler http.Handler

	var themeModule ThemeExport

	prepare := func(opt *PrepareOpt) error {
		b.cache.Purge()
		var projectConf *Config

		projectConf, err := b.LoadConfig(true)
		if err != nil {
			return err
		}
		themeDir := projectConf.Theme

		themeLoader, err := b.GetThemeLoader(themeDir)
		if err != nil {
			return err
		}
		refresh := false
		if opt != nil && opt.NoCache {
			refresh = true
		}
		var themeFs fs.FS
		//log.Infof("refresh: %+v", refresh)
		themeModule, themeFs, err = themeLoader.Load(b.x, themeDir, refresh)
		if err != nil {
			return err
		}

		var dirs MuitDir
		for _, dir := range themeModule.Assets {
			sub, err := fs.Sub(themeFs, dir)
			if err != nil {
				return fmt.Errorf("sub fs '%v' error: %w", dir, err)
			}

			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}
		for _, dir := range projectConf.Assets {
			sub, err := fs.Sub(gobilly.NewStdFs(b.projectFs), dir)
			if err != nil {
				return fmt.Errorf("sub fs '%v' error: %w", dir, err)
			}
			dirs = append(dirs, &DirFs{
				fs: http.FS(sub),
			})
		}

		assetsHandler = http_file_server.FileServer(dirs)
		return nil
	}
	if !o.IsDev {
		err := prepare(nil)
		if err != nil {
			return nil
		}
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		reqPath := strings.Trim(request.URL.Path, "/")

		if o.IsDev {
			// TODO 优化主题刷新逻辑，不应该每次请求都刷新，需要做异步刷新
			opt := PrepareOpt{}
			opt.NoCache = request.Header.Get("Cache-Control") == "no-cache"
			err := prepare(&opt)
			if err != nil {
				writer.Write([]byte(err.Error()))
				return
			}
		}
		for _, p := range themeModule.Pages {
			if reqPath == p.GetPath() {
				body, err := p.Render()
				if err != nil {
					writer.WriteHeader(400)
					writer.Write([]byte(err.Error()))
					return
				}
				writer.Write([]byte(body))
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
	Name       string                         `json:"name"`
	GetContent func(opt GetContentOpt) string `json:"getContent"`
	Meta       map[string]interface{}         `json:"meta"`
	Ext        string                         `json:"ext"`
	Content    string                         `json:"content"`
	IsDir      bool                           `json:"is_dir"`
}

type ArticleTree struct {
	Blog
	Children ArticleTrees `json:"children"`
}

type ArticleTrees []ArticleTree

func (ats ArticleTrees) Sort(f func(a, b interface{}) bool) {
	sort.Slice(ats, func(i, j int) bool {
		return f(ats[i], ats[j])
	})

	for _, v := range ats {
		v.Children.Sort(f)
	}
}

func (ats ArticleTrees) Flat(includeDir bool) ArticleTrees {
	var s ArticleTrees

	for _, v := range ats {
		children := v.Children.Flat(includeDir)
		if len(children) == 0 || includeDir {
			s = append(s, v)
		}
		s = append(s, children...)
	}

	return s
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

	md := newMdRender(func(p string) string {
		if filepath.IsAbs(p) {
		} else {
			p = filepath.Join(dir, p)
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
		GetContent: func(opt GetContentOpt) string {
			var s string
			if withContent {
				s = content
			} else {
				s = string(md.Render(body))
			}
			if opt.Pure {
				d, err := goquery.NewDocumentFromReader(strings.NewReader(s))
				if err != nil {
					return err.Error()
				}

				s = d.Text()
			}
			return s
		},
		Meta:    meta,
		Ext:     ext,
		Content: content,
	}, true, nil
}

type GetContentOpt struct {
	Pure bool `json:"pure"` // 返回纯文本，一般用于做搜索
}

type HtmlLoader struct {
	//assets Assets
}

func (m *HtmlLoader) Load(f fs.FS, filePath string, withContent bool) (Blog, bool, error) {
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

	content := ""
	if withContent {
		content = string(body)
	}
	return Blog{
		Name: name,
		GetContent: func(opt GetContentOpt) string {
			var s = string(body)
			if opt.Pure {
				d, err := goquery.NewDocumentFromReader(strings.NewReader(s))
				if err != nil {
					return err.Error()
				}

				s = d.Text()
			}

			return s
		},
		Meta:    meta,
		Ext:     ext,
		Content: content,
	}, true, nil
}

type getBlogOption struct {
	Sort func(a, b interface{}) bool `json:"sort"`
	Tree bool                        `json:"tree"` // 传递 tree = true 则返回树结构
	Size int                         `json:"size"`
	Page int                         `json:"page"`
}

func (g getBlogOption) cacheKey() string {
	return ""
}

type BlogList struct {
	Total int           `json:"total"`
	List  []ArticleTree `json:"list"`
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
	case ".html":
		return &HtmlLoader{}, true
	}
	return nil, false
}

func MapDir(fsys fs.FS, root string, fn func(path string, d fs.DirEntry) (ArticleTree, error)) (ArticleTrees, error) {
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
func mapDir(fsys fs.FS, name string, d fs.DirEntry, walkDirFn func(path string, d fs.DirEntry) (ArticleTree, error)) ([]ArticleTree, error) {
	dirs, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, err
	}
	var ats []ArticleTree
	for _, d1 := range dirs {
		name1 := path.Join(name, d1.Name())

		a, err := walkDirFn(name1, d1)
		if err != nil {
			return nil, err
		}
		if a.Name == "" {
			continue
		}
		var children ArticleTrees
		if d1.IsDir() {
			children, err = mapDir(fsys, name1, d1, walkDirFn)
			if err != nil {
				return nil, err
			}
		}

		ats = append(ats, ArticleTree{
			Blog:     a.Blog,
			Children: children,
		})
	}
	return ats, nil
}

// getArticles 返回 dir 目录下的所有博客
func (b *Bblog) getArticles(dir string, opt getBlogOption) BlogList {
	var blogs ArticleTrees
	//var total int
	cacheKey := fmt.Sprintf("blog:%v%v", dir, opt)
	x, ok := b.cache.Get(cacheKey)
	if ok {
		blogs = x.(ArticleTrees)
	} else {
		ts, err := MapDir(gobilly.NewStdFs(b.projectFs), dir, func(path string, d fs.DirEntry) (ArticleTree, error) {
			//if err != nil {
			//	return ArticleTree{}, err
			//}
			if d.IsDir() {
				// read dir meta
				var mate = map[string]interface{}{}
				metaFileName := filepath.Join(path, "meta.yaml")
				bs, err := fs.ReadFile(gobilly.NewStdFs(b.projectFs), metaFileName)
				if err != nil {
					if !errors.Is(err, fs.ErrNotExist) {
						return ArticleTree{}, fmt.Errorf("read meta file error: %w", err)
					}
					err = nil
				} else {
					err = yaml.Unmarshal(bs, &mate)
					if err != nil {
						return ArticleTree{}, fmt.Errorf("unmarshal meta file error: %w", err)
					}
				}
				// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式
				for k, v := range mate {
					switch t := v.(type) {
					case time.Time:
						mate[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
					}
				}
				return ArticleTree{Blog: Blog{
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
			loader, ok := b.getLoader(ext)

			if !ok {
				return ArticleTree{}, nil
			}

			blog, ok, err := loader.Load(gobilly.NewStdFs(b.projectFs), path, false)
			if err != nil {
				log.Errorf("load blog (%v) file: %v", path, err)
			}
			if !ok {
				return ArticleTree{}, nil
			}
			// read meta
			metaFileName := path + ".yaml"
			bs, err := fs.ReadFile(gobilly.NewStdFs(b.projectFs), metaFileName)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return ArticleTree{}, fmt.Errorf("read meta file error: %w", err)
				}
				err = nil
			} else {
				var m = map[string]interface{}{}
				err = yaml.Unmarshal(bs, &m)
				if err != nil {
					return ArticleTree{}, fmt.Errorf("unmarshal meta file error: %w", err)
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

			return ArticleTree{Blog: blog}, nil
		})
		if err != nil {
			log.Warnf("MapDir error: %v", err)
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

	return ThemeModule{}, nil
}
