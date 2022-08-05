package fs

import (
	"encoding/binary"
	"github.com/docker/libkv/store"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type file struct {
	kvStore store.Store
	kv      *store.KVPair
	name    string
	content []byte
	attr    [17]byte
	nodefs.File
	//a        *fuse.Attr
	//attrOnce sync.Once
}

// TODO 优化为懒加载，只在读的时候才去 store 读
func newFile(s store.Store, kv *store.KVPair) nodefs.File {
	logrus.Debugf("newFile: %s", kv.Value)
	var attr [17]byte
	if len(kv.Value) >= 17 {
		copy(attr[:], kv.Value)
		kv.Value = kv.Value[17:]
	}

	return &file{
		kvStore: s,
		kv:      kv,
		name:    kv.Key,
		content: kv.Value,
		attr:    attr,
		File:    nodefs.NewDefaultFile(),
	}
}

func (f *file) String() string {
	return f.kv.Key
}

func (f *file) readAttr() (a *fuse.Attr, err error) {
	return &fuse.Attr{
		Ino:         0,
		Size:        0,
		Blocks:      0,
		Atime:       0,
		Mtime:       binary.BigEndian.Uint64(f.attr[9:]),
		Ctime:       binary.BigEndian.Uint64(f.attr[1:]),
		Crtime_:     0,
		Atimensec:   0,
		Mtimensec:   0,
		Ctimensec:   0,
		Crtimensec_: 0,
		Mode:        0,
		Nlink:       0,
		Owner:       fuse.Owner{},
		Rdev:        0,
		Flags_:      0,
	}, nil

}

func (f *file) writeAttr(a *fuse.Attr) (err error) {
	binary.BigEndian.PutUint64(f.attr[1:], a.Ctime)
	binary.BigEndian.PutUint64(f.attr[9:], a.Mtime)

	return
}

func (f *file) Read(buf []byte, offset int64) (fuse.ReadResult, fuse.Status) {
	logrus.Debugf("Read: %s", string(f.kv.Value))

	// skip attr

	end := int(offset) + len(buf)
	if end > len(f.kv.Value) {
		end = len(f.kv.Value)
	}

	copy(buf, f.kv.Value[offset:end])
	return fuse.ReadResultData(f.kv.Value[offset:end]), fuse.OK
}

func (f *file) Write(data []byte, off int64) (uint32, fuse.Status) {
	a, err := f.readAttr()
	if err != nil {
		return 0, fuse.EIO
	}
	logrus.Debugf("Write GetAttr: %+v %s", a, f.attr)
	if a.Ctime == 0 {
		a.Ctime = uint64(time.Now().Unix())
	}
	a.Mtime = uint64(time.Now().Unix())
	err = f.writeAttr(a)
	if err != nil {
		return 0, fuse.EIO
	}

	f.kv.Value = f.kv.Value[:off]
	f.kv.Value = append(f.kv.Value, data...)
	//copy(val[off:], data)

	if err := f.kvStore.Put(f.kv.Key, append(f.attr[:], f.kv.Value...), nil); err != nil {
		logrus.Error(err)
		return uint32(0), fuse.EIO
	}
	return uint32(len(data)), fuse.OK
}

func (f *file) GetAttr(out *fuse.Attr) fuse.Status {
	a, err := f.readAttr()
	if err != nil {
		return fuse.EIO
	}
	logrus.Debugf("GetAttr: %+v %s", a, f.attr)
	out.Size = uint64(len(f.kv.Value))
	out.Ctime = a.Ctime
	out.Mtime = a.Mtime

	if f.kv == nil || strings.HasSuffix(f.name, "/") {
		out.Mode = fuse.S_IFDIR | 0755
		return fuse.OK
	}

	if len(f.kv.Value) > 0 {
		out.Mode = fuse.S_IFREG | 0644
		out.Size = uint64(len(f.kv.Value))
	}
	return fuse.OK
}
