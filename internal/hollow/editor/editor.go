// 使用 Gin 暴露 http api

package editor

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/gorilla/websocket"
	"github.com/thoas/go-funk"
	"github.com/zbysir/hollow/front/editor"
	"github.com/zbysir/hollow/internal/hollow"
	"github.com/zbysir/hollow/internal/pkg/auth"
	"github.com/zbysir/hollow/internal/pkg/config"
	"github.com/zbysir/hollow/internal/pkg/easyfs"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/httpsrv"
	"github.com/zbysir/hollow/internal/pkg/log"
	ws "github.com/zbysir/hollow/internal/pkg/ws"
	"go.uber.org/zap"
	"io/fs"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Editor 是编辑器，提供以下功能：
// - web ui 编辑文章
//  - 上传文件到本地或 OSS
// - 实时预览文章
type Editor struct {
	hub              *ws.WsHub
	projectFsFactory FsFactory
	config           Config
}

type Config struct {
	PreviewDomain string // 只要当访问域名能匹配上时，才会渲染，否则显示编辑器
	Secret        string
}

type FsFactory func(pid int64) (billy.Filesystem, error)

func NewEditor(
	projectFsFactory FsFactory,
	config Config,
) *Editor {
	hub := ws.NewHub()
	return &Editor{
		hub:              hub,
		projectFsFactory: projectFsFactory,
		config:           config,
	}
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
		method := c.Request.Method               // 请求方法
		origin := c.Request.Header.Get("Origin") // 请求头部
		//var headerKeys []string                  // 声明请求头keys
		//for k, _ := range c.Request.Header {
		//	headerKeys = append(headerKeys, k)
		//}
		//headerStr := strings.Join(headerKeys, ", ")
		//if headerStr != "" {
		//	headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		//} else {
		//	headerStr = "access-control-allow-origin, access-control-allow-headers"
		//}
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*") // 这是允许访问所有域
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma, Cookie")
		c.Header("Access-Control-Allow-Methods", "OPTIONS,GET,PUT,POST,DELETE")
		c.Header("Access-Control-Allow-Credentials", "true") //  跨域请求是否需要带cookie信息 默认设置为true

		//放行所有 OPTIONS 方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}

		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, e := range c.Errors {
			err := e.Err
			log.Infof("3 %v", err)

			code := 400
			if errors.Is(err, AuthErr) {
				code = 401
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"code": code,
				"msg":  err.Error(),
			})

			return
		}
	}

}

var AuthErr = errors.New("need login")

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		t, _ := c.Cookie("token")
		if t == "" {
			c.Error(AuthErr)
			c.Abort()
			return
		}
		if !auth.CheckToken(secret, t) {
			c.Error(AuthErr)
			c.Abort()
			return
		}

		c.Next()
	}
}

