// 使用 Gin 暴露 http api

package bblog

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/libkv/store"
	"github.com/gin-gonic/gin"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/zbysir/blog/internal/fs"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

type Editor struct {
}

type fileTreeParams struct {
	Bucket string `form:"bucket"`
	Path   string `form:"path"`
}

type fileModifyParams struct {
	Bucket string `json:"bucket"`
	Path   string `json:"path"`
	Body   string `json:"body"`
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               //请求方法
		origin := c.Request.Header.Get("Origin") //请求头部
		var headerKeys []string                  // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")                                       // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")                                                                                                                                                              // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next() //  处理请求
	}
}

// localhost:9090/api/file/tree
func (a *Editor) Run(ctx context.Context) (err error) {
	r := gin.Default()
	r.Use(Cors())
	api := r.Group("/api")
	api.GET("/file/tree", func(c *gin.Context) {
		var p fileTreeParams
		err = c.BindQuery(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fs, err := NewFsApi("test")
		if err != nil {
			c.Error(err)
			return
		}

		ft, err := fs.FileTree(p.Path, 10)
		c.JSON(200, ft)
	})

	// 打开文件
	api.GET("/file", func(c *gin.Context) {
		var p fileTreeParams
		err = c.BindQuery(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fs, err := NewFsApi("test")
		if err != nil {
			c.Error(err)
			return
		}

		ft, err := fs.GetFile(p.Path)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		c.JSON(200, ft)
	})

	// 写入文件
	api.PUT("/file", func(c *gin.Context) {
		var p fileModifyParams
		err = c.BindJSON(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fs, err := NewFsApi("test")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		err = fs.WriteFile(p.Path, p.Body)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		c.JSON(200, nil)
	})

	r.Run(":9090") // listen and serve on 0.0.0.0:8080
	return nil
}

type Handler struct {
}

type File struct {
	Name      string `json:"name"`
	Path      string `json:"path"` // full path
	IsDir     bool   `json:"is_dir"`
	CreatedAt int64  `json:"created_at"`
	ModifyAt  int64  `json:"modify_at"`
	Body      string `json:"body"`
}

type FileTree struct {
	File
	Items []FileTree `json:"items"`
}

type FsApi struct {
	fs *fs.FS
}

func NewFsApi(bucket string) (*FsApi, error) {
	kvfs, err := fs.NewKVFS(fs.Options{
		Store: string(store.BOLTDB),
		Addrs: []string{"./db.db"},
		Root:  "",
		Config: store.Config{
			Bucket: bucket,
		},
	})
	if err != nil {
		return nil, err
	}
	return &FsApi{
		fs: kvfs,
	}, nil
}

func (f *FsApi) Mkdir(name string) (err error) {
	return
}

func (f *FsApi) RmDir(name string) (err error) {
	return
}

func (f *FsApi) RmFile(name string) (err error) {
	status := f.fs.Unlink(name, nil)
	if !status.Ok() {
		return statusError(status)
	}
	return
}

func statusError(status fuse.Status) error {
	if status.Ok() {
		return nil
	}
	return fmt.Errorf("%s", status.String())
}

func (f *FsApi) GetFile(path string) (fi *File, err error) {
	nf, status := f.fs.Open(path, 0, nil)
	if !status.Ok() {
		return nil, statusError(status)
	}

	var a fuse.Attr
	nf.GetAttr(&a)

	bs, err := ioutil.ReadAll(fs.NewIOReader(nf))
	if err != nil {
		return
	}

	_, name := filepath.Split(path)
	return &File{
		Name:      name,
		Path:      path,
		IsDir:     a.Mode&fuse.S_IFDIR == fuse.S_IFDIR,
		CreatedAt: int64(a.Ctime),
		ModifyAt:  int64(a.Mtime),
		Body:      string(bs),
	}, nil
}

// WriteFile 写文件，如果文件存在则会被覆盖
func (f *FsApi) WriteFile(path string, content string) (err error) {
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

func (f *FsApi) FileTree(base string, deep int) (ft FileTree, err error) {
	_, ft.Name = path.Split(base)
	ft.Path = base
	ft.IsDir = true
	if deep == 0 {
		return
	}
	//f.fs.StatFs()
	fds, status := f.fs.OpenDir(base, nil)
	if !status.Ok() {
		return ft, errors.New(status.String())
	}

	//f.ns.OpenDir()
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
					Name:  fd.Name,
					Path:  srcfp,
					IsDir: false,
				},
				Items: nil,
			})
		}
	}

	return ft, nil
}
