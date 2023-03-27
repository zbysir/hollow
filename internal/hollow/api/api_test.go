package api

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/stretchr/testify/assert"
	"github.com/zbysir/hollow/internal/pkg/signal"
	"sync"
	"testing"
)

func TestEditor(t *testing.T) {
	e := NewEditor(func(pid int64) (billy.Filesystem, error) {
		return osfs.New("./testdata"), nil
	}, Config{PreviewDomain: "preview.blog.bysir.top", Secret: ""})

	ctx, c := signal.NewContext()
	defer c()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := e.Run(ctx, ":9091")
		if err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()
}

func TestMatchPreviewDomain(t *testing.T) {
	assert.Equal(t, true, matchDomain("bysir.top", "bysir.top"))
	assert.Equal(t, true, matchDomain("blog.bysir.top", "blog.bysir.top"))
	assert.Equal(t, true, matchDomain("*.bysir.top", "blog.bysir.top"))
	assert.Equal(t, false, matchDomain("*.bysir.top", "editor.blog.bysir.top"))
	assert.Equal(t, false, matchDomain("bysir.top", "blog.bysir.top"))
	assert.Equal(t, false, matchDomain("preview.blog.bysir.top", "editor.blog.bysir.top:9091"))
}
