package image

import "testing"

func TestCompressImg(t *testing.T) {
	err := CompressImg("../../../workspace/project/statics/img/IMG_20220226_113219.jpg", 1000)
	if err != nil {
		t.Fatal(err)
	}
}
