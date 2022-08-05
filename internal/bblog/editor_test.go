package bblog

import (
	"encoding/json"
	"github.com/docker/libkv/store"
	"github.com/zbysir/blog/internal/fs"
	"testing"
)

func TestFileTree(t *testing.T) {
	kvfs, err := fs.NewKVFS(fs.Options{
		Store: string(store.BOLTDB),
		Addrs: []string{"./db.db"},
		Root:  "",
		Config: store.Config{
			Bucket: "test",
		},
	})
	fa := FsApi{fs: kvfs}
	ft, err := fa.FileTree("/", 1)
	if err != nil {
		t.Fatal(err)
	}

	fa.RmFile("")

	bs, _ := json.MarshalIndent(ft, " ", " ")
	t.Logf("%s", bs)
}

func TestEditor(t *testing.T) {
	e := Editor{}
	err := e.Run(nil)
	if err != nil {
		t.Fatal(err)
	}
}
