package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// 身份验证
func authentication(r *http.Request) (string, []byte, bool) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return "", nil, false
	}

	sessionID, src, ok := authenticator.Decode(data)
	if !ok {
		return "", nil, false
	}
	authenticator.KeepAlive(sessionID)
	return sessionID, src, true
}

// 生成响应
func makeRespone(sessionID string, result []byte) []byte {
	respone := struct {
		OK     bool            `json:"ok"`
		Result json.RawMessage `json:"result"`
	}{
		OK:     true,
		Result: result,
	}
	src, _ := json.Marshal(&respone)
	data := authenticator.Encode(sessionID, src)
	return data
}

// 生成错误响应
func makeErrorRespone(sessionID, reason string) []byte {
	respone := struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}{
		OK:    true,
		Error: reason,
	}
	src, _ := json.Marshal(&respone)
	data := authenticator.Encode(sessionID, src)
	return data
}
