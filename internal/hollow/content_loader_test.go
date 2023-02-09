package hollow

import (
	"github.com/go-git/go-billy/v5/memfs"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"testing"
)

func TestNewMDLoader(t *testing.T) {

	f := memfs.New()
	file, _ := f.Create("index.mdx")
	file.Write([]byte(`---
title: "title"
---

const A = ()=> <> AAA </>

## h1 <A></A>

![](../../statics/img/img.jpg)

![](img.jpg)

`))

	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: nil,
		Debug:       false,
		Fs:          gobilly.NewStdFs(f),
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	m := NewMDLoader(Assets{"statics"}, jx, nil)

	c, err := m.Load("index.mdx", false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", c.Assets)
	t.Logf("%+v", c.Content)
}
