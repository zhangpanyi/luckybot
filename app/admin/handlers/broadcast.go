package handlers

import (
	"encoding/json"
	"net/http"

	"luckybot/app/logic/pusher"
	"luckybot/app/storage/models"
)

// 广播消息请求
type BroadcastRequest struct {
	Message string `json:"message"` // 消息内容
	Tonce   int64  `json:"tonce"`   // 时间戳
}

// 广播消息响应
type BroadcastRespone struct {
	OK bool `json:"ok"` // 是否成功
}

// 广播消息
func Broadcast(w http.ResponseWriter, r *http.Request) {
	// 跨域访问
	allowAccessControl(w)

	// 验证权限
	sessionID, data, ok := authentication(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(makeErrorRespone("", ""))
		return
	}

	// 解析请求参数
	var request BroadcastRequest
	if err := json.Unmarshal(data, &request); err != nil || len(request.Message) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 广播消息
	var jsb []byte
	model := models.SubscriberModel{}
	subscribers, err := model.GetSubscribers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}
	for _, userID := range subscribers {
		pusher.Post(userID, request.Message, true, nil)
	}

	respone := BroadcastRespone{OK: true}
	jsb, err = json.Marshal(&respone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(makeRespone(sessionID, jsb))
}
