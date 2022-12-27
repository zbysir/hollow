package md

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type ImageUrlReplace struct {
	urlReplace func(string) string
}

func NewImageUrlReplace(urlReplace func(string) string) *ImageUrlReplace {
	return &ImageUrlReplace{urlReplace: urlReplace}
}

func (i *ImageUrlReplace) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if i.urlReplace == nil {
		return
	}

	c := node.FirstChild()
	for c != nil {
		if img, ok := c.(*ast.Image); ok {
			img.Destination = []byte(i.urlReplace(string(img.Destination)))
		}
		if c.HasChildren() {
			c = c.FirstChild()
		} else {
			if c.NextSibling() != nil {
				c = c.NextSibling()
			} else {
				c = c.Parent().NextSibling()
			}
		}
	}
}
