package fusefs

import (
	"github.com/docker/libkv/store"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type DbFs = pathfs.FileSystem

func NewDbFs(store store.Store) (DbFs, error) {
	kvfs, err := NewFuseFs(Options{
		Store: store,
		Root:  "",
	})
	if err != nil {
		return nil, err
	}
	return kvfs, nil
}
