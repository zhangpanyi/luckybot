package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/future"
	"github.com/zhangpanyi/luckymoney/app/logic/scriptengine"
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
	account string // 账户名
	amount  uint32 // 资产数量
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
		info.amount = uint32(amount * 100)
		handler.replyEnterAccout(bot, r, info, update, true)
		return
	}

	// 处理提现总览
	result = reMathWithdrawAccout.FindStringSubmatch(data)
	if len(result) == 3 {
		amount, _ := strconv.ParseFloat(result[1], 10)
		info.amount = uint32(amount * 100)
		info.account = result[2]
		handler.replyWithdrawOverview(bot, r, info, update, true)
		return
	}

	// 处理提现请求
	result = reMathWithdrawSubmit.FindStringSubmatch(data)
	if len(result) == 3 {
		amount, _ := strconv.ParseFloat(result[1], 10)
		info.amount = uint32(amount * 100)
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
	var balance uint32
	serverCfg := config.GetServe()
	model := models.AccountModel{}
	account, err := model.GetAccount(fromID, serverCfg.Symbol)
	if err == nil {
		balance = account.Amount
	}

	// 检查输入金额
	fee := serverCfg.WithdrawFee
	result := strings.Split(amount, ".")
	if len(result) == 2 && len(result[1]) > 2 {
		reply := tr(fromID, "lng_withdraw_amount_not_enough")
		handlerError(fmt.Sprintf(reply, fmt.Sprintf("%.2f", float64(balance)/100),
			serverCfg.Symbol, fee, serverCfg.Symbol))
		return
	}

	fAmount, err := strconv.ParseFloat(amount, 10)
	if err != nil {
		reply := tr(fromID, "lng_withdraw_amount_not_enough")
		handlerError(fmt.Sprintf(reply, fmt.Sprintf("%.2f", float64(balance)/100),
			serverCfg.Symbol, fee, serverCfg.Symbol))
		return
	}

	// 检查用户余额
	lAmount := uint32(fAmount * 100)
	if err != nil || account.Amount < (lAmount+uint32(fee*100)) {
		reply := tr(fromID, "lng_withdraw_amount_error")
		handlerError(fmt.Sprintf(reply, fmt.Sprintf("%.2f", float64(balance)/100),
			serverCfg.Symbol, fee, serverCfg.Symbol))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.amount = lAmount
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
	if !scriptengine.Engine.ValidAccount(account) {
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
	reply = fmt.Sprintf(reply, fmt.Sprintf("%.2f", float64(info.amount)/100.0), serverCfg.Symbol, serverCfg.Name)
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
	fee := serverCfg.WithdrawFee
	reply := tr(fromID, "lng_withdraw_overview")
	amount := fmt.Sprintf("%.2f", float64(info.amount)/100.0)
	reply = fmt.Sprintf(reply, info.account, serverCfg.Symbol, serverCfg.Symbol,
		amount, fee, serverCfg.Symbol, fee, serverCfg.Symbol)

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
	fee := serverCfg.WithdrawFee

	// 扣除余额
	model := models.AccountModel{}
	account, err := model.LockAccount(fromID, serverCfg.Symbol, info.amount+uint32(fee*100))
	if err != nil {
		logger.Warnf("Failed to withdraw asset, UserID: %d, Asset: %s, Amount: %d, Fee: %.2f, %v",
			fromID, serverCfg.Symbol, info.amount, fee, err)
		reply := tr(fromID, "lng_withdraw_not_enough")
		bot.AnswerCallbackQuery(query, reply, false, "", 0)
		bot.EditMessageReplyMarkup(query.Message, reply, false, markup)
		return
	}
	logger.Errorf("Withdraw asset success, UserID: %d, Asset: %s, Amount: %d, Fee: %.2f",
		fromID, serverCfg.Symbol, info.amount, fee)

	// 提交成功
	reply := tr(fromID, "lng_withdraw_submit_ok")
	answer := tr(fromID, "lng_withdraw_submit_ok_answer")
	bot.AnswerCallbackQuery(query, answer, false, "", 0)
	bot.EditMessageReplyMarkup(query.Message, reply, true, nil)

	// 记录账户历史
	versionModel := models.AccountVersionModel{}
	versionModel.InsertVersion(fromID, &models.Version{
		Symbol:     serverCfg.Symbol,
		Locked:     int32(info.amount),
		Fee:        uint32(fee * 100),
		Amount:     account.Amount,
		Reason:     models.ReasonWithdraw,
		RefAddress: &info.account,
	})

	// 执行提现操作
	f := future.Manager.NewFuture()
	amount := strconv.FormatFloat(float64(info.amount)/100, 'f', 2, 64)
	go scriptengine.Engine.OnWithdraw(info.account, serverCfg.Symbol, amount, f.ID())
	if err = f.GetResult(); err != nil {
		reply := tr(fromID, "lng_withdraw_transfer_error")
		logger.Warnf("Failed to transfer, UserID: %d, Asset: %s, Amount: %d, Fee: %.2f, %v",
			fromID, serverCfg.Symbol, info.amount, fee, err)
		bot.EditMessageReplyMarkup(query.Message, reply, false, markup)
		return
	}

	// 返回处理结果
	reply = tr(fromID, "lng_withdraw_success")
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
	return
}
