package hollow

import (
	"github.com/PuerkitoBio/goquery"
	jsx "github.com/zbysir/gojsx"
	"io/fs"
	"path/filepath"
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

func (m *MDLoader) replaceImgUrl(dom jsx.VDom, baseDir string) {
	walkVDom(dom, func(d jsx.VDom) {
		i := d["nodeName"]
		nodeName, _ := i.(string)
		if nodeName == "img" {
			attr := d["attributes"].(map[string]interface{})
			src := attr["src"].(string)

			if filepath.IsAbs(src) || strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			} else {
				src = filepath.Join(baseDir, src)
			}

			// 移除 assets 文件夹前缀
			for _, a := range m.assets {
				if strings.HasPrefix(src, "/"+a) {
					src = strings.TrimPrefix(src, "/"+a)
					break
				}
			}

			attr["src"] = src
		}
	})

	return
}

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
	m.replaceImgUrl(dom, fileDir)
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
		Toc:     e.Exports["toc"],
	}, nil
}

func walkVDom(v jsx.VDom, fun func(d jsx.VDom)) {
	if v == nil {
		return
	}
	// 检查所有 img，如果有相对路径，则替换
	fun(v)
	attr := v["attributes"]
	if attr != nil {
		attrMap := attr.(map[string]interface{})
		children := attrMap["children"]
		switch t := children.(type) {
		case []interface{}:
			for _, i := range t {
				switch t := i.(type) {
				case map[string]interface{}:
					walkVDom(t, fun)
				}
			}
		case map[string]interface{}:
			walkVDom(t, fun)

		}
	}

}
