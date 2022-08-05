package fs

import (
	"github.com/docker/libkv/store"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}
func TestKvFs(t *testing.T) {
	fs, err := NewKVFS(Options{
		Store: string(store.BOLTDB),
		Addrs: []string{"./db.db"},
		Root:  "",
		Config: store.Config{
			Bucket: "test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	//_, status := fs.Create("/", 0, 0, nil)
	//if !status.Ok() {
	//	t.Fatal(status)
	//}
	t.Logf("%+v", filepath.Join(""))
	{
		status := fs.Mkdir("", 0, nil)
		if !status.Ok() {
			t.Fatal(status)
		}
	}

	{
		status := fs.Mkdir("src", 0, nil)
		if !status.Ok() {
			t.Fatal(status)
		}
	}
	{
		status := fs.Mkdir("public", 0, nil)
		if !status.Ok() {
			t.Fatal(status)
		}
	}
	{
		nf, status := fs.Create("main.go", 0, 0, nil)
		if !status.Ok() {
			t.Fatal(status)
		}
		t.Logf("%+v", nf)
	}
	//{
	//	nf, status := fs.Create("public/css.css", 0, 0, nil)
	//	if !status.Ok() {
	//		t.Fatal(status)
	//	}
	//	t.Logf("%+v", nf)
	//}

	t.Run("root", func(t *testing.T) {
		ds, status := fs.OpenDir("", nil)
		if !status.Ok() {
			t.Fatal(status)
		}

		for _, d := range ds {
			if d.Mode == fuse.S_IFDIR {
				t.Logf("dir : %+v", d)
			} else {
				t.Logf("file: %+v", d)
			}
		}
	})

	t.Run("public", func(t *testing.T) {
		ds, status := fs.OpenDir("public", nil)
		if !status.Ok() {
			t.Fatal(status)
		}

		for _, d := range ds {
			if d.Mode&fuse.S_IFDIR == fuse.S_IFDIR {
				t.Logf("dir : %+v", d)
			} else {
				t.Logf("file: %+v", d)
			}
		}
	})
	t.Run("create", func(t *testing.T) {
		{
			nf, status := fs.Create("public/css.css", 0, 0, nil)
			if !status.Ok() {
				t.Fatal(status)
			}
			t.Logf("%+v", nf)
		}
	})

	t.Run("write", func(t *testing.T) {

		f, _ := fs.Open("public/css.css", 0, nil)
		_, _ = f.Write([]byte("body {}"), 0)
		//		t.Logf("%+v", n)

		// attr
		var attr fuse.Attr
		status := f.GetAttr(&attr)
		if !status.Ok() {
			t.Fatal(status)
		}
		t.Logf("%+v", attr)

		// read
		bs, err := ioutil.ReadAll(&IOReader{f: f, curr: 0})
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%s", bs)
	})

	t.Run("open", func(t *testing.T) {
		f, _ := fs.Open("public/css.css", 0, nil)

		// attr
		var attr fuse.Attr
		status := f.GetAttr(&attr)
		if !status.Ok() {
			t.Fatal(status)
		}
		t.Logf("%+v", attr)

		// read
		bs, err := ioutil.ReadAll(&IOReader{f: f, curr: 0})
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%s", bs)
	})

}
