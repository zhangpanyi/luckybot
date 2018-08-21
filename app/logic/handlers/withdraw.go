package handlers

import (
	"fmt"
	"regexp"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/storage/models"
)

// 匹配金额
var reMathWithdrawAmount *regexp.Regexp

// 匹配账户
var reMathWithdrawAccout *regexp.Regexp

// 匹配提交
var reMathWithdrawSubmit *regexp.Regexp

func init() {
	var err error
	reMathWithdrawAmount, err = regexp.Compile("^/withdraw/([0-9]+\\.?[0-9]*)/$")
	if err != nil {
		panic(err)
	}

	reMathWithdrawAccout, err = regexp.Compile("^/withdraw/([0-9]+\\.?[0-9]*)/(\\w+)/$")
	if err != nil {
		panic(err)
	}

	reMathWithdrawSubmit, err = regexp.Compile("^/withdraw([0-9]+\\.?[0-9]*)/([\\w|-]+)/submit/$")
	if err != nil {
		panic(err)
	}
}

// 取款
type WithdrawHandler struct {
}

// 取款信息
type withdrawInfo struct {
	account string // 账户名
	asset   string // 资产类型
	amount  uint32 // 资产数量
}

// 消息处理
func (handler *WithdrawHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	// 回复输入金额
	data := update.CallbackQuery.Data
	if data == "/withdraw/" {
		handler.replyEnterWithdrawAmount(bot, r, update)
	}
}

// 消息路由
func (handler *WithdrawHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}

// 处理输入提现金额
func (handler *WithdrawHandler) handleEnterWithdrawAmount(bot *methods.BotExt, r *history.History,
	update *types.Update, amount string) {

	// // 处理错误
	// query := update.CallbackQuery
	// fromID := query.From.ID
	// data := query.Data
	// handlerError := func(reply string) {
	// 	r.Pop()
	// 	menus := [...]methods.InlineKeyboardButton{
	// 		methods.InlineKeyboardButton{
	// 			Text:         tr(fromID, "lng_back_superior"),
	// 			CallbackData: backSuperior(data),
	// 		},
	// 	}
	// 	bot.AnswerCallbackQuery(query, "", false, "", 0)
	// 	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
	// 	bot.SendMessage(fromID, reply, true, markup)
	// }

	// // 获取资产信息
	// bitCNY := getUserAssetAmount(fromID, storage.BitCNYSymbol)
	// bitUSD := getUserAssetAmount(fromID, storage.BitUSDSymbol)

	// // 检查输入金额
	// result := strings.Split(amount, ".")
	// if len(result) == 2 && len(result[1]) > 2 {
	// 	fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
	// 	reply := tr(fromID, "lng_priv_withdraw_amount_not_enough")
	// 	handlerError(fmt.Sprintf(reply, info.asset, bitCNY, bitUSD,
	// 		fmt.Sprintf("%.2f", float64(fee)/100.0), info.asset))
	// 	return
	// }

	// fAmount, err := strconv.ParseFloat(amount, 10)
	// if err != nil {
	// 	fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
	// 	reply := tr(fromID, "lng_priv_withdraw_amount_not_enough")
	// 	handlerError(fmt.Sprintf(reply, info.asset, bitCNY, bitUSD,
	// 		fmt.Sprintf("%.2f", float64(fee)/100.0), info.asset))
	// 	return
	// }

	// // 检查用户余额
	// lAmount := uint32(fAmount * 100)
	// newHandler := storage.AssetStorage{}
	// fee, _ := syncfee.GetFee(storage.GetAssetSymbol(info.asset))
	// asset, err := newHandler.GetAsset(fromID, storage.GetAssetSymbol(info.asset))
	// if err != nil || asset.Amount < (lAmount+fee) {
	// 	reply := tr(fromID, "lng_priv_withdraw_amount_error")
	// 	handlerError(fmt.Sprintf(reply, info.asset, bitCNY, bitUSD,
	// 		fmt.Sprintf("%.2f", float64(fee)/100.0), info.asset))
	// 	return
	// }

	// // 更新下个操作状态
	// r.Clear()
	// info.amount = lAmount
	// update.CallbackQuery.Data = data + amount + "/"
	// handler.handleWithdrawAccout(bot, r, info, update, false)
}

// 回复输入提现金额
func (handler *WithdrawHandler) replyEnterWithdrawAmount(bot *methods.BotExt, r *history.History, update *types.Update) {
	// 处理输入金额
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterWithdrawAmount(bot, r, update, back.Message.Text)
		return
	}

	// 提示输入提现金额
	r.Clear().Push(update)
	query := update.CallbackQuery
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: "/main/",
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	// 获取账户余额
	var balance uint32
	serverCfg := config.GetServe()
	model := models.AccountModel{}
	account, err := model.GetAccount(fromID, serverCfg.Symbol)
	if err == nil {
		balance = account.Amount
	}

	// 回复提现操作提示
	fee := serverCfg.WithdrawFee
	reply := tr(fromID, "lng_withdraw_enter_amount")
	reply = fmt.Sprintf(reply, fmt.Sprintf("%.2f", float64(balance)/100),
		serverCfg.Symbol, fee, serverCfg.Symbol)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)

	answer := tr(fromID, "lng_withdraw_enter_amount_answer")
	answer = fmt.Sprintf(answer, serverCfg.Symbol)
	bot.AnswerCallbackQuery(query, answer, false, "", 0)
}
