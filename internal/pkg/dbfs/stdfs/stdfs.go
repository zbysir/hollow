package stdfs

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"io"
	stdFs "io/fs"
	"os"
	"time"
)

// 实现标准 fs

type Fs struct {
	fs pathfs.FileSystem
}

func NewFs(fs pathfs.FileSystem) *Fs {
	return &Fs{fs: fs}
}

var _ stdFs.FS = (*Fs)(nil)

// Open 实现 fs.Fs
func (f *Fs) Open(name string) (stdFs.File, error) {
	if name == "/" {
		d, status := f.fs.OpenDir(name, nil)
		if status.Ok() {
			return DirFile{
				es:   d,
				name: name,
			}, nil
		}
		return nil, os.ErrNotExist
	}

	// 尝试读文件
	nf, status := f.fs.Open(name, 0, nil)
	//log.Debugf("Open %v %v", name, status)
	if status.Ok() {
		return &sFile{
			f:    nf,
			name: name,
		}, nil
	}

	// 尝试读文件夹
	d, status := f.fs.OpenDir(name, nil)
	//log.Debugf("OpenDir %v %v", name, status)
	if status.Ok() {
		return DirFile{
			es:   d,
			name: name,
		}, nil
	}

	return nil, os.ErrNotExist
}

type DirFile struct {
	es   []fuse.DirEntry
	name string
}

type DirStat struct {
	df *DirFile
}

func (s DirStat) Name() string {
	return s.df.name
}

func (s DirStat) Size() int64 {
	return 0
}

func (s DirStat) Mode() stdFs.FileMode {
	return fuse.S_IFDIR | 0755
}

func (s DirStat) ModTime() time.Time {
	return time.Time{}
}

func (s DirStat) IsDir() bool {
	return true
}

func (s DirStat) Sys() any {
	return nil
}

func (d DirFile) Stat() (stdFs.FileInfo, error) {
	return DirStat{&d}, nil
}

func (d DirFile) Read(bytes []byte) (int, error) {
	return 0, nil
}

func (d DirFile) Close() error {
	return nil
}

func (d DirFile) ReadDir(n int) ([]stdFs.DirEntry, error) {
	var xs []stdFs.DirEntry

	for _, e := range d.es {
		xs = append(xs, &DirEntry{e})
	}

	return xs, nil
}

type sFile struct {
	f           nodefs.File
	readoffsite int64
	name        string
}

func (s *sFile) Name() string {
	return s.name
}

func (s *sFile) Size() int64 {
	var a fuse.Attr
	s.f.GetAttr(&a)
	return int64(a.Size)
}

func (s *sFile) Mode() stdFs.FileMode {
	var a fuse.Attr
	s.f.GetAttr(&a)
	return stdFs.FileMode(a.Mode)
}

func (s *sFile) ModTime() time.Time {
	return time.Now()
}

func (s *sFile) IsDir() bool {
	return false
}

func (s *sFile) Sys() any {
	return nil
}

func (s *sFile) Stat() (stdFs.FileInfo, error) {
	return s, nil
}

func (s *sFile) Read(bytes []byte) (int, error) {
	x, _ := s.f.Read(bytes, s.readoffsite)
	if x.Size() == 0 {
		return 0, io.EOF
	}
	s.readoffsite += int64(x.Size())
	return x.Size(), nil
}

func (s *sFile) Close() error {
	return nil
}

type DirFileInfo struct {
	e fuse.DirEntry
}

func (f DirFileInfo) Name() string {
	return f.e.Name
}

func (f DirFileInfo) Size() int64 {
	return 0
}

func (f DirFileInfo) Mode() stdFs.FileMode {
	if f.e.Mode == 0 {
		return fuse.S_IFDIR | 0755
	}
	return stdFs.FileMode(f.e.Mode)
}

func (f DirFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (f DirFileInfo) IsDir() bool {
	return f.e.Mode&fuse.S_IFDIR == fuse.S_IFDIR
}

func (f DirFileInfo) Sys() any {
	return nil
}

type DirEntry struct {
	e fuse.DirEntry
}

func (d *DirEntry) Name() string {
	return d.e.Name
}

func (d *DirEntry) IsDir() bool {
	return d.e.Mode&fuse.S_IFDIR == fuse.S_IFDIR
}

func (d *DirEntry) Type() stdFs.FileMode {
	return stdFs.FileMode(d.e.Mode)
}

func (d DirEntry) Info() (stdFs.FileInfo, error) {
	return DirFileInfo{d.e}, nil
}
