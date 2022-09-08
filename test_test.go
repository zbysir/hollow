package main

import (
	"context"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/zbysir/blog/internal/bblog"
	"github.com/zbysir/blog/internal/bblog/editor"
	"github.com/zbysir/blog/internal/pkg/db"
	"github.com/zbysir/blog/internal/pkg/gobilly"
	"testing"
)

func TestService(t *testing.T) {
	b, err := bblog.NewBblog(bblog.Option{
		Fs:      gobilly.NewStdFs(osfs.New("./workspace/project")),
		ThemeFs: gobilly.NewStdFs(osfs.New("./workspace/theme")),
	})
	if err != nil {
		t.Fatal(err)
	}

	addr := ":8082"
	t.Logf("listening %v", addr)
	err = b.Service(context.Background(), bblog.ExecOption{IsDev: true}, addr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildAndPublish(t *testing.T) {
	themeFs := gobilly.NewStdFs(osfs.New("./workspace/theme"))
	b, err := bblog.NewBblog(bblog.Option{
		Fs:      gobilly.NewStdFs(osfs.New("./workspace/project")),
		ThemeFs: themeFs,
	})
	if err != nil {
		t.Fatal(err)
	}

	dst := memfs.New()
	err = b.BuildAndPublish(dst, bblog.ExecOption{
		Log:   nil,
		IsDev: false,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuild(t *testing.T) {
	themeFs := gobilly.NewStdFs(osfs.New("./workspace/theme"))
	b, err := bblog.NewBblog(bblog.Option{
		Fs:      gobilly.NewStdFs(osfs.New("./workspace/project")),
		ThemeFs: themeFs,
	})
	if err != nil {
		t.Fatal(err)
	}

	dst := memfs.New()
	err = b.BuildToFs(dst, bblog.ExecOption{
		Log:   nil,
		IsDev: false,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceDbFs(t *testing.T) {
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

	err = b.Service(context.Background(), bblog.ExecOption{IsDev: true}, ":8083")
	if err != nil {
		panic(err)
	}
}

func TestEditor(t *testing.T) {
	e := editor.NewEditor(func(pid int64) (billy.Filesystem, error) {
		return osfs.New("./workspace/project"), nil
	}, func(pid int64) (billy.Filesystem, error) {
		return osfs.New("./workspace/theme"), nil
	})
	err := e.Run(nil, ":9091")
	if err != nil {
		t.Fatal(err)
	}
}
