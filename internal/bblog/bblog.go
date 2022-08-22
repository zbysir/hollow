package bblog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/russross/blackfriday/v2"
	"github.com/zbysir/blog/internal/pkg/log"
	jsx "github.com/zbysir/gojsx"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Page map[string]interface{}
type Pages []Page

func (p Page) GetName() string {
	switch t := p["name"].(type) {
	case goja.Value:
		return t.Export().(string)
	}
	return p["name"].(string)
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
	themeFs   fs.FS // 主题文件
	log       *zap.SugaredLogger
}

type Option struct {
	Fs      fs.FS
	ThemeFs fs.FS
}

type StdFileSystem struct {
}

func (f StdFileSystem) Open(name string) (fs.File, error) {
	//os.Stdout
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
		Debug:       false,
		Transformer: jsx.NewEsBuildTransform(false),
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

	x.RegisterModule("db", map[string]interface{}{
		"loadBlog": b.loadBlog,
	})

	return b, nil
}

type ExecOption struct {
	Env map[string]interface{}
	Log *zap.SugaredLogger
}

// Build 生成静态源文件
func (b *Bblog) Build(configFile string, distPath string, o ExecOption) error {
	start := time.Now()
	c, err := b.Load(configFile, o)
	if err != nil {
		return err
	}
	l := b.log
	if o.Log != nil {
		l = o.Log
	}

	for _, p := range c.Pages {
		var v, err = p.GetComponent()
		if err != nil {
			return err
		}
		body := v.Render()
		name := p.GetName()
		distFile := filepath.Join(distPath, "index.html")
		if name != "" && name != "index" {
			distFile = filepath.Join(distPath, name, "index.html")
		}
		dir := filepath.Dir(distFile)
		_ = os.MkdirAll(dir, 0755)

		err = ioutil.WriteFile(distFile, []byte(body), 0755)
		if err != nil {
			return err
		}

		l.Infof("create pages: %v ", distFile)
	}

	fe := fSExport{fs: b.themeFs}
	for _, a := range c.Assets {
		d := filepath.Dir(configFile)

		srcDir := filepath.Join(d, a)
		err = fe.exportDir(srcDir, distPath)
		if err != nil {
			return err
		}
		l.Infof("copy assets: %v ", a)
	}

	l.Infof("Done in %v", time.Now().Sub(start))
	return nil
}

type DirFs struct {
	prefix string
	fs     http.FileSystem
}

func (d *DirFs) Open(name string) (http.File, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := d.prefix
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
	f, err := d.fs.Open(fullName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Service 运行一个渲染程序
func (b *Bblog) Service(ctx context.Context, configFile string, o ExecOption, addr string, dev bool) error {
	s, err := NewService(addr)
	if err != nil {
		return err
	}
	var c Config
	var assetsHandler http.Handler
	prepare := func() error {
		c, err = b.Load(configFile, o)
		if err != nil {
			return err
		}

		var dirs MuitDir
		for _, i := range c.Assets {
			dirs = append(dirs, &DirFs{
				prefix: i,
				fs:     http.FS(b.themeFs),
			})
		}
		assetsHandler = http.FileServer(dirs)
		return nil
	}
	if !dev {
		err = prepare()
		if err != nil {
			return err
		}
	}

	s.Handler("/", func(writer http.ResponseWriter, request *http.Request) {
		if dev {
			b.x.RefreshRegistry(nil)
			err = prepare()
			if err != nil {
				writer.Write([]byte(err.Error()))
				return
			}
		}

		reqPath := strings.Trim(request.URL.Path, "/")
		//log.Printf("req: %v", reqPath)
		for _, p := range c.Pages {
			if reqPath == p.GetName() {
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
	Load(body []byte) *Blog
}

type MDBlogLoader struct {
	fs fs.FS
}

func (m *MDBlogLoader) Load(path string) (Blog, bool, error) {
	_, name := filepath.Split(path)

	ext := filepath.Ext(path)
	if supportExt[ext] {
		name = strings.TrimSuffix(name, ext)
	} else {
		return Blog{}, false, nil
	}

	// 读取 metadata
	body, err := fs.ReadFile(m.fs, path)
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

	return Blog{
		Name: name,
		GetContent: func() string {
			return string(blackfriday.Run(body))
		},
		Meta: meta,
		Ext:  ext,
	}, true, nil
}

// pp path
func (b *Bblog) loadBlog(pp string) interface{} {
	var blogs []Blog
	err := fs.WalkDir(b.projectFs, pp, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		loader := MDBlogLoader{fs: b.projectFs}

		blog, ok, _ := loader.Load(path)
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
		log.Errorf("get source '%s' error: %v", pp, err)
		return nil
	}

	return blogs
}

type Config struct {
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

func (b *Bblog) Load(configFile string, eo ExecOption) (Config, error) {
	envBs, _ := json.Marshal(eo.Env)
	processCode := fmt.Sprintf("var process = {env: %s}", envBs)

	v, err := b.x.RunJs("root.js", []byte(fmt.Sprintf(`%s;require("%v").default`, processCode, configFile)), false)
	if err != nil {
		return Config{}, err
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

	return Config{raw: raw, Pages: ps, Assets: assets}, nil
}
