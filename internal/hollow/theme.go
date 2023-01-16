package hollow

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/helper/chroot"
	"github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/asynctask"
	git "github.com/zbysir/hollow/internal/pkg/git"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/util"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ThemeExport struct {
	Pages  Pages
	Assets Assets
}

type ThemeLoader interface {
	Load(ctx *RenderContext, x *gojsx.Jsx, refresh bool, enableAsync bool, jsxOpts ...gojsx.OptionExec) (ThemeExport, fs.FS, *asynctask.Task, error)
}

// GitThemeLoader
// e.g. https:/github.com/zbysir/hollow-theme/tree/master/hollow/index
type GitThemeLoader struct {
	asyncTask *asynctask.Manager
	path      string
	cacheFs   billy.Filesystem
}

func NewGitThemeLoader(asyncTask *asynctask.Manager, path string, cacheFs billy.Filesystem) *GitThemeLoader {
	return &GitThemeLoader{asyncTask: asyncTask, path: path, cacheFs: cacheFs}
}

// Load 会缓存 fs ，只有当强制刷新时更新
func (g *GitThemeLoader) Load(ctx *RenderContext, x *gojsx.Jsx, refresh bool, enableAsync bool, jsxOpts ...gojsx.OptionExec) (ThemeExport, fs.FS, *asynctask.Task, error) {
	fileSys := chroot.New(g.cacheFs, "theme")

	_, err := fileSys.Stat(".git")
	if err != nil {
		if os.IsNotExist(err) {
			refresh = true
		} else {
			return ThemeExport{}, nil, nil, err
		}
	}

	taskKey := util.MD5(g.path)
	// 如果有异步任务，则直接显示异步任务
	task, exist := g.asyncTask.GetTask(taskKey)
	if exist {
		return ThemeExport{}, nil, task, nil
	}

	remote, branch, subPath, err := resolveGitUrl(g.path)
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}

	if refresh {
		if enableAsync {
			task, isNew := g.asyncTask.NewTask(taskKey)
			if isNew {
				go func() {
					var err error
					defer func() {
						if err != nil {
							task.Log("error: " + err.Error())
						}
						task.Done()
					}()

					logger := log.New(log.Options{
						IsDev:         false,
						To:            task,
						DisableCaller: true,
						CallerSkip:    0,
						Name:          "",
						DisableTime:   true,
					})

					gt, err := git.NewGit("", fileSys, logger)
					if err != nil {
						return
					}

					err = gt.Pull(remote, branch, true)
					if err != nil {
						return
					}
				}()
			}

			return ThemeExport{}, nil, task, nil
		} else {
			gt, err := git.NewGit("", fileSys, log.Logger())
			if err != nil {
				return ThemeExport{}, nil, nil, nil
			}

			err = gt.Pull(remote, branch, true)
			if err != nil {
				return ThemeExport{}, nil, nil, nil
			}
		}
	}

	subFs, err := fileSys.Chroot(subPath)
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}
	f := gobilly.NewStdFs(subFs)
	jsxOpts = append(jsxOpts, gojsx.WithFs(f))
	theme, err := execTheme(x, filepath.Join("index"), jsxOpts...)
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}
	return theme, f, nil, nil
}

// https://github.com/zbysir/hollow-theme/tree/master/hollow
func resolveGitUrl(u string) (remote string, branch string, subPath string, err error) {
	ss := strings.Split(u, "/tree/")
	if len(ss) != 2 {
		err = fmt.Errorf("bas url: '%v', support url like 'https://github.com/zbysir/hollow-theme/tree/master/hollow'", u)
		return
	}
	remote = ss[0]
	b := strings.Split(ss[1], "/")
	branch = b[0]
	if branch == "" {
		err = fmt.Errorf("bas url: '%v', support url like 'https://github.com/zbysir/hollow-theme/tree/master/hollow'", u)
		return
	}
	if len(b) > 1 {
		subPath = strings.Join(b[1:], "/")
	}

	return
}

type LocalThemeLoader struct {
	f fs.FS
}

func NewFsThemeLoader(rootFs fs.FS) *LocalThemeLoader {
	return &LocalThemeLoader{f: rootFs}
}

func (l *LocalThemeLoader) Load(ctx *RenderContext, x *gojsx.Jsx, refresh bool, enableAsync bool, jsxOpts ...gojsx.OptionExec) (ThemeExport, fs.FS, *asynctask.Task, error) {
	jsxOpts = append(jsxOpts, gojsx.WithFs(l.f))
	theme, err := execTheme(x, "index", jsxOpts...)
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}

	return theme, l.f, nil, nil
}

func execTheme(jsx *gojsx.Jsx, configFile string, jsxOpts ...gojsx.OptionExec) (ThemeExport, error) {
	envBs, _ := json.Marshal(nil)
	processCode := fmt.Sprintf("var process = {env: %s}", envBs)

	// 添加 ./ 告知 module 加载项目文件而不是 node_module
	configFile = "./" + filepath.Clean(configFile)
	code := fmt.Sprintf(`%s; module.exports = require("%s")`, processCode, configFile)
	v, err := jsx.ExecCode([]byte(code), jsxOpts...)
	if err != nil {
		return ThemeExport{}, err
	}

	// 直接 export 会导致 function 无法捕获 panic，不好实现
	raw := exportGojaValue(v.Default.(gojsx.Any).Any).(map[string]interface{})

	pages := raw["pages"].([]interface{})
	ps := make(Pages, len(pages))
	for i, p := range pages {
		ps[i] = p.(map[string]interface{})
	}
	as := raw["assets"].([]interface{})
	assets := make(Assets, len(as))
	for k, v := range as {
		assets[k] = exportGojaValueToString(v)
	}
	configDir := filepath.Dir(configFile)

	for i, a := range assets {
		// 得到相对 themeFs root 的路径，e.g. dark/publish
		dir := filepath.Join(configDir, a)
		assets[i] = dir
	}

	return ThemeExport{Pages: ps, Assets: assets}, nil
}
