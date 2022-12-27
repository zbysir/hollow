package hollow

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	jsx "github.com/zbysir/gojsx"
	md2 "github.com/zbysir/hollow/internal/pkg/md"
	"io/fs"
)

type mdRenderGold struct {
	GoldMdRenderOptions
}

type GoldMdRenderOptions struct {
	jsx              *jsx.Jsx
	fs               fs.FS
	assetsUrlProcess func(string) string
}

func NewGoldMdRender(o GoldMdRenderOptions) *mdRenderGold {
	return &mdRenderGold{GoldMdRenderOptions: o}
}

type MdResult struct {
	Body []byte
	Meta map[string]interface{}
}

func (m *mdRenderGold) Render(src []byte) (MdResult, error) {
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
		goldmark.WithParserOptions(parser.WithASTTransformers(
			util.Prioritized(md2.NewImageUrlReplace(m.assetsUrlProcess), 0),
		)),
	)

	if err := md.Convert(src, &buf, parser.WithContext(context)); err != nil {
		return MdResult{}, err
	}

	meta, err := md2.TryGet(context)
	if err != nil {
		return MdResult{}, err
	}
	strMap := ToStrMap(meta).(map[string]interface{})

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
