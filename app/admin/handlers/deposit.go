package handlers

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/zhangpanyi/luckybot/app/config"
	"github.com/zhangpanyi/luckybot/app/fmath"
	"github.com/zhangpanyi/luckybot/app/logic/handlers/utils"
	"github.com/zhangpanyi/luckybot/app/logic/pusher"
	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 充值请求
type DepositRequest struct {
	UserID int64      `json:"user_id"` // 用户ID
	Amount *big.Float `json:"amount"`  // 充值金额
}

// 充值响应
type DepositRespone struct {
	Amount *big.Float `json:"amount"` // 可用余额
	Locked *big.Float `json:"locked"` // 锁定金额
}

// 充值资产
func Deposit(w http.ResponseWriter, r *http.Request) {
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

	var request DepositRequest
	if err = json.Unmarshal(data, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	zero := big.NewFloat(0)
	if request.Amount != nil {
		request.Amount.SetPrec(fmath.Prec())
	}
	if request.Amount == nil || request.Amount.Cmp(zero) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone("amount must be greater than 0"))
		return
	}

	// 为用户充值
	serveCfg := config.GetServe()
	model := models.AccountModel{}
	account, err := model.Deposit(request.UserID, serveCfg.Symbol, request.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	respone := DepositRespone{Amount: account.Amount, Locked: account.Locked}
	jsb, err := json.Marshal(respone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(err.Error()))
		return
	}

	// 写入操作记录
	versionModel := models.AccountVersionModel{}
	version, err := versionModel.InsertVersion(request.UserID, &models.Version{
		Symbol:  serveCfg.Symbol,
		Balance: request.Amount,
		Amount:  account.Amount,
		Reason:  models.ReasonSystem,
	})

	// 推送充值通知
	if err == nil {
		pusher.Post(request.UserID, utils.MakeHistoryMessage(request.UserID, version), true, nil)
	}

	// 返回余额信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsb)
}
