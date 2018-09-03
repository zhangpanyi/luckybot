package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zhangpanyi/luckybot/app/config"
)

// 身份验证
func authentication(r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	bin, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(bin), ":", 2)
	if len(pair) != 2 {
		return false
	}

	serveCfg := config.GetServe()
	return pair[0] == serveCfg.UserName && pair[1] == serveCfg.Password
}

// 生成错误响应
func makeErrorRespone(reason string) []byte {
	object := map[string]string{
		"error": reason,
	}
	jsb, _ := json.Marshal(&object)
	return jsb
}
