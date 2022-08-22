package storage

import (
	"encoding/json"
	"fmt"
	"github.com/docker/libkv/store"
)

// Project 存储项目信息，如配置
type Project struct {
	db store.Store
}

func NewProject(db store.Store) *Project {
	return &Project{db: db}
}

type ProjectSetting struct {
	GitRemote string `json:"git_remote"` // remote for push
	GitToken  string `json:"git_token"`  // token for push
	ThemeId   int64  `json:"theme_id"`   // 使用哪一个主题
}

func (p *Project) GetSetting(pid int64) (ps *ProjectSetting, exist bool, err error) {
	kv, err := p.db.Get(projectIdKey(pid, "setting"))
	if err != nil {
		if err == store.ErrKeyNotFound {
			return nil, false, nil
		}
		return
	}
	ps = &ProjectSetting{}
	err = json.Unmarshal(kv.Value, ps)
	if err != nil {
		return
	}

	exist = true
	return
}

func projectIdKey(id int64, sub string) string {
	return fmt.Sprintf("project/%v/%v", id, sub)
}

func (p *Project) SetSetting(pid int64, ps *ProjectSetting) (err error) {
	bs, _ := json.Marshal(ps)
	err = p.db.Put(projectIdKey(pid, "setting"), bs, &store.WriteOptions{
		IsDir: false,
		TTL:   0,
	})
	if err != nil {
		return
	}

	return
}
