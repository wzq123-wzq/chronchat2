package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

//小写的函数只能在当前包中使用
func Md5Encode(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//大写的函数可以在其他包中使用
func MD5Encode(str string) string {
	return strings.ToUpper(Md5Encode(str))
}

//加密
func MakrePassword(password, salt string) string {
	return MD5Encode(password + salt)
}

//解密
func ValidPassword(plainpwd, salt, password string) bool {
	return MakrePassword(plainpwd, salt) == password
}