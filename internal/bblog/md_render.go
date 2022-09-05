package bblog

import (
	"fmt"
	"github.com/russross/blackfriday/v2"
	"github.com/zbysir/blog/internal/pkg/log"
	"io"
	"path/filepath"
)

type mdRender struct {
	inner            blackfriday.Renderer
	assetsUrlProcess func(string) string
}

func (m *mdRender) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	switch node.Type {
	case blackfriday.Image:
		if entering {
			if m.assetsUrlProcess != nil {
				node.LinkData.Destination = []byte(m.assetsUrlProcess(string(node.LinkData.Destination)))
			}
		}
	}
	return m.inner.RenderNode(w, node, entering)
}
func (m *mdRender) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	m.inner.RenderHeader(w, ast)
}
func (m *mdRender) RenderFooter(w io.Writer, ast *blackfriday.Node) {
	m.inner.RenderFooter(w, ast)
}

func newMdRender(assetsUrlProcess func(string) string) *mdRender {
	r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	return &mdRender{inner: r, assetsUrlProcess: assetsUrlProcess}
}
func (m *mdRender) Render(src []byte) []byte {
	return blackfriday.Run(src, blackfriday.WithRenderer(m))
}

func renderMd(src []byte) []byte {
	dir := "blogs"
	return newMdRender(func(s string) string {
		log.Infof("path %+v", s)

		rel, _ := filepath.Rel(dir, filepath.Join(s))
		log.Infof("path %+v", rel)

		return fmt.Sprintf("/%v" + rel)
	}).Render(src)
}
