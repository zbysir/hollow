package bblog

import "testing"

func TestExport(t *testing.T) {
	e := fSExport{fs: StdFileSystem{}}
	err := e.exportDir("../bblog", "../.cached")
	if err != nil {
		t.Fatal(err)
	}
}
