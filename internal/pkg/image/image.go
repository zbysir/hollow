package image

import (
	"errors"
	"fmt"
	"github.com/google/martian/log"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)
import "github.com/nfnt/resize"

//CompressImg compress a jpg or png format image, the new images will be named autoly
func CompressImg(source string, width uint) error {
	var err error
	var file *os.File
	reg, _ := regexp.Compile(`^.*\.((png)|(jpg))$`)
	if !reg.MatchString(source) {
		err = errors.New("%s is not a .png or .jpg file")
		return err
	}
	if file, err = os.Open(source); err != nil {
		return err
	}
	defer file.Close()
	name := file.Name()
	ext := strings.ToLower(filepath.Ext(name))
	var img image.Image
	switch ext {
	case ".png":
		if img, err = png.Decode(file); err != nil {
			return err
		}
	case ".jpg":
		if img, err = jpeg.Decode(file); err != nil {
			return err
		}
	default:
		err = fmt.Errorf("images %s name not right", name)
		return err
	}
	resizeImg := resize.Resize(width, 0, img, resize.Lanczos3)
	newName := newName(source, int(width))
	if outFile, err := os.Create(newName); err != nil {
		return err
	} else {
		defer outFile.Close()
		err = jpeg.Encode(outFile, resizeImg, nil)
		if err != nil {
			return err
		}
	}
	abspath, _ := filepath.Abs(newName)
	log.Infof("New imgs successfully save at: %s", abspath)
	return nil
}

//create a file name for the iamges that after resize
func newName(name string, size int) string {
	dir, file := filepath.Split(name)
	return fmt.Sprintf("%s_%d%s", dir, size, file)
}
