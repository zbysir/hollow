package main

import (
	"github.com/russross/blackfriday/v2"
	"github.com/zbysir/blog/internal/bblog"
	"io/fs"
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
			"getContent": func() string {
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
	b, err := bblog.NewBblog("./src/config.ts")
	if err != nil {
		panic(err)
	}

	err = b.Export("docs")
	if err != nil {
		panic(err)
	}
}
