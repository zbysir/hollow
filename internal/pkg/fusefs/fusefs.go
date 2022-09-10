package fusefs

import (
	"fmt"
	"github.com/docker/libkv/store/boltdb"
	"github.com/zbysir/hollow/internal/pkg/log"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/docker/libkv/store"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/sirupsen/logrus"
)

func init() {
	boltdb.Register()
}

type FuseFS struct {
	pathfs.FileSystem
	kvStore store.Store
	root    string
}

var _ pathfs.FileSystem = (*FuseFS)(nil)

type Options struct {
	Root  string
	Store store.Store
}

func NewFuseFs(opts Options) (*FuseFS, error) {
	root := opts.Root
	if root != "" && root[0] == '/' {
		root = root[1:]
	}

	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}

	log.Debugf("root: %v", root)

	f := &FuseFS{
		FileSystem: pathfs.NewDefaultFileSystem(),
		kvStore:    opts.Store,
		root:       root,
	}

	f.Mkdir(root, 0, nil)
	return f, nil
}

func (fs *FuseFS) Create(name string, _ uint32, mode uint32, _ *fuse.Context) (nodefs.File, fuse.Status) {
	if name == "" {
		return nil, fuse.EACCES
	}
	name = path.Join(fs.root, name)
	logrus.Debugf("Create: %v", name)
	nf := newFile(fs.kvStore, &store.KVPair{
		Key: name,
	}, mode)
	_, status := nf.Write([]byte{}, 0)
	return nf, status
}

func (fs *FuseFS) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	name = path.Join(fs.root, name)
	kv, err := fs.kvStore.Get(name)
	logrus.Debugf("Open: %v", name)
	if err != nil {
		return nil, fuse.ENOENT
	}
	return newFile(fs.kvStore, kv, 0), fuse.OK
}

func (fs *FuseFS) OpenDir(path string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	path = filepath.Join(fs.root, path)
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	logrus.Debugf("OpenDir: %v", path)
	_, err := fs.kvStore.Get(path)
	if err != nil {
		return nil, fuse.ENOENT
	}

	kvs, err := fs.kvStore.List(path)
	if err != nil {
		logrus.Error(err)
		return nil, fuse.ENOENT
	}

	var entries []fuse.DirEntry
	for _, kv := range kvs {
		fullPath := kv.Key

		dir, fi := filepath.Split(fullPath)
		if dir == path && fi == "" {
			logrus.Debugf("skipping base %s", fullPath)
			continue
		}

		rmDir := strings.TrimPrefix(fullPath, path)
		rmDir = strings.TrimPrefix(rmDir, "/")
		// 不允许有 /，除了只允许最后一个
		// /src
		// /src/jsx ok
		// /src/js/ ok
		// /src/js/x X
		// src/js/
		index := strings.Index(rmDir, "/")
		if index != -1 && index != len(rmDir)-1 {
			logrus.Debugf("skipping subtree %s %s", dir, fullPath)
			continue
		}

		var mode uint32 = fuse.S_IFREG
		if strings.HasSuffix(rmDir, "/") {
			fullPath = strings.TrimPrefix(rmDir, "/")
			mode = fuse.S_IFDIR
		} else {
			fullPath = fi
		}

		entry := fuse.DirEntry{Name: fullPath, Mode: mode}
		entries = append(entries, entry)
	}

	// 排序 文件夹在前
	sort.Slice(entries, func(i, j int) bool {
		iIsDir := entries[i].Mode&fuse.S_IFDIR == fuse.S_IFDIR
		jIsDir := entries[j].Mode&fuse.S_IFDIR == fuse.S_IFDIR
		if iIsDir && !jIsDir {
			return true
		} else if !iIsDir && jIsDir {
			return false
		}

		return entries[i].Name < entries[j].Name
	})
	return entries, fuse.OK
}

// List 用于方便的调试
func (fs *FuseFS) List(path string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	path = filepath.Join(fs.root, path)
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	logrus.Debugf("OpenDir: %v", path)
	kvs, err := fs.kvStore.List(path)
	if err != nil {
		logrus.Error(err)
		return nil, fuse.ENOENT
	}

	var entries []fuse.DirEntry
	for _, kv := range kvs {
		log.Debugf("kv %v", kv.Key)
		if kv.Key == "" {
			continue
		}
		fullPath := kv.Key
		var mode uint32 = fuse.S_IFREG
		if strings.HasSuffix(fullPath, "/") {
			//fullPath = strings.TrimPrefix(fullPath, "/")
			mode = fuse.S_IFDIR
		}

		entry := fuse.DirEntry{Name: fullPath, Mode: mode}
		entries = append(entries, entry)
	}
	return entries, fuse.OK
}

func (fs *FuseFS) StatFs(name string) *fuse.StatfsOut {
	name = path.Join(fs.root, name)
	logrus.Debugf("StatFs: %s", name)
	kvs, err := fs.kvStore.List(name)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return &fuse.StatfsOut{Files: uint64(len(kvs))}
}

