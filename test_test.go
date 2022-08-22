package main

import (
	"context"
	"github.com/zbysir/blog/internal/bblog"
	"github.com/zbysir/blog/internal/fusefs/stdfs"
	"testing"
)

func TestService(t *testing.T) {
	b, err := bblog.NewBblog(bblog.Option{})
	if err != nil {
		panic(err)
	}

	err = b.Service(context.Background(), "./template/default/config.ts", bblog.ExecOption{}, ":8083", true)
	if err != nil {
		panic(err)
	}
}

func TestServiceFs(t *testing.T) {
	fa, err := bblog.NewFuseFs("./internal/bblog/db.db", "1.theme")
	if err != nil {
		t.Fatal(err)
	}

	fProject, err := bblog.NewFuseFs("./internal/bblog/db.db", "1.project")
	if err != nil {
		t.Fatal(err)
	}

	b, err := bblog.NewBblog(bblog.Option{ThemeFs: stdfs.NewFs(fa), Fs: stdfs.NewFs(fProject)})
	if err != nil {
		panic(err)
	}

	err = b.Service(context.Background(), "./config.ts", bblog.ExecOption{}, ":8083", true)
	if err != nil {
		panic(err)
	}
}
