package editor

import (
	"github.com/zbysir/blog/internal/bblog/storage"
	"github.com/zbysir/blog/internal/pkg/db"
	"testing"
)

func TestEditor(t *testing.T) {
	d, err := db.NewKvDb("./database")
	if err != nil {
		t.Fatal(err)
	}

	st, err := d.Open("main", "default")
	if err != nil {
		t.Fatal(err)
	}
	e := NewEditor(d, storage.NewProject(st))
	err = e.Run(nil, ":9091")
	if err != nil {
		t.Fatal(err)
	}
}
