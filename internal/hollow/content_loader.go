package hollow

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	jsx "github.com/zbysir/gojsx"
	"gopkg.in/yaml.v3"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type ContentLoader interface {
	Load(fs fs.FS, filePath string, withContent bool) (Content, error)
}

type MDLoader struct {
	assets Assets
}

type GetContentOpt struct {
	Pure bool `json:"pure"` // 返回纯文本，一般用于做搜索
}

func (m *MDLoader) Load(f fs.FS, filePath string, withContent bool) (Content, error) {
	fileDir, name := filepath.Split(filePath)
	if !strings.HasPrefix(fileDir, "/") {
		fileDir = "/" + fileDir
	}

	ext := filepath.Ext(filePath)

	// 读取 metadata
	body, err := fs.ReadFile(f, filePath)
	if err != nil {
		return Content{}, err
	}

	var meta = map[string]interface{}{}
	if bytes.HasPrefix(body, []byte("---\n")) {
		bbs := bytes.SplitN(body, []byte("---"), 3)
		if len(bbs) > 2 {
			metaByte := bbs[1]
			err = yaml.Unmarshal(metaByte, &meta)
			if err != nil {
				return Content{}, fmt.Errorf("parse file metadata error: %w", err)
			}

			body = bbs[2]
		}
	}

	// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式。在 js 中好处理
	for k, v := range meta {
		switch t := v.(type) {
		case time.Time:
			meta[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
		}
	}

	md := newMdRender(func(p string) string {
		if filepath.IsAbs(p) {
		} else {
			p = filepath.Join(fileDir, p)
		}

		// 移除 assets 文件夹前缀
		for _, a := range m.assets {
			if strings.HasPrefix(p, "/"+a) {
				p = strings.TrimPrefix(p, "/"+a)
				break
			}
		}
		return p
	})

	content := ""
	if withContent {
		content = string(md.Render(body))
	}
	return Content{
		Name: name,
		GetContent: func(opt GetContentOpt) string {
			var s string
			if withContent {
				s = content
			} else {
				s = string(md.Render(body))
			}
			if opt.Pure {
				d, err := goquery.NewDocumentFromReader(strings.NewReader(s))
				if err != nil {
					return err.Error()
				}

				s = d.Text()
			}
			return s
		},
		Meta:    meta,
		Ext:     ext,
		Content: content,
	}, nil
}

type HtmlLoader struct {
	//assets Assets
}

func (m *HtmlLoader) Load(f fs.FS, filePath string, withContent bool) (Content, error) {
	dir, name := filepath.Split(filePath)
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}

	ext := filepath.Ext(filePath)

	// 读取 metadata
	body, err := fs.ReadFile(f, filePath)
	if err != nil {
		return Content{}, err
	}

	var meta = map[string]interface{}{}
	if bytes.HasPrefix(body, []byte("---\n")) {
		bbs := bytes.SplitN(body, []byte("---"), 3)
		if len(bbs) > 2 {
			metaByte := bbs[1]
			err = yaml.Unmarshal(metaByte, &meta)
			if err != nil {
				return Content{}, fmt.Errorf("parse file metadata error: %w", err)
			}

			body = bbs[2]
		}
	}

	// 格式化为 Mon Jan 02 2006 15:04:05 GMT-0700 (MST) 格式。在 js 中好处理
	for k, v := range meta {
		switch t := v.(type) {
		case time.Time:
			meta[k] = t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
		}
	}

	content := ""
	if withContent {
		content = string(body)
	}
	return Content{
		Name: name,
		GetContent: func(opt GetContentOpt) string {
			var s = string(body)
			if opt.Pure {
				d, err := goquery.NewDocumentFromReader(strings.NewReader(s))
				if err != nil {
					return err.Error()
				}

				s = d.Text()
			}

			return s
		},
		Meta:    meta,
		Ext:     ext,
		Content: content,
	}, nil
}

type JsxLoader struct {
	//assets Assets
	x *jsx.Jsx
}

func (m *JsxLoader) Load(f fs.FS, filePath string, withContent bool) (Content, error) {
	body, err := m.x.Render("./"+filePath, nil, jsx.WithRenderFs(f))
	//log.Errorf("filePath %+v %v", body, err)
	if err != nil {
		return Content{}, err
	}

	return Content{
		Name:       "",
		GetContent: nil,
		Meta:       nil,
		Ext:        "",
		Content:    body,
		IsDir:      false,
	}, nil
}
