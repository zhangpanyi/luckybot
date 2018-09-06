package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 获取动作请求
type GetActionsRequest struct {
	UserID int64 `json:"user_id"` // 用户ID
	Offset uint  `json:"offset"`  // 偏移量
	Limit  uint  `json:"limit"`   // 返回数量
}

// 获取动作响应
type GetActionsRespone struct {
	Sum     int               `json:"sum"`     // 动作总量
	Count   int               `json:"count"`   // 返回数量
	Actions []*models.Version `json:"actions"` // 动作列表
}

// 获取用户操作
func GetActions(w http.ResponseWriter, r *http.Request) {
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

	var request GetActionsRequest
	if err = json.Unmarshal(data, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	// 查询用户历史
	model := models.AccountVersionModel{}
	actions, sum, err := model.GetVersions(request.UserID, request.Offset, request.Limit, true)
	if err != nil {
		logger.Warnf("Failed to query user actions, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}
	respone := GetActionsRespone{Sum: sum, Count: len(actions), Actions: actions}
	jsb, err := json.Marshal(respone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	// 返回余额信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsb)
}
