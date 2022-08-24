// 使用 Gin 暴露 http api

package editor

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zbysir/blog/internal/bblog"
	"github.com/zbysir/blog/internal/bblog/storage"
	"github.com/zbysir/blog/internal/pkg/db"
	"github.com/zbysir/blog/internal/pkg/easyfs"
	"github.com/zbysir/blog/internal/pkg/git"
	"github.com/zbysir/blog/internal/pkg/gobilly"
	"github.com/zbysir/blog/internal/pkg/log"
	ws "github.com/zbysir/blog/internal/pkg/ws"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Editor struct {
	hub *ws.WsHub
	//dbDir   string
	project *storage.Project
	db      *db.KvDb
}

func NewEditor(db *db.KvDb, project *storage.Project) *Editor {
	hub := ws.NewHub()
	return &Editor{hub: hub, db: db, project: project}
}

type fileTreeParams struct {
	ProjectId int64  `form:"project_id"`
	Bucket    string `form:"bucket"`
	Path      string `form:"path"`
}

type deleteFileParams struct {
	IsDir     bool   `form:"is_dir"`
	Path      string `form:"path"`
	Bucket    string `form:"bucket"`
	ProjectId int64  `form:"project_id"`
}

type fileModifyParams struct {
	ProjectId int64  `json:"project_id"`
	Bucket    string `json:"bucket"`
	Path      string `json:"path"`
	Body      string `json:"body"`
}
type publishParams struct {
	ProjectId int64 `json:"project_id"`
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

var upgrader = websocket.Upgrader{
	// 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 一个项目可以有多个 FS，比如存储源文件，比如存储主题
func (a *Editor) projectFs(pid int64, bucket string) (*gobilly.DbFs, error) {
	st, err := a.db.Open(fmt.Sprintf("project_%v", pid), bucket)
	if err != nil {
		return nil, err
	}
	fs := gobilly.NewDbFs(st)
	if err != nil {
		return nil, fmt.Errorf("new fs error: %w", err)
	}
	return fs, nil
}

// localhost:9090/api/file/tree
func (a *Editor) Run(ctx context.Context, addr string) (err error) {
	r := gin.Default()
	r.Use(Cors())

	r.Any("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Error(err)
			return
		}
		a.hub.Add("", conn)
	})

	api := r.Group("/api", func(c *gin.Context) {})

	api.GET("/file/tree", func(c *gin.Context) {
		var p fileTreeParams
		err = c.BindQuery(&p)
		if err != nil {
			c.Error(err)
			return
		}
		if p.Bucket == "" || p.ProjectId == 0 {
			c.AbortWithError(400, fmt.Errorf("invalide params"))
			return
		}
		fs, err := a.projectFs(p.ProjectId, p.Bucket)
		if err != nil {
			c.Error(err)
			return
		}

		ft, err := easyfs.NewFs(gobilly.NewStdFs(fs)).FileTree(p.Path, 10)
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
		fs, err := a.projectFs(p.ProjectId, p.Bucket)
		if err != nil {
			c.Error(err)
			return
		}
		f, err := easyfs.NewFs(gobilly.NewStdFs(fs)).GetFile(p.Path)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		c.JSON(200, f)
	})

	// 写入文件
	api.PUT("/file", func(c *gin.Context) {
		var p fileModifyParams
		err = c.BindJSON(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fs, err := a.projectFs(p.ProjectId, p.Bucket)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		f, err := fs.Open(p.Path)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		_, err = f.Write([]byte(p.Body))
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, nil)
	})

	// 新建文件
	api.POST("/file", func(c *gin.Context) {
		var p fileModifyParams
		err = c.BindJSON(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fs, err := a.projectFs(p.ProjectId, p.Bucket)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		_, err = fs.Create(p.Path)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		c.JSON(200, nil)
	})

	// 批量上传文件
	api.PUT("/file/upload", func(c *gin.Context) {
		var p fileTreeParams
		err = c.Bind(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fm, err := c.MultipartForm()
		if err != nil {
			c.Error(err)
			return
		}
		//for _, f := range fm.File {
		//	for _, f := range f {
		//		log.Infof("files %+v", f)
		//
		//	}
		//}

		fs, err := a.projectFs(p.ProjectId, p.Bucket)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		var allFileName []string
		for _, f := range fm.File {
			for _, f := range f {
				ff, err := f.Open()
				if err != nil {
					c.Error(err)
					return
				}
				body, err := ioutil.ReadAll(ff)
				if err != nil {
					c.Error(err)
					return
				}

				_, kv, err := mime.ParseMediaType(f.Header["Content-Disposition"][0])
				if err != nil {
					c.Error(err)
					return
				}

				fname := kv["filename"]
				if fname == "" {
					fname = f.Filename
				}
				fname = strings.TrimPrefix(fname, "/")
				fullPath := filepath.Join(p.Path, fname)
				allFileName = append(allFileName, fullPath)

				f, err := fs.Create(fullPath)
				if err != nil {
					c.AbortWithError(400, err)
					return
				}
				_, err = f.Write(body)
				if err != nil {
					c.AbortWithError(500, err)
					return
				}
			}
		}
		c.JSON(200, allFileName)
	})

	// 删除文件
	api.DELETE("/file", func(c *gin.Context) {
		var p deleteFileParams
		err = c.Bind(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fs, err := a.projectFs(p.ProjectId, p.Bucket)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		if p.IsDir {
			err = fs.Remove(p.Path)
			if err != nil {
				c.AbortWithError(400, err)
				return
			}
		} else {
			err = fs.Remove(p.Path)
			if err != nil {
				c.AbortWithError(400, err)
				return
			}
		}
		c.JSON(200, nil)
	})

	// 新建文件夹
	api.POST("/directory", func(c *gin.Context) {
		var p fileModifyParams
		err = c.BindJSON(&p)
		if err != nil {
			c.Error(err)
			return
		}
		fs, err := a.projectFs(p.ProjectId, p.Bucket)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		err = fs.MkdirAll(p.Path, 0)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		c.JSON(200, nil)
	})

	api.POST("/publish", func(c *gin.Context) {
		var p publishParams
		err = c.BindJSON(&p)
		if err != nil {
			c.Error(err)
			return
		}

		sett, exist, err := a.project.GetSetting(p.ProjectId)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		if !exist {
			c.AbortWithError(400, fmt.Errorf("project [%d] setting is empty, please config it", p.ProjectId))
			return
		}

		fsTheme, err := a.projectFs(p.ProjectId, "theme")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		fs, err := a.projectFs(p.ProjectId, "project")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		b, err := bblog.NewBblog(bblog.Option{
			Fs:      gobilly.NewStdFs(fs),
			ThemeFs: gobilly.NewStdFs(fsTheme),
		})
		if err != nil {
			c.Error(err)
			return
		}

		out := &WsSink{hub: a.hub}

		logWs := log.New(log.Options{
			IsDev:         false,
			To:            out,
			DisableCaller: true,
			CallerSkip:    0,
			Name:          "",
		})

		err = b.Build("./config.ts", "docs", bblog.ExecOption{
			Env: sett.Env,
			Log: logWs.Named("[Bblog]"),
		})
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		g := git.NewGit(sett.GitToken, logWs.Named("[Git]"))
		err = g.Push("docs", sett.GitRemote, time.Now().String(), "docs", true)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
	})

	err = r.Run(addr)
	if err != nil {
		return err
	}
	return nil
}

// WsSink 将 log 写入到 WS
type WsSink struct {
	hub *ws.WsHub
}

func (w *WsSink) Write(p []byte) (n int, err error) {
	//p = append(p, []byte("\r\n")...)
	err = w.hub.SendAll(p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *WsSink) Sync() error {
	return nil
}

func (w *WsSink) Close() error {
	return nil
}
