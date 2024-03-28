// md5的加密
package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

//转小写
func Md5Encode(data string) string {

	hasher := md5.New()
    _, err := hasher.Write([]byte(data))
    if err != nil {
        return ""
    }
    return hex.EncodeToString(hasher.Sum(nil))
}

//转大写
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

//加随机数， 加密操作
func MakePassword(plainpwd,salt string) string {
	return Md5Encode(plainpwd + salt)
}

//解密
func ValidPassword(plainpwd, salt string, password string) bool {
	return Md5Encode(plainpwd + salt) == password
}