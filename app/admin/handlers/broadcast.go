package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/zhangpanyi/luckybot/app/logic/pusher"
	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 广播消息请求
type BroadcastRequest struct {
	Message string `json:"message"` // 消息内容
}

// 广播消息响应
type BroadcastRespone struct {
	OK bool `json:"ok"` // 是否成功
}

// 广播消息
func Broadcast(w http.ResponseWriter, r *http.Request) {
	// 验证权限
	if !authentication(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 解析请求参数
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(err.Error()))
		return
	}
	defer r.Body.Close()

	var request BroadcastRequest
	if err = json.Unmarshal(data, &request); err != nil || len(request.Message) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	// 广播消息
	var jsb []byte
	model := models.SubscriberModel{}
	subscribers, err := model.GetSubscribers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, userID := range subscribers {
		pusher.Post(userID, request.Message, true, nil)
	}

	respone := BroadcastRespone{OK: true}
	jsb, err = json.Marshal(&respone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsb)
}
