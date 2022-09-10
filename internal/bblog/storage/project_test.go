package storage

import (
	"github.com/zbysir/hollow/internal/pkg/db"
	"testing"
)

func TestNewProject(t *testing.T) {
	kvDb, err := db.NewKvDb("./database")
	if err != nil {
		t.Fatal(err)
	}

	store, err := kvDb.Open("main", "default")
	if err != nil {
		t.Fatal(err)
	}

	p := NewProject(store)
	s, exist, err := p.GetSetting(1)
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Logf("%v %+v", exist, s)

	err = p.SetSetting(1, &ProjectSetting{
		GitRemote: "https://github.com/zbysir/zbysir.github.io.git",
		GitToken:  "x",
		ThemeId:   0,
	})
	if err != nil {
		t.Fatal(err)
	}

	s, exist, err = p.GetSetting(1)
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Logf("%v %+v", exist, s)
}
