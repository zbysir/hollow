package qiniu

import (
	"context"
	"go.uber.org/zap"
	"io"
	"io/fs"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type Qiniu struct {
	mac *qbox.Mac
}

func NewQiniu(access, secret string) *Qiniu {
	if access == "" {
		panic("access can't be empty")
	}
	if secret == "" {
		panic("secret can't be empty")
	}

	return &Qiniu{
		mac: qbox.NewMac(access, secret),
	}
}

func (q Qiniu) Uploader() *Uploader {
	return &Uploader{
		mac: q.mac,
	}
}

var zoneMap = map[string]*storage.Zone{
	"huadong":  &storage.ZoneHuadong,
	"huabei":   &storage.ZoneHuabei,
	"huanan":   &storage.ZoneHuanan,
	"beimei":   &storage.ZoneBeimei,
	"xinjiapo": &storage.ZoneXinjiapo,
}

func GetZoneByName(name string) (zone *storage.Zone, exist bool) {
	zone, exist = zoneMap[name]
	return
}

type Uploader struct {
	mac *qbox.Mac
}

// 自定义返回值结构体
type PutRsp struct {
	Key    string
	Hash   string
	Fsize  int
	Bucket string
}

// UploadFs 上传一个 fs
func (u Uploader) UploadFs(log *zap.SugaredLogger, bucket string, keyPrefix string, f fs.FS) (err error) {
	zone, err := storage.GetZone(u.mac.AccessKey, bucket)
	if err != nil {
		return err
	}

	err = fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if log != nil {
			log.Infof("uploading %v", path)
		}
		//path = path

		_, err = u.UploadFile(zone, bucket, keyPrefix+path, path)
		return err

	})
	if err != nil {
		return
	}
	return
}

// 服务端表单直传 + 自定义回 JSON
// key: 自定义上传文件名称 可以说是时间+string.后缀的形式
// localFile: 填入你本地图片的绝对地址，你也可以把图片放入项目文件中
func (u Uploader) UploadFile(zone *storage.Zone, bucket string, key string, localFile string) (ret PutRsp, err error) {
	// 上传文件自定义返回值结构体
	putPolicy := storage.PutPolicy{
		Scope:      bucket,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)"}`,
	}
	upToken := putPolicy.UploadToken(u.mac)

	cfg := storage.Config{
		Zone:          zone,
		UseCdnDomains: false,
		UseHTTPS:      false,
	}

	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(&cfg)

	putExtra := storage.PutExtra{} // 可选配置 自定义返回字段
	err = formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, &putExtra)
	if err != nil {
		return
	}

	return
}

// 服务端上传一个Reader
func (u Uploader) UploadReader(log *zap.SugaredLogger, bucket string, key string, reader io.Reader, size int64) (ret PutRsp, err error) {
	zone, err := storage.GetZone(u.mac.AccessKey, bucket)
	if err != nil {
		return
	}

	putExtra := storage.PutExtra{} // 可选配置 自定义返回字段
	putPolicy := storage.PutPolicy{
		Scope:      bucket,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)"}`,
	}
	upToken := putPolicy.UploadToken(u.mac)

	cfg := storage.Config{
		Zone:          zone,
		UseCdnDomains: false,
		UseHTTPS:      false,
	}

	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(&cfg)

	if key == "" {
		err = formUploader.PutWithoutKey(context.Background(), &ret, upToken, reader, size, &putExtra)
	} else {
		err = formUploader.Put(context.Background(), &ret, upToken, key, reader, size, &putExtra)
	}
	if err != nil {
		return
	}

	return
}
