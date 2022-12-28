package mdext

import (
	"bytes"
	"fmt"
	"github.com/tdewolff/parse/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	jsx2 "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/htmlparser"
	"io"
	"io/fs"
	"regexp"
)

type jsx struct {
	x  *jsx2.Jsx
	fs fs.FS
}

func NewJsx(x *jsx2.Jsx, fs fs.FS) *jsx {
	return &jsx{x: x, fs: fs}
}

func (e *jsx) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&jsxParser{}, 0),
			util.Prioritized(NewJsCodeParser(), 0),
		),
		//parser.WithInlineParsers(util.Prioritized(&jsxParser{}, 0)),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&jsxRender{x: e.x, fs: e.fs}, 0),
	))
}

var jsxKind = ast.NewNodeKind("Jsx")

type jsxNode struct {
	ast.BaseBlock
	pc  parser.Context
	tag string
}

func (j *jsxNode) Kind() ast.NodeKind {
	return jsxKind
}

// IsRaw return true 不解析 block 中的内容
func (j *jsxNode) IsRaw() bool {
	return true
}

func (j *jsxNode) Dump(source []byte, level int) {
	ast.DumpHelper(j, source, level, nil, nil)
}

func (j *jsxNode) HasBlankPreviousLines() bool {
	return true
}

func (j *jsxNode) SetBlankPreviousLines(v bool) {
	return
}

type jsxParser struct {
}

// InlineParser 暂时不实现
//func (j *jsxParser) Parse(parent ast.Node, reader text.Reader, pc parser.Context) ast.Node {
//	panic("implement me")
//}

var _ parser.BlockParser = (*jsxParser)(nil)

//var _ parser.InlineParser = (*jsxParser)(nil) // InlineParser 暂时不实现

func (j *jsxParser) Trigger() []byte {
	return []byte{'<'}
}

var htmlTagReg = regexp.MustCompile(`^[ ]{0,3}<(/[ ]*)?([a-zA-Z]+[a-zA-Z0-9\-]*)`)

func (j *jsxParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	node := &jsxNode{
		BaseBlock: ast.BaseBlock{},
		pc:        pc,
	}
	line, segment := reader.PeekLine()
	if pos := pc.BlockOffset(); pos < 0 || line[pos] != '<' {
		return nil, parser.NoChildren
	}
	match := htmlTagReg.FindAllSubmatch(line, -1)
	if match == nil {
		return nil, parser.NoChildren
	}

	tagName := match[0][2]
	// 大写首字母
	if !(tagName[0] >= 'A' && tagName[0] <= 'Z') {
		return nil, parser.NoChildren
	}

	_, s := reader.Position()
	offset := s.Start
	bs := reader.Source()[offset:]

	buf := bytes.NewBufferString(string(bs))
	start, end, ok, err := ParseToClose(buf)
	if err != nil {
		return nil, parser.NoChildren
	}
	if !ok {
		return nil, parser.NoChildren
	}

	node.tag = string(tagName)
	code := GetJsCode(pc)

	// 简单判断 变量是否存在于 code，如果存在则说明是 JsxElement
	tr := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, tagName))
	if !tr.MatchString(code) {
		return nil, parser.NoChildren
	}

	segment = text.NewSegment(start+offset, end+offset)
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return node, parser.Close

}

func (j *jsxParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (j *jsxParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	//node.RemoveChildren(node)
}

func (j *jsxParser) CanInterruptParagraph() bool {
	return true
}

func (j *jsxParser) CanAcceptIndentedLine() bool {
	return true
}

type jsxRender struct {
	x  *jsx2.Jsx
	fs fs.FS
}

func (j *jsxRender) Render(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	lines := node.Lines()
	var b bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		b.Write(line.Value(src))
	}

	jsxNode := node.(*jsxNode)
	code := GetJsCode(jsxNode.pc)
	code += "\n"
	code += b.String()

	v, err := j.x.RunJs([]byte(code), jsx2.WithTransform(true), jsx2.WithRunFileName("root.tsx"), jsx2.WithRunFs(j.fs))
	if err != nil {
		return ast.WalkContinue, err
	}
	vd := jsx2.VDom(v.Export().(map[string]interface{}))

	w.Write([]byte(vd.Render()))

	return ast.WalkContinue, nil
}

func (j *jsxRender) RegisterFuncs(registerer renderer.NodeRendererFuncRegisterer) {
	registerer.Register(jsxKind, j.Render)
}

func ParseToClose(buf *bytes.Buffer) (start, end int, ok bool, err error) {
	input := parse.NewInput(buf)

	l := htmlparser.NewLexer(input)

	nesting := 0
	var currTag []byte
	var matchTag []byte
	pos := 0

	for i := 0; i < 20 && end == 0; i++ {
		err := l.Err()
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, 0, false, err
		}

		tp, bs := l.Next()

		//log.Infof("%s %s", tp, bs)

		begin := pos
		pos += len(bs)
		switch tp {
		case htmlparser.StartTagToken:
			currTag = bs[1:]
			if len(matchTag) == 0 {
				matchTag = bs[1:]
				nesting += 1
				start = begin
			} else if bytes.Equal(matchTag, bs[1:]) {
				nesting += 1
			}
		case htmlparser.StartTagVoidToken:
			if bytes.Equal(matchTag, currTag) {
				nesting -= 1
				if nesting == 0 {
					end = pos
					break
				}
			}
		case htmlparser.EndTagToken:
			if bytes.Equal(matchTag, bs[2:len(bs)-1]) {
				nesting -= 1
				if nesting == 0 {
					end = pos
					break
				}
			}
		}
	}
	if end != 0 {
		return start, end, true, nil
	}

	return
}
