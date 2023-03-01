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

	tocItem, err := toc(dom)
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
		Toc:     tocItem,
		Assets:  assets,
	}, nil
}

type TocItem struct {
	Title string     `json:"title"`
	Items []*TocItem `json:"items"`
	Id    string     `json:"id"`
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

func toc(d jsx.VDom) ([]*TocItem, error) {
	appendChild := func(n *TocItem) *TocItem {
		child := new(TocItem)
		n.Items = append(n.Items, child)
		return child
	}

	lastChild := func(n *TocItem) *TocItem {
		if len(n.Items) > 0 {
			return n.Items[len(n.Items)-1]
		}
		return appendChild(n)
	}

	var root TocItem

	stack := []*TocItem{&root}

	walkVDom(d, func(n jsx.VDom) {
		tag := n["nodeName"].(string)

		var level int64
		switch tag {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			level, _ = strconv.ParseInt(strings.TrimPrefix(tag, "h"), 10, 64)
		}
		if level == 0 {
			return
		}

		for len(stack) < int(level) {
			parent := stack[len(stack)-1]
			stack = append(stack, lastChild(parent))
		}

		for len(stack) > int(level) {
			stack = stack[:len(stack)-1]
		}

		parent := stack[len(stack)-1]
		target := lastChild(parent)
		if len(target.Title) > 0 || len(target.Items) > 0 {
			target = appendChild(parent)
		}

		c, _ := lookupMap[interface{}](n, "attributes", "children")
		if c != nil {
			r := jsx.Render(c)
			target.Title = strings.TrimSpace(r)
		}
		id, ok := lookupMap[string](n, "attributes", "id")
		if ok {
			target.Id = id
		}
	})

	return root.Items, nil
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
