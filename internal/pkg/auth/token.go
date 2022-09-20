package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

var t = time.Now()

func CreateToken(key string) string {
	return MD5(fmt.Sprintf("hollow%vhollow", key))
}

func CheckToken(key string, token string) bool {
	return CreateToken(key) == token
}

func MD5(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}
