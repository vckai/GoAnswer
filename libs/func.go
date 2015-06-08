package libs

import (
	"crypto/md5"
	"fmt"
	"io"
)

// 加密密码,转成md5
func GenPwd(password string) string {
	h := md5.New()
	io.WriteString(h, password)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GenRes() {

}

func GetReq() {

}
