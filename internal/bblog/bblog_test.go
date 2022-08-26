package bblog

import (
	"fmt"
	"github.com/zbysir/blog/internal/pkg/db"
	"github.com/zbysir/blog/internal/pkg/gobilly"
	"testing"
)

func TestSource(t *testing.T) {
	b, err := NewBblog(Option{})
	if err != nil {
		t.Fatal(err)
	}

	x := b.getBlog("../../blogs")
	bs := x
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

	c, err := b.loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	th, err := b.loadTheme("./src/config.ts", c)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", th)
}

func TestBuildFromFs(t *testing.T) {
	d, err := db.NewKvDb("./editor/database")
	if err != nil {
		t.Fatal(err)
	}
	st, err := d.Open(fmt.Sprintf("project_1"), "theme")
	if err != nil {
		t.Fatal(err)
	}
	fsTheme := gobilly.NewDbFs(st)

	st2, err := d.Open(fmt.Sprintf("project_1"), "project")
	if err != nil {
		t.Fatal(err)
	}
	fs := gobilly.NewDbFs(st2)

	b, err := NewBblog(Option{
		Fs:      gobilly.NewStdFs(fs),
		ThemeFs: gobilly.NewStdFs(fsTheme),
	})

	err = b.Build("docs", ExecOption{
		//Out: &WsSink{hub: hub},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
}
