package easyfs

import (
	"encoding/json"
	"github.com/docker/libkv/store"
	"github.com/sirupsen/logrus"
	"github.com/zbysir/blog/internal/fusefs"
	"testing"
)

const dbFile = "../../bblog/db.db"

func TestFileTree(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	kvfs, err := fusefs.NewFuseFs(fusefs.Options{
		Store: string(store.BOLTDB),
		Addrs: []string{dbFile},
		Root:  "",
		Config: store.Config{
			Bucket: "1.theme",
		},
	})
	fa := NewFs(kvfs)

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

//func TestFileList(t *testing.T) {
//	logrus.SetLevel(logrus.DebugLevel)
//	kvfs, err := fusefs.NewFuseFs(fusefs.Options{
//		Store: string(store.BOLTDB),
//		Addrs: []string{dbFile},
//		Root:  "",
//		Config: store.Config{
//			Bucket: "1.theme",
//		},
//	})
//	fa := NewFs(kvfs)
//
//	//err = fa.Mkdir("/")
//	//if err != nil {
//	//	t.Fatal(err)
//	//}
//	//fa.fs.KvDelete("//")
//	//err = fa.RmFile("")
//	//if err != nil {
//	//	t.Fatal(err)
//	//}
//
//	//err = fa.Mkdir("src")
//	//if err != nil {
//	//	t.Fatal(err)
//	//}
//	//err = fa.Mkdir("src/js")
//	//if err != nil {
//	//	t.Fatal(err)
//	//}
//	//fa.WriteFile("src/js/inde.js", "1")
//
//	ft, err := fa.FileList("/")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	bs, _ := json.MarshalIndent(ft, " ", " ")
//	t.Logf("%s", bs)
//}
