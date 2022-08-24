package bblog

import (
	"fmt"
	"github.com/zbysir/blog/internal/pkg/db"
	"github.com/zbysir/blog/internal/pkg/dbfs/stdfs"
	"testing"
)

func TestSource(t *testing.T) {
	b, err := NewBblog(Option{})
	if err != nil {
		t.Fatal(err)
	}

	x := b.loadBlog("../../blogs")
	bs := x.([]Blog)
	for _, b := range bs {
		t.Logf("%+v", b)

		t.Logf("content: %s", b.GetContent())
	}
}

func TestLoad(t *testing.T) {
	b, err := NewBblog(Option{})
	if err != nil {
		t.Fatal(err)
	}

	c, err := b.Load("./src/config.ts", ExecOption{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", c)
}

func TestExportFs(t *testing.T) {
	d, err := db.NewKvDb("./editor/database")
	if err != nil {
		t.Fatal(err)
	}
	st, err := d.Open(fmt.Sprintf("project_1"), "theme")
	if err != nil {
		t.Fatal(err)
	}
	fsTheme, err := fusefs.NewDbFs(st)
	if err != nil {
		t.Fatal(err)
	}

	st2, err := d.Open(fmt.Sprintf("project_1"), "project")
	if err != nil {
		t.Fatal(err)
	}
	fs, err := fusefs.NewDbFs(st2)
	if err != nil {
		t.Fatal(err)
	}

	b, err := NewBblog(Option{
		Fs:      stdfs.NewFs(fs),
		ThemeFs: stdfs.NewFs(fsTheme),
	})

	err = b.Build("./config.ts", "docs", ExecOption{
		Env: map[string]interface{}{"base": "/blog"},
		//Out: &WsSink{hub: hub},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
}
