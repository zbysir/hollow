package bblog

import (
	"encoding/json"
	"testing"
)

func TestFileTree(t *testing.T) {
	fa := FsApi{fs: StdFileSystem{}}
	ft, err := fa.FileTree("../", 2)
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := json.MarshalIndent(ft, " ", " ")
	t.Logf("%s", bs)

}
