package bblog

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	jsx "github.com/zbysir/gojsx"
	git2 "github.com/zbysir/hollow/internal/pkg/git"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/log"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// file://Users/bysir/front/bysir/hollow-theme/hollow
func resolveFileUrl(u string) (absPath string, err error) {
	return "/" + strings.TrimPrefix(u, "file://"), nil
}

type ThemeExport struct {
	Pages  Pages
	Assets Assets
}

type ThemeLoader interface {
	Load(x *jsx.Jsx, path string, refresh bool) (ThemeExport, fs.FS, error)
}

// GitThemeLoader
// e.g. https:/github.com/zbysir/hollow-theme/tree/master/hollow/index
type GitThemeLoader struct {
}

func NewGitThemeLoader() *GitThemeLoader {
	return &GitThemeLoader{}
}

// Load 会缓存 fs ，只有当强制刷新时更新
func (g *GitThemeLoader) Load(x *jsx.Jsx, path string, refresh bool) (ThemeExport, fs.FS, error) {
	fileSys := osfs.New("./.themecache")

	_, err := fileSys.Stat(".")
	if err != nil {
		if os.IsNotExist(err) {
			refresh = true
		} else {
			return ThemeExport{}, nil, err
		}
	}
	remote, branch, subPath, err := resolveGitUrl(path)
	if err != nil {
		return ThemeExport{}, nil, err
	}

	if refresh {
		gt, err := git2.NewGit("", fileSys, log.StdLogger)
		if err != nil {
			return ThemeExport{}, nil, err
		}

		err = gt.Pull(remote, branch, true)
		if err != nil {
			return ThemeExport{}, nil, err
		}
	}

	fileSys, err = fileSys.Chroot(subPath)
	if err != nil {
		return ThemeExport{}, nil, err
	}
	f := gobilly.NewStdFs(fileSys)
	return loadTheme(x, f, filepath.Join("index"))
}

// https://github.com/zbysir/hollow-theme/tree/master/hollow
func resolveGitUrl(u string) (remote string, branch string, subPath string, err error) {
	return "https://github.com/zbysir/hollow-theme", "master", "hollow", nil
}

type LocalThemeLoader struct {
	f   fs.FS
	dir string
}

func NewLocalThemeLoader(f fs.FS, dir string) *LocalThemeLoader {
	return &LocalThemeLoader{f: f, dir: dir}
}

func (l *LocalThemeLoader) Load(x *jsx.Jsx, path string, refresh bool) (ThemeExport, fs.FS, error) {
	filePath := filepath.Join(path, "index")
	if l.dir != "" {
		filePath = filepath.Join(l.dir, "index")
	}

	return loadTheme(x, l.f, filePath)
}

func loadTheme(x *jsx.Jsx, fs fs.FS, configFile string) (ThemeExport, fs.FS, error) {
	envBs, _ := json.Marshal(nil)
	processCode := fmt.Sprintf("var process = {env: %s}", envBs)

	// 添加 ./ 告知 module 加载项目文件而不是 node_module
	configFile = "./" + filepath.Clean(configFile)
	v, err := x.RunJs([]byte(fmt.Sprintf(`%s;require("%v").default`, processCode, configFile)), jsx.WithRunFs(fs))
	if err != nil {
		return ThemeExport{}, nil, err
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

	return ThemeExport{Pages: ps, Assets: assets}, fs, nil
}
