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
	b, err := NewBblog(Option{
		Fs: osfs.New("../../docs"),
	})
	if err != nil {
		t.Fatal(err)
	}

	as := b.getArticles("contents", getBlogOption{
		Sort: func(a, b interface{}) bool {
			//c := a.(ArticleTree).Meta["date"]
			//a.(ArticleTree).Meta
			//log.Infof("c %+v", a)
			return false
		},
		Flat: false,
	})
	for _, b := range as.List {
		t.Logf("%+v %v", b.Name, b.IsDir)
		for _, b := range b.Children {
			t.Logf("\t\t%+v %v", b.Name, b.IsDir)
		}
	}

	//bs, err := json.MarshalIndent(as.List, " ", " ")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//t.Logf("%s", bs)
}

func TestLoad(t *testing.T) {
	b, err := NewBblog(Option{
		Fs: osfs.New("../../workspace"),
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err := b.LoadConfig(true)
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

	st2, err := d.Open(fmt.Sprintf("project_1"), "project")
	if err != nil {
		t.Fatal(err)
	}
	fs := gobilly.NewDbFs(st2)

	b, err := NewBblog(Option{
		Fs: fs,
	})

	err = b.Build("docs", ExecOption{
		//Out: &WsSink{hub: hub},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
}
