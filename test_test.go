package main

import (
	"context"
	"github.com/zbysir/blog/internal/bblog"
	"github.com/zbysir/blog/internal/pkg/db"
	"github.com/zbysir/blog/internal/pkg/dbfs"
	"github.com/zbysir/blog/internal/pkg/dbfs/stdfs"
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
	d, err := db.NewKvDb("./internal/bblog/editor/database")
	if err != nil {
		t.Fatal(err)
	}
	st, err := d.Open("project_1", "theme")
	if err != nil {
		t.Fatal(err)
	}
	fa, err := dbfs.NewDbFs(st)
	if err != nil {
		t.Fatal(err)
	}
	stp, err := d.Open("project_1", "project")
	if err != nil {
		t.Fatal(err)
	}
	fProject, err := dbfs.NewDbFs(stp)
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
