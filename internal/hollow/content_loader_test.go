package hollow

import (
	"github.com/go-git/go-billy/v5/memfs"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"sync"
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
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := m.Load("index.mdx", false)
			if err != nil {
				t.Fatal(err)
			}

		}()
	}

	wg.Wait()
}
