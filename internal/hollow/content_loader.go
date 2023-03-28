package hollow

import (
	"github.com/PuerkitoBio/goquery"
	jsx "github.com/zbysir/gojsx"
	"io/fs"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ContentLoader interface {
	Load(filePath string, withContent bool) (Content, error)
}

type MDLoader struct {
	assets Assets
	jsx    *jsx.Jsx
	module map[string]map[string]interface{}
}

func NewMDLoader(assets Assets, jsx *jsx.Jsx, module map[string]map[string]interface{}) *MDLoader {
	return &MDLoader{assets: assets, jsx: jsx, module: module}
}

type GetContentOpt struct {
	Pure bool `json:"pure"` // 返回纯文本，一般用于做搜索
}

func processContent(content string, opt GetContentOpt) string {
	if opt.Pure {
		d, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err == nil {
			content = d.Text()
		}
	}
	return content
}

type relativeFs struct {
	sub      fs.FS
	relative string
}

func (r *relativeFs) Open(name string) (fs.File, error) {
	re := filepath.Join(r.relative, name)
	return r.sub.Open(re)
}

func (m *MDLoader) replaceImgUrl(dom jsx.VDom, baseDir string) (as Assets) {
	walkVDom(dom, func(d jsx.VDom) {
		i := d["nodeName"]
		nodeName, _ := i.(string)
		if nodeName == "img" {
			attr := d["attributes"].(map[string]interface{})
			src := attr["src"].(string)

			isOut := strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")
			if isOut {
				return
			}

			if !filepath.IsAbs(src) {
				src = filepath.Join(baseDir, src)
			}

			var inAssets bool
			// 移除 assets 文件夹前缀
			for _, a := range m.assets {
				if strings.HasPrefix(src, "/"+a) {
					src = strings.TrimPrefix(src, "/"+a)
					inAssets = true
					break
				}
			}
			if !inAssets {
				src, _ = url.PathUnescape(src)
				as = append(as, src)
				src = filepath.Join("/__source", src)
			}

			attr["src"] = src
		}
	})

	return
}

// replaceAttrDot 替换 __ 为 .
func replaceAttrDot(dom jsx.VDom) {
	walkVDom(dom, func(d jsx.VDom) {
		attr := d["attributes"].(map[string]interface{})
		for k, v := range attr {
			if strings.Contains(k, "__") {
				delete(attr, k)
				attr[strings.ReplaceAll(k, "__", ".")] = v
			}
		}
	})

	return
}

func (m *MDLoader) Load(filePath string, withContent bool) (Content, error) {
	var os = []jsx.OptionExec{
		jsx.WithAutoExecJsx(nil),
	}

	for p, v := range m.module {
		os = append(os, jsx.WithNativeModule(p, v))
	}
	e, err := m.jsx.Exec("./"+filePath, os...)
	if err != nil {
		return Content{}, err
	}
	fileDir, name := filepath.Split(filePath)
	if !strings.HasPrefix(fileDir, "/") {
		fileDir = "/" + fileDir
	}

	metai := e.Exports["meta"]
	meta, _ := metai.(map[string]interface{})

	for k, v := range meta {
		switch t := v.(type) {
		case time.Time:
			// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式，前端才能处理
			meta[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
		}
	}

	var dom jsx.VDom
	switch t := e.Default.(type) {
	case jsx.VDom:
		dom = t
	default:
		panic(t)
	}
	assets := m.replaceImgUrl(dom, fileDir)
	replaceAttrDot(dom)

	tocItem := generateTOC(dom)
	if err != nil {
		return Content{}, err
	}
	content := dom.Render()

	ext := filepath.Ext(filePath)
	name = strings.TrimSuffix(name, ext)

	return Content{
		Name: name,
		GetContent: func(opt GetContentOpt) string {
			return processContent(content, opt)
		},
		Meta:    meta,
		Ext:     ext,
		Content: content,
		IsDir:   false,
		Toc:     tocItem.Items,
		Assets:  assets,
	}, nil
}

type TocItem struct {
	Title string     `json:"title,omitempty"`
	Items []*TocItem `json:"items,omitempty"`
	Id    string     `json:"id,omitempty"`
	level int
}

func (n *TocItem) AddChild(child *TocItem) {
	n.Items = append(n.Items, child)
}

func (n *TocItem) Dump(deep int) string {
	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", deep*2))
	sb.WriteString(n.Title)
	sb.WriteString("\n")
	for _, c := range n.Items {
		sb.WriteString(c.Dump(deep + 1))
	}
	return sb.String()
}

func lookupMapI(m map[string]interface{}, keys ...string) (interface{}, bool) {
	var c interface{} = m
	for _, k := range keys {
		mm, ok := c.(map[string]interface{})
		if !ok {
			return nil, false
		}

		i, ok := mm[k]
		if !ok {
			return nil, false
		}

		c = i
	}

	return c, true
}

// lookupMap({a: {b: 1}}, "a", "b") => 1
func lookupMap[T any](m map[string]interface{}, keys ...string) (t T, b bool) {
	r, ok := lookupMapI(m, keys...)
	if ok {
		if m, ok := r.(T); ok {
			return m, true
		}
	}
	return t, false
}

func generateTOC(d jsx.VDom) *TocItem {
	root := &TocItem{}
	currentNodes := make([]*TocItem, 1)
	currentNodes[0] = root
	walkVDom(d, func(n jsx.VDom) {
		tag := n["nodeName"].(string)

		var level int
		switch tag {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			leveli, _ := strconv.ParseInt(strings.TrimPrefix(tag, "h"), 10, 64)
			level = int(leveli)
		}
		if level == 0 {
			return
		}

		node := &TocItem{}
		children, _ := lookupMap[interface{}](n, "attributes", "children")
		if children != nil {
			node.Title = strings.TrimSpace(jsx.Render(children))
		}
		id, ok := lookupMap[string](n, "attributes", "id")
		if ok {
			node.Id = id
		}
		node.level = level

		if level >= len(currentNodes) {
			last := currentNodes[len(currentNodes)-1]
			// 当小于才算子级，否则是平级
			if last.level < level {
				parent := last
				parent.AddChild(node)
				currentNodes = append(currentNodes, node)
			} else {
				parent := currentNodes[len(currentNodes)-2]
				parent.AddChild(node)
				currentNodes[len(currentNodes)-1] = node
			}
		} else {
			parent := currentNodes[level-1]
			parent.AddChild(node)
			currentNodes[level] = node
		}

	})

	return root
}

func walkVDom(v interface{}, fun func(d jsx.VDom)) {
	if v == nil {
		return
	}

	var c interface{}
	switch t := v.(type) {
	case map[string]interface{}:
		fun(t)
		c, _ = lookupMapI(t, "attributes", "children")
	case jsx.VDom:
		fun(t)
		c, _ = lookupMapI(t, "attributes", "children")
	case []interface{}:
		for _, i := range t {
			walkVDom(i, fun)
		}
	}

	if c != nil {
		walkVDom(c, fun)
	}
}
