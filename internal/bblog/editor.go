// 使用 Gin 暴露 http api

package bblog

import (
	"context"
	"io/fs"
	"path"
)

type Editor struct {
}

func (a *Editor) Run(ctx context.Context) (err error) {
	return nil
}

type Handler struct {
}

type FileTree struct {
	Name  string     `json:"name"`
	Path  string     `json:"path"` // full path
	IsDir bool       `json:"is_dir"`
	Items []FileTree `json:"items"`
}

type FsApi struct {
	fs fs.FS
}

func (f *FsApi) FileTree(base string, deep int) (ft FileTree, err error) {
	_, ft.Name = path.Split(base)
	ft.Path = base
	ft.IsDir = true
	if deep == 0 {
		return
	}

	fds, err := fs.ReadDir(f.fs, base)
	if err != nil {
		return ft, err
	}

	for _, fd := range fds {
		srcfp := path.Join(base, fd.Name())

		if fd.IsDir() {
			ftw, err := f.FileTree(srcfp, deep-1)
			if err != nil {
				return ft, err
			}
			ft.Items = append(ft.Items, ftw)
		} else {
			ft.Items = append(ft.Items, FileTree{
				Name:  fd.Name(),
				Path:  srcfp,
				IsDir: false,
				Items: nil,
			})
		}
	}

	return ft, nil
}
