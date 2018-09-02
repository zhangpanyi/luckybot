package deposit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/zhangpanyi/luckybot/app/storage/models"

	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/luckybot/app/logic/handlers"
	"github.com/zhangpanyi/luckybot/app/logic/pusher"
	"github.com/zhangpanyi/luckybot/app/logic/scriptengine"
)

// 充值请求
type DepositRequest struct {
	TxID   string `json:"txid"`   // 交易ID
	Height uint64 `json:"heigth"` // 区块高度
	From   string `json:"from"`   // 来源地址
	To     string `json:"to"`     // 目标地址
	Asset  string `json:"asset"`  // 资产名称
	Amount string `json:"amount"` // 充值金额
	Memo   string `json:"memo"`   // 备注信息
}

// 生成错误响应
func makeErrorRespone(reason string) []byte {
	object := map[string]string{
		"error": reason,
	}
	jsb, _ := json.Marshal(&object)
	return jsb
}

// 充值处理
func HandleDeposit(w http.ResponseWriter, r *http.Request) {
	// 读取数据
	defer r.Body.Close()
	jsb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(fmt.Sprintf("invalid body, %v", err)))
		return
	}

	// 解析数据
	var request DepositRequest
	if err = json.Unmarshal(jsb, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone(fmt.Sprintf("invalid request, %v", err)))
		return
	}

	// 检查重复充值
	depositModel := models.DepositModel{}
	if depositModel.Exist(request.TxID) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone("repeat deposit"))
		return
	}

	// 充值是否有效
	ok := scriptengine.Engine.ValidTransaction(request.TxID, request.From, request.To,
		request.Asset, request.Amount, request.Memo)
	if !ok {
		logger.Infof("Failed to deposit, invalid transaction, txid: %s, from: %s, to: %s, asset: %s, amount: %s, memo: %s",
			request.TxID, request.From, request.To, request.Asset, request.Amount, request.Memo)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone("invalid transaction"))
		return
	}

	// 获取充值金额
	amount, ok := big.NewFloat(0).SetString(request.Amount)
	if !ok {
		logger.Infof("Failed to deposit, amount invalid, amount: %s", request.Amount)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone("invalid amount"))
		return
	}

	// 获取用户ID
	userID, err := strconv.ParseInt(request.Memo, 10, 64)
	if err != nil {
		logger.Infof("Failed to deposit, not found user id from memo, memo: %s, %v", request.Memo, err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(makeErrorRespone("not found user"))
		return
	}

	// 防止重复充值
	if err = depositModel.Add(request.TxID, jsb); err != nil {
		logger.Infof("Failed to deposit, txid: %s, from: %s, to: %s, asset: %s, amount: %s, memo: %s, %v",
			request.TxID, request.From, request.To, request.Asset, request.Amount, request.Memo, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone("repeat deposit"))
		return
	}

	// 增加用户资产
	model := models.AccountModel{}
	account, err := model.Deposit(userID, request.Asset, amount)
	if err != nil {
		logger.Infof("Failed to deposit, txid: %s, from: %s, to: %s, asset: %s, amount: %s, memo: %s, %v",
			request.TxID, request.From, request.To, request.Asset, request.Amount, request.Memo, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(makeErrorRespone(fmt.Sprintf("deposit failure, %v", err)))
		return
	}

	// 写入充值记录
	versionModel := models.AccountVersionModel{}
	version, err := versionModel.InsertVersion(userID, &models.Version{
		Symbol:         request.Asset,
		Balance:        amount,
		Amount:         account.Amount,
		Reason:         models.ReasonDeposit,
		RefTxID:        &request.TxID,
		RefBlockHeight: &request.Height,
	})

	// 推送充值通知
	if err == nil {
		pusher.Post(userID, handlers.MakeHistoryMessage(userID, version), true, nil)
	}

	w.WriteHeader(http.StatusOK)
}
