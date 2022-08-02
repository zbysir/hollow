package main

import (
	"github.com/zbysir/blog/internal/bblog"
)

func main() {
	b, err := bblog.NewBblog(bblog.Option{})
	if err != nil {
		panic(err)
	}

	err = b.Export("./template/default/config.ts", "docs", bblog.ExecOption{
		Env: map[string]interface{}{"base": "/blog"},
	})
	if err != nil {
		panic(err)
	}
}