var upgrader = websocket.Upgrader{
	// 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 一个项目可以有多个 FS，比如存储源文件，比如存储主题
func (a *Editor) projectFs(pid int64, bucket string) (billy.Filesystem, error) {
	return a.projectFsFactory(pid)

}

// localhost:9090/api/file/tree
func (a *Editor) Run(ctx context.Context, addr string) (err error) {
	if !config.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(Cors())
	var handleEditorFront = func(c *gin.Context) {
		sub, _ := fs.Sub(editor.EditorFront, "build")
		http.FileServer(http.FS(sub)).ServeHTTP(c.Writer, c.Request)
		return
	}

	var handleRender = func(c *gin.Context) {
		fsSource, err := a.projectFsFactory(0)
		if err != nil {
			c.Error(err)
			return
		}
		b, err := hollow.NewHollow(hollow.Option{
			SourceFs: fsSource,
		})

		b.ServiceHandle(hollow.ExecOption{
			Log:   nil,
			IsDev: true,
		})(c.Writer, c.Request)
		//c.Abort()
	}

	// 编辑 或者 实时预览
	r.NoRoute(func(c *gin.Context) {
		if matchDomain(a.config.PreviewDomain, c.Request.Host) {
			handleRender(c)
		} else {
			handleEditorFront(c)
		}
	})

	var gateway = r.Group("/").Use(ErrorHandler())

	gateway.Use(Auth(a.config.Secret)).GET("/ws/:key", func(c *gin.Context) {
		key := c.Param("key")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Error(err)
			return
		}
		a.hub.Add(key, conn)
	})

	var api = r.Group("/api").Use(ErrorHandler(), Cors())
	api.POST("/auth", func(c *gin.Context) {
		// 创建 token
		var p struct {
			Secret string `json:"secret"`
		}
		err = c.BindJSON(&p)
		if err != nil {
			c.Error(err)
			return
		}

		// 如果是空，则验证 token
		if p.Secret == "" {
			t, _ := c.Cookie("token")
			if t == "" {
				c.Error(AuthErr)
				return
			}
			if !auth.CheckToken(a.config.Secret, t) {
				c.Error(AuthErr)
				return
			}

			c.JSON(200, "ok")
			return
		}
		log.Infof("a.config.Secret, %v %v", a.config.Secret, p.Secret)
		if p.Secret != a.config.Secret {
			c.Error(AuthErr)
			return
		}
		t := auth.CreateToken(p.Secret)
		c.SetCookie("token", t, 7*24*3600, "", "localhost:9432", false, true)
		c.JSON(200, "ok")
	})

	apiAuth := api.Use(Auth(a.config.Secret))

	apiAuth.GET("/setting", func(c *gin.Context) {
		c.JSON(200, map[string]interface{}{
			"preview_domain": a.config.PreviewDomain,
		})
	})
	apiAuth.GET("/config", func(c *gin.Context) {
		//fsTheme, err := a.projectFs(0, "theme")
		//if err != nil {
		//	c.Error(err)
		//	return
		//}

		fsSource, err := a.projectFs(0, "project")
		if err != nil {
			c.Error(err)
			return
		}
		b, err := hollow.NewHollow(hollow.Option{
			SourceFs: fsSource,
			//ThemeFs: fsTheme,
		})

		conf, err := b.LoadConfig(false)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(200, conf)
	})

	apiAuth.GET("/file/tree", func(c *gin.Context) {
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

		ft, err := easyfs.GetFileTree(gobilly.NewStdFs(fs), p.Path, 10)
		c.JSON(200, ft)
	})

	// 打开文件
	apiAuth.GET("/file", func(c *gin.Context) {
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
		f, err := easyfs.GetFile(gobilly.NewStdFs(fs), p.Path)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		c.JSON(200, f)
	})

	// 写入文件
	apiAuth.PUT("/file", func(c *gin.Context) {
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

		f, err := fs.OpenFile(p.Path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
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
	apiAuth.POST("/file", func(c *gin.Context) {
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
	apiAuth.PUT("/file/upload", func(c *gin.Context) {
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
	apiAuth.DELETE("/file", func(c *gin.Context) {
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
	apiAuth.POST("/directory", func(c *gin.Context) {
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

	apiAuth.POST("/publish", func(c *gin.Context) {
		var p publishParams
		err = c.BindJSON(&p)
		if err != nil {
			c.Error(err)
			return
		}

		fs, err := a.projectFs(p.ProjectId, "project")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		key := funk.RandomString(6)

		b, err := hollow.NewHollow(hollow.Option{
			SourceFs: fs,
			//ThemeFs: fsTheme,
		})
		if err != nil {
			c.Error(err)
			return
		}

		start := time.Now()

		go func() {
			defer a.hub.Close(key)

			logWs := NewWsLog(a.hub, key)
			hollowLog := logWs.Named("[Hollow]")
			hollowLog.Infof("start publish")

			dst := memfs.New()
			err = b.BuildAndPublish(context.Background(), dst, hollow.ExecOption{
				Log: logWs,
			})
			if err != nil {
				hollowLog.Errorf("publish fail: %v", err)
				return
			}
			hollowLog.Infof("publish success in %s", time.Now().Sub(start))
		}()

		c.JSON(200, key)
	})

	apiAuth.POST("/pull", func(c *gin.Context) {
		//fsTheme, err := a.projectFs(0, "theme")
		//if err != nil {
		//	c.AbortWithError(400, err)
		//	return
		//}

		filesystem, err := a.projectFs(0, "project")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		key := funk.RandomString(6)

		b, err := hollow.NewHollow(hollow.Option{
			SourceFs: filesystem,
		})
		if err != nil {
			c.Error(err)
			return
		}

		start := time.Now()

		go func() {
			defer func() {
				a.hub.Close(key)
			}()

			logWs := NewWsLog(a.hub, key)
			holloLog := logWs.Named("[Hollow]")
			holloLog.Infof("start pull")

			err = b.PullProject(hollow.ExecOption{
				Log: logWs,
			})
			if err != nil {
				holloLog.Errorf("pull fail: %v", err)
				return
			}
			holloLog.Infof("pull success in %s", time.Now().Sub(start))
		}()

		c.JSON(200, key)
	})

	apiAuth.POST("/push", func(c *gin.Context) {
		//fsTheme, err := a.projectFs(0, "theme")
		//if err != nil {
		//	c.AbortWithError(400, err)
		//	return
		//}

		fs, err := a.projectFs(0, "project")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		key := funk.RandomString(6)

		b, err := hollow.NewHollow(hollow.Option{
			SourceFs: fs,
			//ThemeFs: fsTheme,
		})
		if err != nil {
			c.Error(err)
			return
		}

		start := time.Now()

		go func() {
			defer func() {
				a.hub.Close(key)
			}()

			logWs := NewWsLog(a.hub, key)
			holloLog := logWs.Named("[Hollow]")
			holloLog.Infof("start push")

			err = b.PushProject(hollow.ExecOption{
				Log: logWs,
			})
			if err != nil {
				holloLog.Errorf("push fail: %v", err)
				return
			}
			holloLog.Infof("push success in %s", time.Now().Sub(start))
		}()

		c.JSON(200, key)
	})

	s, err := httpsrv.NewService(addr)
	if err != nil {
		return
	}
	s.Handler("/", r.Handler().ServeHTTP)
	err = s.Start(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
			log.Infof("http service shutdown")
		} else {
			return err
		}
	}
	return nil
}

func NewWsLog(hub *ws.WsHub, key string) *zap.SugaredLogger {
	logWs := log.New(log.Options{
		IsDev:         false,
		To:            hub.GetKeyWrite(key),
		DisableCaller: true,
		CallerSkip:    0,
		Name:          "",
		DisableTime:   true,
	})
	return logWs
}

func matchDomain(match, domain string) bool {
	// split port
	ss := strings.LastIndex(domain, ":")
	if ss != -1 {
		domain = domain[:ss]
	}
	return matchDomainArray(strings.Split(match, "."), strings.Split(domain, "."))
}

func matchDomainArray(match, domain []string) bool {
	if len(match) != len(domain) {
		return false
	}

	if len(match) == 0 {
		return true
	}
	if match[0] == "*" {
		return true
	}
	if match[0] != domain[0] {
		return false
	}
	return matchDomainArray(match[1:], domain[1:])
}
