package bblog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/russross/blackfriday/v2"
	jsx "github.com/zbysir/gojsx"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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
	x  *jsx.Jsx
	fs fs.FS
}

type Option struct {
	Fs fs.FS
}

type StdFileSystem struct {
}

func (f StdFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func NewBblog(o Option) (*Bblog, error) {
	var err error
	x, err := jsx.NewJsx(jsx.Option{
		SourceCache: jsx.NewFileCache("./.cache"),
		SourceFs:    nil,
		Debug:       true,
		Transformer: jsx.NewEsBuildTransform(false),
	})
	if err != nil {
		panic(err)
	}

	if o.Fs == nil {
		o.Fs = StdFileSystem{}
	}
	b := &Bblog{
		x:  x,
		fs: o.Fs,
	}

	x.RegisterModule("db", map[string]interface{}{
		"getSource": b.getSource,
	})

	return b, nil
}

type ExecOption struct {
	Env map[string]interface{}
}

func (b *Bblog) Export(configFile string, distPath string, o ExecOption) error {
	c, err := b.Load(configFile, o)
	if err != nil {
		return err
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
		_ = os.MkdirAll(dir, os.ModePerm)

		err = ioutil.WriteFile(distFile, []byte(body), os.ModePerm)
		if err != nil {
			return err
		}

		log.Printf("create pages: %v ", distFile)
	}

	fe := fSExport{fs: b.fs}
	for _, a := range c.Assets {
		d := filepath.Dir(configFile)
		err = fe.exportDir(filepath.Join(d, a), distPath)
		if err != nil {
			return err
		}
		log.Printf("copy assets: %v ", a)
	}
	return nil
}

//var memoCache = map[string]interface{}{}
//func useMemo[T](key string, t T,dep ...interface{})(a T){
//
//}

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
		base, _ := filepath.Split(configFile)

		var dirs MuitDir
		for _, i := range c.Assets {
			dirs = append(dirs, http.Dir(filepath.Join(base, i)))
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
			b.x.RefreshRegistry()
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

type MuitDir []http.Dir

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
	GetContent interface{}            `json:"getContent"`
	Meta       map[string]interface{} `json:"meta"`
	Ext        string                 `json:"ext"`
}

type BlogLoader interface {
	Load(body []byte) *Blog
}

type MDBlogLoader struct {
	fs fs.FS
}

func (m *MDBlogLoader) Load(path string) (Blog, bool) {
	_, name := filepath.Split(path)

	ext := filepath.Ext(path)
	if supportExt[ext] {
		name = strings.TrimSuffix(name, ext)
	} else {
		return Blog{}, false
	}

	return Blog{
		Name: name,
		GetContent: func() string {
			//log.Printf("getContent %v", path)
			source, err := fs.ReadFile(m.fs, path)
			if err != nil {
				panic(err)
			}
			source = blackfriday.Run(source)
			return string(source)
		},
		Meta: nil,
		Ext:  ext,
	}, true
}

// pp path
func (b *Bblog) getSource(pp string) interface{} {
	var blogs []Blog
	err := fs.WalkDir(b.fs, pp, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		loader := MDBlogLoader{fs: b.fs}

		blog, ok := loader.Load(path)
		if !ok {
			return nil
		}
		// read meta
		metaFileName := path + ".yaml"
		bs, err := fs.ReadFile(b.fs, metaFileName)
		if err != nil {
			return fmt.Errorf("read meta file error: %w", err)
		}
		var m = map[string]interface{}{}
		err = yaml.Unmarshal(bs, &m)
		if err != nil {
			return fmt.Errorf("unmarshal meta file error: %w", err)
		}
		blog.Meta = m

		blogs = append(blogs, blog)

		return nil
	})
	if err != nil {
		panic(err)
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
