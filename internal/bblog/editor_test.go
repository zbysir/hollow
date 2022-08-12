package bblog

import (
	"encoding/json"
	"github.com/docker/libkv/store"
	"github.com/sirupsen/logrus"
	"github.com/zbysir/blog/internal/fs"
	"testing"
)

func TestFileTree(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	kvfs, err := fs.NewKVFS(fs.Options{
		Store: string(store.BOLTDB),
		Addrs: []string{"./db.db"},
		Root:  "",
		Config: store.Config{
			Bucket: "test",
		},
	})
	fa := FsApi{fs: kvfs}

	//err = fa.Mkdir("src")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//err = fa.Mkdir("src/js")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fa.WriteFile("src/js/inde.js", "1")

	ft, err := fa.FileTree("/", 4)
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(ft, " ", " ")
	t.Logf("%s", bs)
}

func TestFileList(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	kvfs, err := fs.NewKVFS(fs.Options{
		Store: string(store.BOLTDB),
		Addrs: []string{"./db.db"},
		Root:  "",
		Config: store.Config{
			Bucket: "test",
		},
	})
	fa := FsApi{fs: kvfs}

	//err = fa.Mkdir("src")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//err = fa.Mkdir("src/js")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fa.WriteFile("src/js/inde.js", "1")

	ft, err := fa.FileList("/")
	if err != nil {
		t.Fatal(err)
	}

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
