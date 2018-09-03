package handlers

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckybot/app/config"
	"github.com/zhangpanyi/luckybot/app/fmath"
	"github.com/zhangpanyi/luckybot/app/future"
	"github.com/zhangpanyi/luckybot/app/logic/scriptengine"
	"github.com/zhangpanyi/luckybot/app/storage/models"
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

	reMathWithdrawSubmit, err = regexp.Compile("^/withdraw/([0-9]+\\.?[0-9]*)/([\\w|-]+)/submit/$")
	if err != nil {
		panic(err)
	}
}

// 取款
type WithdrawHandler struct {
}

// 取款信息
type withdrawInfo struct {
	account string  // 账户名
	amount  float64 // 资产数量
}

// 消息处理
func (handler *WithdrawHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	// 回复输入金额
	info := new(withdrawInfo)
	data := update.CallbackQuery.Data
	if data == "/withdraw/" {
		handler.replyEnterWithdrawAmount(bot, r, info, update)
	}

	// 处理输入账户名
	result := reMathWithdrawAmount.FindStringSubmatch(data)
	if len(result) == 2 {
		amount, _ := strconv.ParseFloat(result[1], 10)
		info.amount = amount
		handler.replyEnterAccout(bot, r, info, update, true)
		return
	}

	// 处理提现总览
	result = reMathWithdrawAccout.FindStringSubmatch(data)
	if len(result) == 3 {
		amount, _ := strconv.ParseFloat(result[1], 10)
		info.amount = amount
		info.account = result[2]
		handler.replyWithdrawOverview(bot, r, info, update, true)
		return
	}

	// 处理提现请求
	result = reMathWithdrawSubmit.FindStringSubmatch(data)
	if len(result) == 3 {
		amount, _ := strconv.ParseFloat(result[1], 10)
		info.amount = amount
		info.account = result[2]
		handler.handleWithdraw(bot, r, info, update.CallbackQuery)
		return
	}
}

// 消息路由
func (handler *WithdrawHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}

