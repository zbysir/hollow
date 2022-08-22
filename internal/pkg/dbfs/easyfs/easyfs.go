package easyfs

import (
	"errors"
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/zbysir/blog/internal/pkg/dbfs"
	"io/ioutil"
	"path"
	"path/filepath"
)

// 提供最简单的文件操作 API

type File struct {
	Name      string `json:"name"`
	Path      string `json:"path"` // full path
	DirPath   string `json:"dir_path"`
	IsDir     bool   `json:"is_dir"`
	CreatedAt int64  `json:"created_at"`
	ModifyAt  int64  `json:"modify_at"`
	Body      string `json:"body"`
}

type FileTree struct {
	File
	Items []FileTree `json:"items"`
}

type Fs struct {
	fs pathfs.FileSystem
}

func NewFs(fs pathfs.FileSystem) *Fs {
	return &Fs{
		fs: fs,
	}
}

func (f *Fs) Mkdir(name string) (err error) {
	status := f.fs.Mkdir(name, 0, nil)
	if !status.Ok() {
		return statusError(status)
	}
	return
}

func (f *Fs) RmDir(name string) (err error) {
	status := f.fs.Rmdir(name, nil)
	if !status.Ok() {
		return statusError(status)
	}
	return
}

func (f *Fs) RmFile(name string) (err error) {
	status := f.fs.Unlink(name, nil)
	if !status.Ok() {
		return statusError(status)
	}
	return
}

func (f *Fs) GetFile(path string) (fi *File, err error) {
	nf, status := f.fs.Open(path, 0, nil)
	if !status.Ok() {
		return nil, statusError(status)
	}

	var a fuse.Attr
	nf.GetAttr(&a)

	bs, err := ioutil.ReadAll(dbfs.NewIOReader(nf))
	if err != nil {
		return
	}

	dir, name := filepath.Split(path)
	return &File{
		Name:      name,
		Path:      path,
		DirPath:   dir,
		IsDir:     a.Mode&fuse.S_IFDIR == fuse.S_IFDIR,
		CreatedAt: int64(a.Ctime),
		ModifyAt:  int64(a.Mtime),
		Body:      string(bs),
	}, nil
}

// WriteFile 写文件，如果文件存在则会被覆盖
func (f *Fs) WriteFile(path string, content string) (err error) {
	dir, filename := filepath.Split(path)
	if filename == "" {
		return errors.New("filename can't be empty")
	}

	if dir != "" {
		status := f.fs.Mkdir(dir, 0, nil)
		if !status.Ok() {
			return statusError(status)
		}
	}

	nf, status := f.fs.Create(path, 0, 0, nil)
	if !status.Ok() {
		return statusError(status)
	}
	_, status = nf.Write([]byte(content), 0)
	if !status.Ok() {
		return statusError(status)
	}

	return
}

func (f *Fs) FileTree(base string, deep int) (ft FileTree, err error) {
	_, ft.Name = path.Split(base)
	ft.Path = base
	ft.IsDir = true
	ft.DirPath = base
	if deep == 0 {
		return
	}

	fds, status := f.fs.OpenDir(base, nil)
	if !status.Ok() {
		return ft, statusError(status, base)
	}
	for _, fd := range fds {
		srcfp := path.Join(base, fd.Name)

		if fd.Mode&fuse.S_IFDIR == fuse.S_IFDIR {
			ftw, err := f.FileTree(srcfp, deep-1)
			if err != nil {
				return ft, err
			}
			ft.Items = append(ft.Items, ftw)
		} else {
			ft.Items = append(ft.Items, FileTree{
				File: File{
					Name:      fd.Name,
					Path:      srcfp,
					DirPath:   base,
					IsDir:     false,
					CreatedAt: 0,
					ModifyAt:  0,
					Body:      "",
				},
				Items: nil,
			})
		}
	}

	return ft, nil
}

func statusError(status fuse.Status, msg ...string) error {
	if status.Ok() {
		return nil
	}
	return fmt.Errorf("%s %s", status.String(), msg)
}
