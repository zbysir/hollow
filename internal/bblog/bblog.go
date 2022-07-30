package bblog

import (
	"context"
	"fmt"
	"github.com/dop251/goja"
	"github.com/russross/blackfriday/v2"
	jsx "github.com/zbysir/gojsx"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Page map[string]interface{}
type Pages []Page

func (p Page) GetName() string {
	return p["name"].(string)
}
func (p Page) GetComponent() jsx.VDom {
	var v jsx.VDom
	switch t := p["component"].(type) {
	case map[string]interface{}:
		// for: component: Index(props)
		v = t
	case func(goja.FunctionCall) goja.Value:
		// for: component: () => Index(props)
		v = t(goja.FunctionCall{}).Export().(map[string]interface{})
	}

	return v
}

type Bblog struct {
	x          *jsx.Jsx
	configFile string
}

func NewBblog(configFile string) (*Bblog, error) {
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

	b := &Bblog{
		x:          x,
		configFile: configFile,
	}

	x.RegisterModule("db", map[string]interface{}{
		"getSource": b.getSource,
	})

	return b, nil
}

func (b *Bblog) Export(distPath string) error {
	c, err := b.Load(b.configFile)
	if err != nil {
		return err
	}

	for _, p := range c.Pages {
		var v = p.GetComponent()
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
			panic(err)
		}

		log.Printf("create: %v ", distFile)
	}
	return nil
}

//var memoCache = map[string]interface{}{}
//func useMemo[T](key string, t T,dep ...interface{})(a T){
//
//}

func (b *Bblog) Service(ctx context.Context, addr string, dev bool) error {
	s, err := NewService(addr)
	if err != nil {
		return err
	}
	var c Config
	var assetsHandler http.Handler
	prepare := func() error {
		c, err = b.Load(b.configFile)
		if err != nil {
			return err
		}
		base, _ := filepath.Split(b.configFile)

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
				x := p.GetComponent().Render()
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

// pp path
func (b *Bblog) getSource(pp string) interface{} {
	var blogs []map[string]interface{}
	err := filepath.WalkDir(pp, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		_, name := filepath.Split(path)

		blogs = append(blogs, map[string]interface{}{
			"name": name,
			"getContent": func() string {
				//log.Printf("getContent %v", path)
				source, err := os.ReadFile(path)
				if err != nil {

					panic(err)
				}
				source = blackfriday.Run(source)
				return string(source)
			},
		})

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

func (b *Bblog) Load(configFile string) (Config, error) {
	v, err := b.x.RunJs(configFile, []byte(fmt.Sprintf(`require("%v").default`, configFile)), false)
	if err != nil {
		return Config{}, err
	}

	raw := v.Export().(map[string]interface{})

	pages := raw["pages"].([]interface{})
	ps := make(Pages, len(pages))
	for i, p := range pages {
		ps[i] = p.(map[string]interface{})
	}
	as := raw["assets"].([]interface{})
	assets := make(Assets, len(as))
	for k, v := range as {
		assets[k] = v.(string)
	}

	return Config{raw: raw, Pages: ps, Assets: assets}, nil
}
