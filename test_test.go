package main

import (
	"context"
	"github.com/zbysir/blog/internal/bblog"
	"testing"
)

func TestService(t *testing.T) {
	b, err := bblog.NewBblog("./src/config.ts", bblog.Option{})
	if err != nil {
		panic(err)
	}

	err = b.Service(context.Background(), ":8083", true)
	if err != nil {
		panic(err)
	}
}
