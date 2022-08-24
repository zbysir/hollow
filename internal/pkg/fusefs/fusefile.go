package fusefs

import (
	"bytes"
	"encoding/gob"
	"github.com/docker/libkv/store"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type file struct {
	kvStore    store.Store
	kv         *store.KVPair
	name       string
	content    []byte
	attrx      []byte
	createMode uint32
	nodefs.File
}

// TODO 优化为懒加载，只在读的时候才去 store 读
func newFile(s store.Store, kv *store.KVPair, createMode uint32) nodefs.File {
	logrus.Debugf("newFile: %s", kv.Value)

	value := kv.Value

	var attrx []byte
	var content = value

	// 第一行是元数据
	sp := bytes.SplitAfterN(value, []byte(`\n\n`), 1)
	if len(sp) == 2 {
		attrx = sp[0]
		content = sp[1]
	}

	return &file{
		kvStore:    s,
		kv:         kv,
		name:       kv.Key,
		content:    content,
		attrx:      attrx,
		File:       nodefs.NewDefaultFile(),
		createMode: createMode,
	}
}

func (f *file) String() string {
	return f.kv.Key
}

type Attr struct {
	MTime uint64
	CTime uint64
	Size  uint64
	Mode  uint32
}

func (a *Attr) toByte() []byte {
	return encode(a)
}

func (a *Attr) fromByte(bs []byte) {
	decode(bs, a)
}

func encode(data interface{}) []byte {
	//Buffer类型实现了io.Writer接口
	var buf bytes.Buffer
	//得到编码器
	enc := gob.NewEncoder(&buf)
	//调用编码器的Encode方法来编码数据data
	enc.Encode(data)
	//编码后的结果放在buf中
	return buf.Bytes()
}

func decode(data []byte, r interface{}) {
	buf := bytes.NewReader(data)
	//获取一个解码器，参数需要实现io.Reader接口
	dec := gob.NewDecoder(buf)
	//调用解码器的Decode方法将数据解码，用Q类型的q来接收
	dec.Decode(r)
	return
}

func (f *file) readAttr() (a *fuse.Attr, err error) {
	a = &fuse.Attr{}
	var at Attr
	at.fromByte(f.attrx)

	a.Mode = at.Mode
	a.Mtime = at.MTime
	a.Ctime = at.CTime
	a.Size = at.Size

	return a, nil

}

func (f *file) writeAttr(a *fuse.Attr) (err error) {
	var at = Attr{
		MTime: a.Mtime,
		CTime: a.Ctime,
		Size:  a.Size,
		Mode:  a.Mode,
	}
	f.attrx = at.toByte()
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
	if a.Mode == 0 {
		a.Mode = f.createMode
	}

	if a.Ctime == 0 {
		a.Ctime = uint64(time.Now().Unix())
	}
	a.Mtime = uint64(time.Now().Unix())

	body := f.kv.Value[:off]
	body = append(body, data...)
	a.Size = uint64(len(body))

	err = f.writeAttr(a)
	if err != nil {
		return 0, fuse.EIO
	}

	fileAll := append(f.attrx, '\n', '\n')
	fileAll = append(fileAll, body...)
	if err := f.kvStore.Put(f.kv.Key, body, nil); err != nil {
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
