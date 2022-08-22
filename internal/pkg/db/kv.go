package db

import (
	"fmt"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"path/filepath"
)

func init() {
	boltdb.Register()
}

type KvDb struct {
	dbDir string
}

func NewKvDb(dbDir string) (*KvDb, error) {
	return &KvDb{
		dbDir: dbDir,
	}, nil
}

func (k *KvDb) Open(database, table string) (store.Store, error) {
	kv, err := libkv.NewStore(store.BOLTDB, []string{filepath.Join(k.dbDir, fmt.Sprintf("%s.boltdb", database))}, &store.Config{Bucket: table})
	if err != nil {
		return nil, err
	}
	return kv, nil
}
