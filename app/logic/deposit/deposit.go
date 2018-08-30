package deposit

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/zhangpanyi/luckymoney/app/storage/models"

	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/luckymoney/app/logic/scriptengine"
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

// 充值处理
func HandleDeposit(w http.ResponseWriter, r *http.Request) {
	// 读取数据
	defer r.Body.Close()
	jsb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 解析数据
	var request DepositRequest
	if err = json.Unmarshal(jsb, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 检查重复充值
	depositModel := models.DepositModel{}
	if depositModel.Exist(request.TxID) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 充值是否有效
	ok := scriptengine.Engine.ValidTransaction(request.TxID, request.From, request.To,
		request.Asset, request.Amount, request.Memo)
	if !ok {
		logger.Infof("Failed to deposit, invalid transaction, txid: %s, from: %s, to: %s, asset: %s, amount: %s, memo: %s",
			request.TxID, request.From, request.To, request.Asset, request.Amount, request.Memo)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 获取充值金额
	amount, ok := big.NewFloat(0).SetString(request.Amount)
	if !ok {
		logger.Infof("Failed to deposit, amount invalid, amount: %s", request.Amount)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 获取用户ID
	userID, err := strconv.ParseInt(request.Memo, 10, 64)
	if err != nil {
		logger.Infof("Failed to deposit, not found user id from memo, memo: %s, %v", request.Memo, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 防止重复充值
	if err = depositModel.Add(request.TxID, jsb); err != nil {
		logger.Infof("Failed to deposit, txid: %s, from: %s, to: %s, asset: %s, amount: %s, memo: %s, %v",
			request.TxID, request.From, request.To, request.Asset, request.Amount, request.Memo, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 增加用户资产
	model := models.AccountModel{}
	account, err := model.Deposit(userID, request.Asset, amount)
	if err != nil {
		logger.Infof("Failed to deposit, txid: %s, from: %s, to: %s, asset: %s, amount: %s, memo: %s, %v",
			request.TxID, request.From, request.To, request.Asset, request.Amount, request.Memo, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 写入充值记录
	versionModel := models.AccountVersionModel{}
	versionModel.InsertVersion(userID, &models.Version{
		Symbol:         request.Asset,
		Balance:        amount,
		Amount:         account.Amount,
		Reason:         models.ReasonDeposit,
		RefTxID:        &request.TxID,
		RefBlockHeight: &request.Height,
	})

	w.WriteHeader(http.StatusOK)
}
