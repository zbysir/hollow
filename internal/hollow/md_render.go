package hollow

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"io"
)

type mdRender struct {
	inner            markdown.Renderer
	assetsUrlProcess func(string) string
}

func (m *mdRender) RenderNode(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch t := node.(type) {
	case *ast.Image:
		t.Destination = []byte(m.assetsUrlProcess(string(t.Destination)))
		return m.inner.RenderNode(w, t, entering), true
	}
	return 0, false
}

func newMdRender(assetsUrlProcess func(string) string) *mdRender {
	return &mdRender{
		inner:            html.NewRenderer(html.RendererOptions{}),
		assetsUrlProcess: assetsUrlProcess,
	}
}

func (m *mdRender) Render(src []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	pars := parser.NewWithExtensions(extensions)
	return markdown.ToHTML(src, pars, html.NewRenderer(html.RendererOptions{
		AbsolutePrefix:             "",
		FootnoteAnchorPrefix:       "",
		FootnoteReturnLinkContents: "",
		CitationFormatString:       "",
		HeadingIDPrefix:            "",
		HeadingIDSuffix:            "",
		Title:                      "",
		CSS:                        "",
		Icon:                       "",
		Head:                       nil,
		Flags:                      0,
		RenderNodeHook:             m.RenderNode,
		Comments:                   nil,
		Generator:                  "",
	}))
}

// renderMd 渲染 md 片段，不会处理其中的图片 url（因为没有上下文）
func renderMd(src []byte) []byte {
	return newMdRender(func(s string) string {
		return s
	}).Render(src)
}
