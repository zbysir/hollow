package hollow

import (
	"github.com/go-git/go-billy/v5/memfs"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"testing"
)

func TestNewMDLoader(t *testing.T) {
	jx, err := jsx.NewJsx(jsx.Option{
		SourceCache: nil,
		Debug:       false,
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	f := memfs.New()
	file, _ := f.Create("index.mdx")
	file.Write([]byte(`---
title: "title"
---

const A = ()=> <> AAA </>

## h1 <A></A>

![](../../statics/img/img.jpg)

`))

	m := NewMDLoader(Assets{"statics"}, jx)
	x, err := m.Load(gobilly.NewStdFs(f), "index.mdx", false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", x.Content)
}
