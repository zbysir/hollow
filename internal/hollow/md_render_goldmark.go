package hollow

import (
	"bytes"
	"github.com/gomarkdown/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	jsx "github.com/zbysir/gojsx"
	md2 "github.com/zbysir/hollow/internal/pkg/md"
	"io/fs"
)

type mdRenderGold struct {
	inner            markdown.Renderer
	assetsUrlProcess func(string) string
	jsx              *jsx.Jsx
	fs               fs.FS
}

func newGoldMdRender(jsx *jsx.Jsx, fs fs.FS) *mdRenderGold {
	return &mdRenderGold{jsx: jsx, fs: fs}
}

func (m *mdRenderGold) Render(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	context := parser.NewContext()
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			md2.Meta,
			md2.NewJsx(m.jsx, m.fs),
		),
	)

	if err := md.Convert(src, &buf, parser.WithContext(context)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