// 处理输入提现金额
func (handler *WithdrawHandler) handleEnterWithdrawAmount(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	update *types.Update, amount string) {

	// 错误处理
	query := update.CallbackQuery
	fromID := query.From.ID
	data := query.Data
	handlerError := func(reply string) {
		r.Pop()
		menus := [...]methods.InlineKeyboardButton{
			methods.InlineKeyboardButton{
				Text:         tr(fromID, "lng_back_superior"),
				CallbackData: "/main/",
			},
		}
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 获取账户余额
	balance := big.NewFloat(0)
	serverCfg := config.GetServe()
	model := models.AccountModel{}
	account, err := model.GetAccount(fromID, serverCfg.Symbol)
	if err == nil {
		balance = account.Amount
	}

	// 检查输入金额
	result := strings.Split(amount, ".")
	fee := big.NewFloat(serverCfg.WithdrawFee)
	if len(result) == 2 && len(result[1]) > 2 {
		reply := tr(fromID, "lng_withdraw_amount_not_enough")
		handlerError(fmt.Sprintf(reply, balance.String(),
			serverCfg.Symbol, fee.String(), serverCfg.Symbol))
		return
	}

	fAmount, err := strconv.ParseFloat(amount, 10)
	if err != nil {
		reply := tr(fromID, "lng_withdraw_amount_not_enough")
		handlerError(fmt.Sprintf(reply, balance.String(),
			serverCfg.Symbol, fee.String(), serverCfg.Symbol))
		return
	}

	// 检查用户余额
	if err != nil || account.Amount.Cmp(fmath.Add(big.NewFloat(fAmount), fee)) == -1 {
		reply := tr(fromID, "lng_withdraw_amount_error")
		handlerError(fmt.Sprintf(reply, balance.String(),
			serverCfg.Symbol, fee.String(), serverCfg.Symbol))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.amount = fAmount
	update.CallbackQuery.Data = data + amount + "/"
	handler.replyEnterAccout(bot, r, info, update, false)
}

// 回复输入提现金额
func (handler *WithdrawHandler) replyEnterWithdrawAmount(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	update *types.Update) {

	// 处理输入金额
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterWithdrawAmount(bot, r, info, update, back.Message.Text)
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
	balance := big.NewFloat(0)
	serverCfg := config.GetServe()
	model := models.AccountModel{}
	account, err := model.GetAccount(fromID, serverCfg.Symbol)
	if err == nil {
		balance = account.Amount
	}

	// 回复提现操作提示
	fee := big.NewFloat(serverCfg.WithdrawFee)
	reply := tr(fromID, "lng_withdraw_enter_amount")
	reply = fmt.Sprintf(reply, balance.String(),
		serverCfg.Symbol, fee.String(), serverCfg.Symbol)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)

	answer := tr(fromID, "lng_withdraw_enter_amount_answer")
	answer = fmt.Sprintf(answer, serverCfg.Symbol)
	bot.AnswerCallbackQuery(query, answer, false, "", 0)
}

// 处理输入账户名
func (handler *WithdrawHandler) handleEnterWithdrawAccout(bot *methods.BotExt, r *history.History,
	info *withdrawInfo, update *types.Update, account string) {

	// 处理错误
	query := update.CallbackQuery
	fromID := query.From.ID
	data := query.Data
	handlerError := func(reply string) {
		r.Pop()
		menus := [...]methods.InlineKeyboardButton{
			methods.InlineKeyboardButton{
				Text:         tr(fromID, "lng_back_superior"),
				CallbackData: backSuperior(data),
			},
		}
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 检查帐号合法
	if !scriptengine.Engine.ValidAddress(account) {
		handlerError(tr(fromID, "lng_withdraw_account_error"))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.account = account
	update.CallbackQuery.Data = data + account + "/"
	handler.replyWithdrawOverview(bot, r, info, update, false)
}

// 回复输入账户名
func (handler *WithdrawHandler) replyEnterAccout(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	update *types.Update, edit bool) {

	// 处理输入金额
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterWithdrawAccout(bot, r, info, update, back.Message.Text)
		return
	}

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: backSuperior(query.Data),
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	// 回复请求结果
	r.Clear().Push(update)
	serverCfg := config.GetServe()
	reply := tr(fromID, "lng_withdraw_enter_account")
	reply = fmt.Sprintf(reply, big.NewFloat(info.amount).String(), serverCfg.Symbol, serverCfg.Name)
	if !edit {
		bot.SendMessage(fromID, reply, true, markup)
	} else {
		bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
	}

	answer := tr(fromID, "lng_withdraw_enter_account_answer")
	bot.AnswerCallbackQuery(query, fmt.Sprintf(answer, info.account), false, "", 0)
}

// 处理提现概览
func (handler *WithdrawHandler) replyWithdrawOverview(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	update *types.Update, edit bool) {

	// 应答请求
	fromID := update.CallbackQuery.From.ID
	answer := tr(fromID, "lng_withdraw_overview_answer")
	bot.AnswerCallbackQuery(update.CallbackQuery, answer, false, "", 0)

	// 格式化信息
	serverCfg := config.GetServe()
	amount := big.NewFloat(info.amount)
	fee := big.NewFloat(serverCfg.WithdrawFee)
	reply := tr(fromID, "lng_withdraw_overview")
	reply = fmt.Sprintf(reply, info.account, amount.String(), serverCfg.Symbol,
		amount.String(), fee.String(), serverCfg.Symbol, fee.String(), serverCfg.Symbol)

	// 生成菜单按钮
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_withdraw_submit"),
			CallbackData: update.CallbackQuery.Data + "submit/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: backSuperior(update.CallbackQuery.Data),
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	// 回复信息总览
	if !edit {
		bot.SendMessage(fromID, reply, true, markup)
	} else {
		bot.EditMessageReplyMarkup(update.CallbackQuery.Message, reply, true, markup)
	}
}

// 处理提现
func (handler *WithdrawHandler) handleWithdraw(bot *methods.BotExt, r *history.History, info *withdrawInfo,
	query *types.CallbackQuery) {

	// 生成菜单
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_menu"),
			CallbackData: "/main/",
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	// 获取手续费
	serverCfg := config.GetServe()
	fee := big.NewFloat(serverCfg.WithdrawFee)

	// 扣除余额
	model := models.AccountModel{}
	amount := big.NewFloat(info.amount)
	account, err := model.LockAccount(fromID, serverCfg.Symbol, fmath.Add(amount, fee))
	if err != nil {
		logger.Warnf("Failed to withdraw, user: %d, asset: %s, amount: %s, fee: %s, %v",
			fromID, serverCfg.Symbol, amount.String(), fee.String(), err)
		reply := tr(fromID, "lng_withdraw_not_enough")
		bot.AnswerCallbackQuery(query, reply, false, "", 0)
		bot.EditMessageReplyMarkup(query.Message, reply, false, markup)
		return
	}
	logger.Errorf("Withdraw submitted, user: %d, asset: %s, amount: %s, fee: %s",
		fromID, serverCfg.Symbol, amount.String(), fee.String())

	// 提交成功
	reply := tr(fromID, "lng_withdraw_submit_ok")
	answer := tr(fromID, "lng_withdraw_submit_ok_answer")
	bot.AnswerCallbackQuery(query, answer, false, "", 0)
	bot.EditMessageReplyMarkup(query.Message, reply, true, nil)

	// 记录账户历史
	versionModel := models.AccountVersionModel{}
	versionModel.InsertVersion(fromID, &models.Version{
		Symbol:     serverCfg.Symbol,
		Locked:     amount,
		Fee:        fee,
		Amount:     account.Amount,
		Reason:     models.ReasonWithdraw,
		RefAddress: &info.account,
	})

	// 开始转账提示
	reply = tr(fromID, "lng_withdraw_agreed")
	bot.EditMessageReplyMarkup(query.Message, reply, true, nil)

	// 执行提现操作
	f := future.NewFuture()
	zero := big.NewFloat(0)
	go scriptengine.Engine.OnWithdraw(info.account, serverCfg.Symbol, amount.String(), f)
	txid, err := f.GetResult()
	if err != nil {
		// 解锁资产
		account, err := model.UnlockAccount(fromID, serverCfg.Symbol, fmath.Add(amount, fee))
		if err == nil {
			versionModel.InsertVersion(fromID, &models.Version{
				Symbol:     serverCfg.Symbol,
				Locked:     fmath.Sub(zero, amount),
				Fee:        fee,
				Amount:     account.Amount,
				Reason:     models.ReasonWithdrawFailure,
				RefAddress: &info.account,
			})
		} else {
			logger.Warnf(`Failed to unlock account when withdraw failure, user: %d, asset: %s, \
			amount: %s, fee: %s, %v`, fromID, serverCfg.Symbol, amount.String(), fee.String(), err)
		}

		// 返回信息
		reply := tr(fromID, "lng_withdraw_transfer_error")
		logger.Warnf("Failed to transfer, user: %d, asset: %s, amount: %s, fee: %s, %v",
			fromID, serverCfg.Symbol, amount.String(), fee.String(), err)
		bot.EditMessageReplyMarkup(query.Message, reply, false, markup)
		return
	}

	// 记录提现成功
	account, err = model.Withdraw(fromID, serverCfg.Symbol, fmath.Add(amount, fee))
	if err == nil {
		versionModel.InsertVersion(fromID, &models.Version{
			Symbol:     serverCfg.Symbol,
			Balance:    fmath.Sub(zero, fmath.Add(amount, fee)),
			Locked:     fmath.Sub(zero, amount),
			Fee:        fee,
			Amount:     account.Amount,
			Reason:     models.ReasonWithdrawSuccess,
			RefAddress: &info.account,
			RefTxID:    &txid,
		})
	} else {
		logger.Warnf(`Failed to unlock account when withdraw success, user: %d, asset: %s, \
		amount: %s, fee: %s, %v`, fromID, serverCfg.Symbol, amount.String(), fee.String(), err)
	}

	// 返回处理结果
	reply = tr(fromID, "lng_withdraw_success")
	bot.EditMessageReplyMarkup(query.Message, fmt.Sprintf(reply, txid), true, markup)
	return
}
