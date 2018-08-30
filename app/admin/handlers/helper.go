package handlers

import (
	"encoding/json"
	"net/http"
)

// 身份验证
func authentication(r *http.Request) bool {
	return true
}

// 生成错误响应
func makeErrorRespone(reason string) []byte {
	object := map[string]string{
		"error": reason,
	}
	jsb, _ := json.Marshal(&object)
	return jsb
}
