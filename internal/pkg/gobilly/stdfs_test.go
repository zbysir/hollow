package gobilly

import (
	"github.com/zbysir/hollow/internal/pkg/db"
	"github.com/zbysir/hollow/internal/pkg/log"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestStdFs(t *testing.T) {
	log.SetDev(true)

	d, err := db.NewKvDb("")
	if err != nil {
		t.Fatal(err)
	}
	st, err := d.Open("test", "theme")
	if err != nil {
		t.Fatal(err)
	}

	f := NewDbFs(st)
	std := NewStdFs(f)

	t.Run("walk", func(t *testing.T) {
		err = fs.WalkDir(std, "", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%v %v %v", d.IsDir(), path, d.Name())
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("walkos", func(t *testing.T) {
		err = filepath.WalkDir("./", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%v %v %v", d.IsDir(), path, d.Name())
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	})

}
