package main

import (
	"github.com/zbysir/blog/internal/bblog"
)

func main() {
	b, err := bblog.NewBblog(bblog.Option{})
	if err != nil {
		panic(err)
	}

	err = b.Build("docs", bblog.ExecOption{})
	if err != nil {
		panic(err)
	}
}
