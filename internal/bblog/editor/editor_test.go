package editor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEditor(t *testing.T) {
	e := NewEditor(nil, nil, Config{PreviewDomain: "abc"})
	err := e.Run(nil, ":9091")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMatchPreviewDomain(t *testing.T) {
	assert.Equal(t, true, matchDomain("bysir.top", "bysir.top"))
	assert.Equal(t, true, matchDomain("blog.bysir.top", "blog.bysir.top"))
	assert.Equal(t, true, matchDomain("*.bysir.top", "blog.bysir.top"))
	assert.Equal(t, false, matchDomain("*.bysir.top", "editor.blog.bysir.top"))
	assert.Equal(t, false, matchDomain("bysir.top", "blog.bysir.top"))
	assert.Equal(t, false, matchDomain("preview.blog.bysir.top", "editor.blog.bysir.top:9091"))
}
