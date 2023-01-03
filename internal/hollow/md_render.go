package hollow

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"github.com/zbysir/gojsx/pkg/mdx"
	"github.com/zbysir/hollow/internal/pkg/mdext"
)

type mdRenderGold struct {
	GoldMdRenderOptions
}

type GoldMdRenderOptions struct {
	// Jsx Element 的渲染方法，如果传递则会解析并渲染 Jsx，如果是空则只会按照 md 格式处理。
	JsxRender        renderer.NodeRendererFunc
	AssetsUrlProcess func(string) string
}

func NewGoldMdRender(o GoldMdRenderOptions) *mdRenderGold {
	return &mdRenderGold{GoldMdRenderOptions: o}
}

type MdResult struct {
	Body []byte
	Meta map[string]interface{}
}

func (m *mdRenderGold) Render(src []byte) (MdResult, error) {
	var extenders = []goldmark.Extender{
		meta.Meta,
		extension.GFM,
	}
	if m.JsxRender != nil {
		extenders = append(extenders, mdx.NewMdJsx("md"))
	}

	var buf bytes.Buffer
	context := parser.NewContext()
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithXHTML(),
		),
		goldmark.WithExtensions(extenders...),
		goldmark.WithParserOptions(parser.WithASTTransformers(
			util.Prioritized(mdext.NewImageUrlReplace(m.AssetsUrlProcess), 0),
		)),
	)

	if err := md.Convert(src, &buf, parser.WithContext(context)); err != nil {
		return MdResult{}, err
	}

	mt, err := meta.TryGet(context)
	if err != nil {
		return MdResult{}, err
	}
	strMap := ToStrMap(mt).(map[string]interface{})

	return MdResult{
		Body: buf.Bytes(),
		Meta: strMap,
	}, nil
}

// ToStrMap gopkg.in/yaml.v2 会解析出 map[interface{}]interface{} 这样的结构，不支持 json 序列化。需要手动转一次
func ToStrMap(i interface{}) interface{} {
	switch t := i.(type) {
	case map[string]interface{}:
		m := map[string]interface{}{}
		for k, v := range t {
			m[k] = ToStrMap(v)
		}
		return m
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range t {
			m[k.(string)] = ToStrMap(v)
		}
		return m
	case []interface{}:
		m := make([]interface{}, len(t))
		for i, v := range t {
			m[i] = ToStrMap(v)
		}
		return m
	default:
		return i
	}
}
