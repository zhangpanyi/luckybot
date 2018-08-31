package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zhangpanyi/luckymoney/app/storage/models"
)

// 获取订户
func Subscribers(w http.ResponseWriter, r *http.Request) {
	// 验证权限
	if !authentication(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 查询订阅用户
	model := models.SubscriberModel{}
	users, err := model.GetSubscribers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	// 返回处理结果
	jsb, err := json.Marshal(&users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsb)
}
