package bblog

import (
	"fmt"
	"github.com/zbysir/blog/internal/pkg/db"
	"github.com/zbysir/blog/internal/pkg/dbfs/stdfs"
	"testing"
)

func TestExport(t *testing.T) {
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
	//ft, err := fs.FileTree("./", -1)
	//if err != nil {
	//	t.Fatal(err)
	//}

	//t.Logf("%+v", ft)

	e := fSExport{fs: stdfs.NewFs(fsTheme)}
	err = e.exportDir("/", "../.cached")
	if err != nil {
		t.Fatal(err)
	}
}
