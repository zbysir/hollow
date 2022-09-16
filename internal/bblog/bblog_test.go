package bblog

import (
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/zbysir/hollow/internal/pkg/db"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"path/filepath"
	"testing"
)

func TestSource(t *testing.T) {
	b, err := NewBblog(Option{})
	if err != nil {
		t.Fatal(err)
	}

	x := b.getArticles("../../blogs", getBlogOption{})
	bs := x
	for _, b := range bs.List {
		t.Logf("%+v", b)

		t.Logf("content: %s", b.GetContent())
	}
}

func TestLoad(t *testing.T) {
	b, err := NewBblog(Option{
		Fs:      osfs.New("../../workspace/project"),
		ThemeFs: osfs.New("../../workspace/theme"),
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err := b.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	themeDir := c.Theme

	configFile := filepath.Join(themeDir, "config.tsx")

	th, err := b.loadTheme(configFile)
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
		Fs:      fs,
		ThemeFs: fsTheme,
	})

	err = b.Build("docs", ExecOption{
		//Out: &WsSink{hub: hub},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
}
