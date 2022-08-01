package main

import (
	"github.com/zbysir/blog/internal/bblog"
)

func main() {
	b, err := bblog.NewBblog("./src/config.ts", bblog.Option{})
	if err != nil {
		panic(err)
	}

	err = b.Export("docs")
	if err != nil {
		panic(err)
	}
}