// TODO 完成 GetAttr
func (fs *FuseFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	name = path.Join(fs.root, name)
	logrus.Debugf("GetAttr: %s", name)
	now := time.Now()
	attr := &fuse.Attr{
		Mtime:     uint64(now.Unix()),
		Mtimensec: uint32(now.UnixNano()),
		Atime:     uint64(now.Unix()),
		Atimensec: uint32(now.UnixNano()),
		Ctime:     uint64(now.Unix()),
		Ctimensec: uint32(now.UnixNano()),
		Mode:      fuse.S_IFDIR | 0755, // default to dir
	}

	if name == "" {
		return attr, fuse.OK
	}

	kv, err := fs.kvStore.Get(name)
	if err != nil {
		if err == store.ErrKeyNotFound {
			// check if this is a dir, ie the key name might have a trailing "/"
			kv, err = fs.kvStore.Get(name + "/")
		}
		if err != nil {
			return nil, fuse.ENOENT
		}
	}

	if !strings.HasSuffix(kv.Key, "/") {
		attr.Mode = fuse.S_IFREG | 0644
		attr.Size = uint64(len(kv.Value))
	}
	return attr, fuse.OK
}

func (fs *FuseFS) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	if name == "" {
		return fuse.ENOENT
	}
	if name == "/" {
		name = ""
	}
	name = path.Join(fs.root, name)
	if !strings.HasSuffix(name, "/") {
		name += "/"
	}
	logrus.Debugf("Mkdir: %s", name)
	if err := fs.kvStore.Put(name, nil, nil); err != nil {
		logrus.Error(err)
		return fuse.ENOENT
	}
	return fuse.OK
}

func (fs *FuseFS) Rename(oldName string, newName string, context *fuse.Context) fuse.Status {
	oldName = path.Join(fs.root, oldName)
	newName = path.Join(fs.root, newName)
	logrus.Debugf("Rename: %s -> %s", oldName, newName)
	kv, err := fs.kvStore.Get(oldName)
	if err != nil {
		logrus.Error(err)
		return fuse.EIO
	}

	var undo func()
	newKv, err := fs.kvStore.Get(newName)
	if err == nil {
		undo = func() {
			fs.kvStore.Put(oldName, kv.Value, nil)
			fs.kvStore.Put(newKv.Key, newKv.Value, nil)
		}
	} else {
		undo = func() {
			fs.kvStore.Put(oldName, kv.Value, nil)
			fs.kvStore.Delete(newName)
		}
	}

	if err := fs.kvStore.Put(newName, kv.Value, nil); err != nil {
		logrus.Error(err)
		return fuse.EIO
	}

	if err := fs.kvStore.Delete(oldName); err != nil {
		undo()
		logrus.Error(err)
		return fuse.EIO
	}

	return fuse.OK
}

func (fs *FuseFS) Rmdir(name string, context *fuse.Context) fuse.Status {
	name = path.Join(fs.root, name)
	logrus.Debugf("Rmdir: %s", name)
	if err := fs.kvStore.DeleteTree(name); err != nil {
		logrus.Error(err)
		return fuse.EIO
	}
	return fuse.OK
}

func (fs *FuseFS) Unlink(name string, context *fuse.Context) fuse.Status {
	//if name == "" {
	//	return fuse.ENOENT
	//}
	name = path.Join(fs.root, name)
	logrus.Debugf("Unlink: %s", name)

	if err := fs.kvStore.Delete(name); err != nil {
		logrus.Error(err)
		return fuse.EIO
	}
	return fuse.OK
}

func (fs *FuseFS) KvDelete(name string) error {
	if err := fs.kvStore.Delete(name); err != nil {
		return err
	}
	return nil
}

func (fs *FuseFS) Truncate(name string, size uint64, context *fuse.Context) fuse.Status {
	name = path.Join(fs.root, name)
	kv, err := fs.kvStore.Get(name)
	if err != nil {
		return fuse.EIO
	}

	if size > uint64(len(kv.Value)) {
		size = uint64(len(kv.Value))
	}

	if err := fs.kvStore.Put(name, kv.Value[:size], nil); err != nil {
		return fuse.EIO
	}
	return fuse.OK
}

func (fs *FuseFS) String() string {
	return "kvfs"
}

func (fs *FuseFS) NewServer(mountPoint string) (*fuse.Server, error) {
	conn := nodefs.NewFileSystemConnector(pathfs.NewPathNodeFs(fs, nil).Root(), nil)
	return fuse.NewServer(conn.RawFS(), mountPoint, &fuse.MountOptions{
		Name: fs.String(),
	})
}

type IOReader struct {
	f    nodefs.File
	curr int
}

func NewIOReader(f nodefs.File) *IOReader {
	return &IOReader{f: f}
}

func (r *IOReader) Read(p []byte) (n int, err error) {
	rr, status := r.f.Read(p, int64(r.curr))
	if !status.Ok() {
		return 0, fmt.Errorf("%v", status.String())
	}
	rSize := rr.Size()
	if rSize == 0 || rSize < len(p) {
		return rSize, io.EOF
	}
	r.curr += rSize
	return rSize, nil
}
