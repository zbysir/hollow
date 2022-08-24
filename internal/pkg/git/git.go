package git

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"go.uber.org/zap"
	"io"
	"time"
)

type Git struct {
	personalAccessTokens string
	log                  *zap.SugaredLogger
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
func NewGit(personalAccessTokens string, log *zap.SugaredLogger) *Git {
	return &Git{
		personalAccessTokens: personalAccessTokens,
		log:                  log,
	}
}

// Push 一个文件夹到 远端仓库
func (g *Git) Push(dst billy.Filesystem, repo string, commitMsg string, branch string, force bool) error {
	start := time.Now()

	dot, _ := dst.Chroot(".git")

	s := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	r, err := git.Init(s, dst)
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
		} else {
			return fmt.Errorf("PlainInit error: %w", err)
		}
	} else {
		g.log.Infof("git init %v", dst.Root())
	}

	if r == nil {
		g.log.Infof("git open %v", dst.Root())
		r, err = git.Open(s, dst)
		if err != nil {
			return fmt.Errorf("PlainOpen error: %w", err)
		}
	}

	wt, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree error: %w", err)
	}
	g.log.Infof("git checkout %v", branch)

	err = r.DeleteRemote("origin")
	if err != nil {
		if err == git.ErrRemoteNotFound {

		} else {
			return err
		}
	}
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{repo},
		Fetch: nil,
	})
	if err != nil {
		return fmt.Errorf("CreateRemote error: %w", err)
	}

	_, err = wt.Add(".")
	if err != nil {
		return fmt.Errorf("add error: %w", err)
	}

	g.log.Infof("git commit %v", commitMsg)
	_, err = wt.Commit(commitMsg, &git.CommitOptions{
		All:       false,
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
	var progress io.Writer = &logWrite{
		log: g.log,
	}
	name := plumbing.NewBranchReferenceName(branch)
	g.log.Infof("git push %v %v", repo, name)
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			// refs/heads/*
			config.RefSpec(fmt.Sprintf("%v:%v", h.Name(), name)),
		},
		Auth: &http.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			// https://docs.github.com/cn/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
			Password: g.personalAccessTokens,
		},
		Progress:          progress,
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
