package bblog

import (
	"github.com/russross/blackfriday/v2"
	"io"
	"path/filepath"
	"strings"
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
		//default:
		//	log.Infof("%+v %+v", node.Type, node)
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
	dir := "/blogs/2020"
	return newMdRender(func(s string) string {
		p := s
		if filepath.IsAbs(s) {
		} else {
			p = filepath.Join(dir, s)
		}

		assets := []string{"static"}
		// 移除 assets 文件夹前缀
		for _, a := range assets {
			if strings.HasPrefix(p, "/"+a) {
				p = strings.TrimPrefix(p, "/"+a)
				break
			}
		}

		return p
	}).Render(src)
}
