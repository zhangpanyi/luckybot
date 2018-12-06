package handlers

import (
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/zhangpanyi/luckybot/app/config"
	"github.com/zhangpanyi/luckybot/app/storage"
	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 获取余额请求
type GetBalanceRequest struct {
	UserID int64 `json:"user_id"` // 用户ID
	Tonce  int64 `json:"tonce"`   // 时间戳
}

// 获取余额响应
type GetBalanceRespone struct {
	Amount *big.Float `json:"amount"` // 可用余额
	Locked *big.Float `json:"locked"` // 锁定金额
}

// 获取余额
func GetBalance(w http.ResponseWriter, r *http.Request) {
	// 验证权限
	sessionID, data, ok := authentication(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(makeErrorRespone("", ""))
		return
	}

	// 解析请求参数
	var request GetBalanceRequest
	if err := json.Unmarshal(data, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 获取账户余额
	serveCfg := config.GetServe()
	model := models.AccountModel{}
	account, err := model.GetAccount(request.UserID, serveCfg.Symbol)
	if err != nil && err != storage.ErrNoBucket {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	respone := GetBalanceRespone{Amount: account.Amount, Locked: account.Locked}
	jsb, err := json.Marshal(respone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(sessionID, err.Error()))
		return
	}

	// 返回余额信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(makeRespone(sessionID, jsb))
}
