package main

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/russross/blackfriday/v2"
	jsx "github.com/zbysir/gojsx"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func getBlog(pp string) interface{} {
	var blogs []map[string]interface{}
	err := filepath.WalkDir(pp, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		_, name := filepath.Split(path)

		blogs = append(blogs, map[string]interface{}{
			"name": name,
			"link": fmt.Sprintf("%v.html", name),
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

func main() {
	var err error
	x, err := jsx.NewJsx(jsx.WithSourceCache(jsx.NewFileCache("./.cache")), jsx.WithDebug(true))
	if err != nil {
		panic(err)
	}

	x.RegisterModule("db", map[string]interface{}{
		"getBlog": getBlog,
	})

	v, err := x.RunJs("./src/config.ts", []byte(`require("./src/config.ts").default`), false)
	if err != nil {
		panic(err)
	}
	m := v.Export().(map[string]interface{})

	pages := m["pages"].([]interface{})
	for _, p := range pages {
		x := p.(map[string]interface{})
		var v jsx.VDom
		switch t := x["component"].(type) {
		case map[string]interface{}:
			// for: component: Index(props)
			v = t
		case func(goja.FunctionCall) goja.Value:
			// for: component: () => Index(props)
			v = t(goja.FunctionCall{}).Export().(map[string]interface{})
		}
		body := v.Render()

		name := x["name"].(string)
		distFile := filepath.Join("dist", name+".html")
		dir := filepath.Dir(distFile)
		_ = os.MkdirAll(dir, os.ModePerm)

		err = ioutil.WriteFile(distFile, []byte(body), os.ModePerm)
		if err != nil {
			panic(err)
		}

		log.Printf("create: %v ", distFile)
	}
}
