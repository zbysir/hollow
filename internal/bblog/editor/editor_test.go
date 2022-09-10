package editor

import (
	"github.com/zbysir/hollow/internal/bblog/storage"
	"github.com/zbysir/hollow/internal/pkg/db"
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

	// st, err := a.db.Open(fmt.Sprintf("project_%v", pid), bucket)
	//	if err != nil {
	//		return nil, err
	//	}
	//	fs := gobilly.NewDbFs(st)
	//	if err != nil {
	//		return nil, fmt.Errorf("new fs error: %w", err)
	//	}
}
