package hollow

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	jsx "github.com/zbysir/gojsx"
	"github.com/zbysir/hollow/internal/pkg/asynctask"
	git "github.com/zbysir/hollow/internal/pkg/git"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/log"
	"github.com/zbysir/hollow/internal/pkg/util"
	"io/fs"
	"os"
	"path/filepath"
)

type ThemeExport struct {
	Pages  Pages
	Assets Assets
}

type ThemeLoader interface {
	Load(ctx context.Context, x *jsx.Jsx, path string, refresh bool, enableAsync bool) (ThemeExport, fs.FS, *asynctask.Task, error)
}

// GitThemeLoader
// e.g. https:/github.com/zbysir/hollow-theme/tree/master/hollow/index
type GitThemeLoader struct {
	asyncTask *asynctask.Manager
}

func NewGitThemeLoader(asyncTask *asynctask.Manager) *GitThemeLoader {
	return &GitThemeLoader{asyncTask: asyncTask}
}

// Load 会缓存 fs ，只有当强制刷新时更新
func (g *GitThemeLoader) Load(ctx context.Context, x *jsx.Jsx, path string, refresh bool, enableAsync bool) (ThemeExport, fs.FS, *asynctask.Task, error) {
	fileSys := osfs.New("./.themecache")

	_, err := fileSys.Stat(".")
	if err != nil {
		if os.IsNotExist(err) {
			refresh = true
		} else {
			return ThemeExport{}, nil, nil, err
		}
	}
	remote, branch, subPath, err := resolveGitUrl(path)
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}

	if refresh {
		if enableAsync {
			task, isNew := g.asyncTask.NewTask(util.MD5(path))
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
			gt, err := git.NewGit("", fileSys, log.StdLogger)
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
	theme, err := execTheme(x, f, filepath.Join("index"))
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}
	return theme, f, nil, nil
}

// https://github.com/zbysir/hollow-theme/tree/master/hollow
func resolveGitUrl(u string) (remote string, branch string, subPath string, err error) {
	return "https://github.com/zbysir/hollow-theme", "master", "hollow", nil
}

type LocalThemeLoader struct {
	f fs.FS
}

func NewLocalThemeLoader(rootFs fs.FS) *LocalThemeLoader {
	return &LocalThemeLoader{f: rootFs}
}

func (l *LocalThemeLoader) Load(ctx context.Context, x *jsx.Jsx, path string, refresh bool, enableAsync bool) (ThemeExport, fs.FS, *asynctask.Task, error) {
	theme, err := execTheme(x, l.f, "index")
	if err != nil {
		return ThemeExport{}, nil, nil, err
	}

	return theme, l.f, nil, nil
}

func execTheme(x *jsx.Jsx, filesys fs.FS, configFile string) (ThemeExport, error) {
	envBs, _ := json.Marshal(nil)
	processCode := fmt.Sprintf("var process = {env: %s}", envBs)

	// 添加 ./ 告知 module 加载项目文件而不是 node_module
	configFile = "./" + filepath.Clean(configFile)
	v, err := x.RunJs([]byte(fmt.Sprintf(`%s;require("%v").default`, processCode, configFile)), jsx.WithRunFs(filesys))
	if err != nil {
		return ThemeExport{}, fmt.Errorf("execTheme '%v' error: %w", configFile, err)
	}

	// 直接 export 会导致 function 无法捕获 panic，不好实现
	raw := exportGojaValue(v).(map[string]interface{})

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
