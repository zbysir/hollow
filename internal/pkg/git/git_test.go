package git

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/zbysir/hollow/internal/pkg/log"
	"os"
	"testing"
)

func TestPush(t *testing.T) {
	token, ok := os.LookupEnv("GIT_TOKEN")
	if !ok {
		t.Fatal("can't get token from env")
	}
	g, err := NewGit(token, osfs.New("./testdata"), log.StdLogger)
	if err != nil {
		t.Fatal(err)
	}
	err = g.Push("https://github.com/zbysir/2.git", "master", "from test", true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("ok")
}

func TestPull(t *testing.T) {
	token, ok := os.LookupEnv("GIT_TOKEN")
	if !ok {
		t.Fatal("can't get token from env")
	}
	g, err := NewGit(token, osfs.New("./testdata"), log.StdLogger)
	if err != nil {
		t.Fatal(err)
	}
	err = g.Pull("https://github.com/zbysir/2.git", "master", true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("ok")
}
