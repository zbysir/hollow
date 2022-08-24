package main

import (
	"context"
	"github.com/zbysir/blog/internal/bblog"
	"github.com/zbysir/blog/internal/pkg/db"
	"github.com/zbysir/blog/internal/pkg/gobilly"
	"testing"
)

func TestService(t *testing.T) {
	b, err := bblog.NewBblog(bblog.Option{})
	if err != nil {
		panic(err)
	}

	err = b.Service(context.Background(), "./template/default/config.ts", bblog.ExecOption{}, ":8082", true)
	if err != nil {
		panic(err)
	}
}

func TestServiceFs(t *testing.T) {
	d, err := db.NewKvDb("./internal/bblog/editor/database")
	if err != nil {
		t.Fatal(err)
	}
	st, err := d.Open("project_1", "theme")
	if err != nil {
		t.Fatal(err)
	}
	fa := gobilly.NewDbFs(st)
	stp, err := d.Open("project_1", "project")
	if err != nil {
		t.Fatal(err)
	}
	fProject := gobilly.NewDbFs(stp)

	b, err := bblog.NewBblog(bblog.Option{ThemeFs: gobilly.NewStdFs(fa), Fs: gobilly.NewStdFs(fProject)})
	if err != nil {
		panic(err)
	}

	err = b.Service(context.Background(), "./config.ts", bblog.ExecOption{}, ":8083", true)
	if err != nil {
		panic(err)
	}
}
