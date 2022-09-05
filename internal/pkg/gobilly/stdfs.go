package gobilly

import (
	"github.com/go-git/go-billy/v5"
	"io"
	"io/fs"
	"os"
)

// StdFs implement io/fs
type StdFs struct {
	under billy.Filesystem
}

var _ fs.FS = (*StdFs)(nil)
var _ io.Seeker = (*stdFile)(nil)

func NewStdFs(under billy.Filesystem) *StdFs {
	return &StdFs{under: under}
}

func (s *StdFs) Open(name string) (fs.File, error) {
	f, err := s.under.Open(name)
	if err != nil {
		return nil, err
	}
	return &stdFile{
		f:     f,
		path:  name,
		under: s.under,
	}, nil
}

var _ fs.FS = (*StdFs)(nil)
var _ fs.ReadDirFile = (*stdFile)(nil)

type stdFile struct {
	f     billy.File
	path  string
	under billy.Filesystem
}

func (s *stdFile) Seek(offset int64, whence int) (int64, error) {
	return s.f.Seek(offset, whence)
}

type dirEntry struct {
	f os.FileInfo
}

func (d *dirEntry) Name() string {
	return d.f.Name()
}

func (d *dirEntry) IsDir() bool {
	return d.f.IsDir()
}

func (d *dirEntry) Type() fs.FileMode {
	return d.f.Mode()
}

func (d *dirEntry) Info() (fs.FileInfo, error) {
	return d.f, nil
}

func (s *stdFile) ReadDir(n int) ([]fs.DirEntry, error) {
	ds, err := s.under.ReadDir(s.path)
	if err != nil {
		return nil, err
	}
	de := make([]fs.DirEntry, len(ds))
	for i, d := range ds {
		de[i] = &dirEntry{d}
	}
	return de, nil
}

func (s *stdFile) Stat() (fs.FileInfo, error) {
	return s.under.Stat(s.path)
}

func (s *stdFile) Read(bytes []byte) (int, error) {
	return s.f.Read(bytes)
}

func (s *stdFile) Close() error {
	return s.f.Close()
}

var _ fs.File = (*stdFile)(nil)
