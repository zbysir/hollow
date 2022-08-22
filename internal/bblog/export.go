package bblog

import (
	"fmt"
	"github.com/zbysir/blog/internal/pkg/log"
	"io"
	"io/fs"
	"os"
	"path"
)

// fSExport 导出 fs 中的文件到本地
type fSExport struct {
	fs fs.FS
}

func (f *fSExport) exportFile(src, dst string) error {
	var err error
	var srcfd fs.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = f.fs.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = fs.Stat(f.fs, src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func (f *fSExport) exportDir(src string, dst string) error {
	var err error
	var fds []os.DirEntry
	var srcinfo os.FileInfo

	if srcinfo, err = fs.Stat(f.fs, src); err != nil {
		return fmt.Errorf("fs.State %s error: %w", src, err)
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}
	if fds, err = fs.ReadDir(f.fs, src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = f.exportDir(srcfp, dstfp); err != nil {
				log.Warnf("exportDir error: %s", err)
			}
		} else {
			if err = f.exportFile(srcfp, dstfp); err != nil {
				log.Warnf("exportFile error: %s", err)
			}
		}
	}
	return nil
}
