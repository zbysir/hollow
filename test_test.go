package main

import (
	"context"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/zbysir/hollow/internal/hollow"
	"github.com/zbysir/hollow/internal/pkg/db"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/log"
	"net/http"
	"testing"
)

func TestService(t *testing.T) {
	b, err := hollow.NewHollow(hollow.Option{
		SourceFs: osfs.New("./workspace"),
	})
	if err != nil {
		t.Fatal(err)
	}

	addr := ":8083"
	t.Logf("listening %v", addr)
	err = b.Service(context.Background(), hollow.ExecOption{IsDev: true}, addr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildAndPublish(t *testing.T) {
	b, err := hollow.NewHollow(hollow.Option{
		SourceFs: osfs.New("./workspace"),
	})
	if err != nil {
		t.Fatal(err)
	}

	dst := memfs.New()
	err = b.BuildAndPublish(hollow.NewRenderContext(), dst, hollow.ExecOption{
		Log: log.New(log.Options{
			IsDev:         false,
			DisableCaller: true,
			CallerSkip:    0,
			Name:          "",
			DisableTime:   true,
		}),
		IsDev: false,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuild(t *testing.T) {
	b, err := hollow.NewHollow(hollow.Option{
		SourceFs: osfs.New("/Users/bysir/goproj/bysir/zbysir.github.io"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = b.Build(hollow.NewRenderContext(), "./.dist", hollow.ExecOption{
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
	stp, err := d.Open("project_1", "project")
	if err != nil {
		t.Fatal(err)
	}
	fProject := gobilly.NewDbFs(stp)

	b, err := hollow.NewHollow(hollow.Option{SourceFs: fProject})
	if err != nil {
		panic(err)
	}

	err = b.Service(context.Background(), hollow.ExecOption{IsDev: true}, ":8083")
	if err != nil {
		panic(err)
	}
}

func TestFile(t *testing.T) {
	fs := http.FileServer(http.Dir("./.dist")) //去静态目录找 得到fs对象：文件服务器

	http.Handle("/", fs)

	http.ListenAndServe(":8090", nil)
}
