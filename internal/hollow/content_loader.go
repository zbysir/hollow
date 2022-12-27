package hollow

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	jsx "github.com/zbysir/gojsx"
	"io/fs"
	"path/filepath"
	"strings"
)

type ContentLoader interface {
	Load(fs fs.FS, filePath string, withContent bool) (Content, error)
}

type MDLoader struct {
	assets Assets
	jsx    *jsx.Jsx
}

func NewMDLoader(assets Assets, jsx *jsx.Jsx) *MDLoader {
	return &MDLoader{assets: assets, jsx: jsx}
}

type GetContentOpt struct {
	Pure bool `json:"pure"` // 返回纯文本，一般用于做搜索
}

type relativeFs struct {
	sub      fs.FS
	relative string
}

func RelativeFs(fs fs.FS, relative string) fs.FS {
	return &relativeFs{
		sub:      fs,
		relative: relative,
	}
}

func (r *relativeFs) Open(name string) (fs.File, error) {
	re := filepath.Join(r.relative, name)
	return r.sub.Open(re)
}
func trapBOM(fileBytes []byte) []byte {
	trimmedBytes := bytes.Trim(fileBytes, "\xef\xbb\xbf")
	return trimmedBytes
}

func (m *MDLoader) Load(f fs.FS, filePath string, withContent bool) (Content, error) {
	fileDir, name := filepath.Split(filePath)
	if !strings.HasPrefix(fileDir, "/") {
		fileDir = "/" + fileDir
	}

	ext := filepath.Ext(filePath)

	body, err := fs.ReadFile(f, filePath)
	if err != nil {
		return Content{}, err
	}
	body = trapBOM(body)

	mdRender := NewGoldMdRender(GoldMdRenderOptions{
		jsx: m.jsx,
		fs:  RelativeFs(f, fileDir),
		assetsUrlProcess: func(p string) string {
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
		},
	})

	mdr, err := mdRender.Render(body)
	if err != nil {
		return Content{}, err
	}

	content := ""
	if withContent {
		content = string(mdr.Body)
	}

	return Content{
		Name: name,
		GetContent: func(opt GetContentOpt) string {
			s := string(mdr.Body)
			if opt.Pure {
				d, err := goquery.NewDocumentFromReader(strings.NewReader(s))
				if err != nil {
					return err.Error()
				}

				s = d.Text()
			}
			return s
		},
		Meta:    mdr.Meta,
		Ext:     ext,
		Content: content,
	}, nil
}
