package git

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/zbysir/hollow/internal/pkg/easyfs"
	"github.com/zbysir/hollow/internal/pkg/gobilly"
	"github.com/zbysir/hollow/internal/pkg/log"
	"go.uber.org/zap"
	"io/fs"
	"time"
)

// Git 封装 go-git
// go-git 是 git 的子集，很多功能并不支持，查看不同：https://github.com/go-git/go-git/blob/master/COMPATIBILITY.md
// 比如 pull 只支持 fast-forward，不支持 stash，所以在封装的时候回做一些取舍：
// merge 与 解决冲突是十分复杂的操作，hollow 无法实现它们，所以在同步的时候采取以下策略：
//  - pull 时 如果传递 force=true，如果遇到 non-fast-forward，则会将远端文件全部下载下来，cp 到本地，相同文件保留最新的一个。尽量将降低影响。
//  - push：为了避免 push 的冲突，每次 push 都是 force 的，为了避免远端文件丢失，每次 push 之前都会 pull 一次。
// non-fast-forward: 当本地有提交，pull 都会报错 non-fast-forward。
type Git struct {
	log  *zap.SugaredLogger
	dir  billy.Filesystem
	r    *git.Repository
	auth *http.BasicAuth
}

type logWrite struct {
	log *zap.SugaredLogger
}

func (l *logWrite) Write(p []byte) (n int, err error) {
	l.log.Infof("%s", p)
	return len(p), nil
}

// NewGit return Git
// https://docs.github.com/cn/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
func NewGit(personalAccessTokens string, dir billy.Filesystem, log *zap.SugaredLogger) (g *Git, err error) {
	var auth *http.BasicAuth
	if personalAccessTokens != "" {
		auth = &http.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			// https://docs.github.com/cn/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
			Password: personalAccessTokens,
		}
	}
	g = &Git{
		log:  log,
		dir:  dir,
		r:    nil,
		auth: auth,
	}
	g.r, err = g.initRepo(dir)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Git) initRepo(dir billy.Filesystem) (*git.Repository, error) {
	dot, _ := dir.Chroot(".git")
	s := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	r, err := git.Init(s, dir)
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
		} else {
			return nil, fmt.Errorf("PlainInit error: %w", err)
		}
	} else {
		g.log.Infof("git init %v", dir.Root())
	}

	if r == nil {
		g.log.Infof("git open %v", dir.Root())
		r, err = git.Open(s, dir)
		if err != nil {
			return nil, fmt.Errorf("PlainOpen error: %w", err)
		}
	}

	return r, nil
}

func getFileLastCommitAt(r *git.Repository, filename string) (t time.Time) {
	l, _ := r.Log(&git.LogOptions{
		From:       plumbing.Hash{},
		Order:      0,
		FileName:   &filename,
		PathFilter: nil,
		All:        false,
		Since:      nil,
		Until:      nil,
	})
	l.ForEach(func(commit *object.Commit) error {
		t = commit.Author.When
		return fmt.Errorf("skil")
	})

	return
}

