package git

import (
	"github.com/zbysir/blog/internal/pkg/log"
	"os"
	"testing"
)

func TestPush(t *testing.T) {
	token, ok := os.LookupEnv("git_token")
	if !ok {
		t.Fatal("can't get token from env")
	}
	g := NewGit(token, log.StdLogger)
	err := g.Push("./testdata", "https://github.com/zbysir/2.git", "from test", "docs", true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("ok")
}
