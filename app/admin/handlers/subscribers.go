package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 获取订户请求
type GetSubscribersRequest struct {
	Tonce int64 `json:"tonce"` // 时间戳
}

// 获取订户
func Subscribers(w http.ResponseWriter, r *http.Request) {
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
	var request GetSubscribersRequest
	if err := json.Unmarshal(data, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 查询订阅用户
	model := models.SubscriberModel{}
	users, err := model.GetSubscribers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 返回处理结果
	jsb, err := json.Marshal(&users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(makeRespone(sessionID, jsb))
}