// Sync 当无法 pull 成功时，将使用 sync 逻辑将文件同步下来。
func (g *Git) Sync(remote string, branch string) error {
	dir := memfs.New()
	dot, _ := dir.Chroot(".git")
	s := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	referenceName := plumbing.NewBranchReferenceName(branch)
	r, err := git.Clone(s, dir, &git.CloneOptions{
		URL:               remote,
		Auth:              g.auth,
		RemoteName:        "",
		ReferenceName:     referenceName,
		SingleBranch:      false,
		NoCheckout:        false,
		Depth:             0,
		RecurseSubmodules: 0,
		Progress: &logWrite{
			log: g.log,
		},
		Tags:            0,
		InsecureSkipTLS: false,
		CABundle:        nil,
	})
	if err != nil {
		return err
	}

	stdFs := gobilly.NewStdFs(dir)
	err = fs.WalkDir(stdFs, "./", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == ".git" {
			return fs.SkipDir
		}

		// copy if file is latest or not exist
		locStat, err := g.dir.Stat(path)
		needCopy := false
		if err != nil {
			if err == fs.ErrNotExist {
				err = nil
				needCopy = true
			} else {
				return fmt.Errorf("stat local file '%v' error: %w", path, err)
			}
		} else {
			if locStat.ModTime().Before(getFileLastCommitAt(r, path)) {
				needCopy = true
			}
		}
		if needCopy {
			log.Infof("copy file: %+v", path)
			err = easyfs.CopyFile(path, path, stdFs, g.dir)
			if err != nil {
				return fmt.Errorf("copy file '%v' to local error: %w", path, err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *Git) Pull(remote string, branch string, force bool) error {
	remoteName := "origin-temp"
	err := g.r.DeleteRemote(remoteName)
	if err != nil {
		if err == git.ErrRemoteNotFound {
			err = nil
		} else {
			return err
		}
	}

	_, err = g.r.CreateRemote(&config.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{remote},
		Fetch: nil,
	})
	if err != nil {
		return fmt.Errorf("CreateRemote error: %w", err)
	}
	defer func() {
		err = g.r.DeleteRemote(remoteName)
		if err != nil {
			if err == git.ErrRemoteNotFound {
				err = nil
			} else {
				g.log.Errorf("DeleteReomte error: %v", err)
			}
		}
	}()

	wt, err := g.r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree error: %w", err)
	}

	// local branch
	//localBranch := branch + "-tmo"
	//localReferenceName := plumbing.NewBranchReferenceName(localBranch)
	//
	//s, err := g.r.Reference(localReferenceName, false)
	//if err != nil {
	//	if err == plumbing.ErrReferenceNotFound {
	//		err = wt.Checkout(&git.CheckoutOptions{
	//			Hash:   plumbing.Hash{},
	//			Branch: localReferenceName,
	//			Create: true,
	//			Force:  true,
	//			Keep:   false,
	//		})
	//		if err != nil {
	//			return fmt.Errorf("checkout %v with create error: %+v", branch, err)
	//		}
	//	} else {
	//		err = wt.Checkout(&git.CheckoutOptions{
	//			Hash:   s.Hash(),
	//			Branch: "",
	//			Create: false,
	//			Force:  true,
	//			Keep:   false,
	//		})
	//		if err != nil {
	//			return fmt.Errorf("checkout %v error: %+v", branch, err)
	//		}
	//	}
	//	return err
	//}

	referenceName := plumbing.NewBranchReferenceName(branch)
	err = wt.Pull(&git.PullOptions{
		RemoteName:        remoteName,
		ReferenceName:     referenceName,
		SingleBranch:      true,
		Depth:             0,
		Auth:              g.auth,
		RecurseSubmodules: 0,
		Progress: &logWrite{
			log: g.log,
		},
		Force:           force,
		InsecureSkipTLS: false,
		CABundle:        nil,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			err = nil
		} else if err == git.ErrNonFastForwardUpdate {
			if force {
				err = g.Sync(remote, branch)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			return fmt.Errorf("pull error: %v", err)
		}
	}

	return nil
}

// Push 一个本地仓库 推送到 远端仓库
func (g *Git) Push(remote string, branch string, commitMsg string, force bool) error {
	start := time.Now()
	r := g.r

	g.log.Infof("git checkout %v", branch)
	remoteName := "origin-temp"
	err := r.DeleteRemote(remoteName)
	if err != nil {
		if err == git.ErrRemoteNotFound {
			err = nil
		} else {
			return err
		}
	}
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{remote},
		Fetch: nil,
	})
	if err != nil {
		return fmt.Errorf("CreateRemote error: %w", err)
	}

	defer func() {
		err = g.r.DeleteRemote(remoteName)
		if err != nil {
			if err == git.ErrRemoteNotFound {
				err = nil
			} else {
				g.log.Errorf("DeleteReomte error: %v", err)
			}
		}
	}()

	wt, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree error: %w", err)
	}
	patterns, err := gitignore.ReadPatterns(wt.Filesystem, nil)
	if err != nil {
		return fmt.Errorf("ReadPatterns error: %w", err)
	}
	wt.Excludes = patterns
	err = wt.AddWithOptions(&git.AddOptions{
		All:  true,
		Path: "",
		Glob: "",
	})
	if err != nil {
		return fmt.Errorf("add error: %w", err)
	}

	g.log.Infof("git commit %v", commitMsg)
	_, err = wt.Commit(commitMsg, &git.CommitOptions{
		All:       true,
		Author:    nil,
		Committer: nil,
		Parents:   nil,
		SignKey:   nil,
	})
	if err != nil {
		return fmt.Errorf("commit error: %w", err)
	}

	h, err := r.Head()
	if err != nil {
		return fmt.Errorf("head error: %w", err)
	}

	//ll, err := r.Log(&git.LogOptions{})
	//ll.ForEach(func(commit *object.Commit) error {
	//	stat, err := commit.Stats()
	//	if err != nil {
	//		return err
	//	}
	//	for _, i := range stat {
	//		log.Infof("commit %s: file %v: %v", commit.ID(), i.Name, i)
	//	}
	//	//fs, err := commit.Files()
	//	//if err != nil {
	//	//	return err
	//	//}
	//	//
	//	//fs.ForEach(func(file *object.File) error {
	//	//	ls, err := file.Lines()
	//	//	if err != nil {
	//	//		return err
	//	//	}
	//	//	log.Infof("commit %s: file %s: %v", commit.ID(), file.Name, ls)
	//	//	return nil
	//	//})
	//	return nil
	//})
	name := plumbing.NewBranchReferenceName(branch)
	g.log.Infof("git push %v %v", remote, name)
	err = r.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs: []config.RefSpec{
			// refs/heads/*
			config.RefSpec(fmt.Sprintf("%v:%v", h.Name(), name)),
		},
		Auth: g.auth,
		Progress: &logWrite{
			log: g.log,
		},
		Prune:             false,
		Force:             force,
		InsecureSkipTLS:   false,
		CABundle:          nil,
		RequireRemoteRefs: nil,
	})
	if err != nil {
		return fmt.Errorf("push error: %w", err)
	}

	g.log.Infof("Done in %v", time.Now().Sub(start))

	return nil
}
