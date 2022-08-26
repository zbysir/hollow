package easyfs

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestFileTree(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	ft, err := GetFileTree(os.DirFS("./"), ".", 4)
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(ft, " ", " ")
	t.Logf("%s", bs)
}
