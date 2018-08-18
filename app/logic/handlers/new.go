package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/logic/algo"
	"github.com/zhangpanyi/luckymoney/app/storage/models"
)

// 匹配类型
var reMathType *regexp.Regexp

// 匹配金额
var reMathAmount *regexp.Regexp

// 匹配数量
var reMathNumber *regexp.Regexp

func init() {
	var err error
	reMathType, err = regexp.Compile("^/new/(rand|equal)/$")
	if err != nil {
		panic(err)
	}

	reMathAmount, err = regexp.Compile("^/new/(rand|equal)/([0-9]+\\.?[0-9]*)/$")
	if err != nil {
		panic(err)
	}

	reMathNumber, err = regexp.Compile("^/new/(rand|equal)/([0-9]+\\.?[0-9]*)/(\\d+)/$")
	if err != nil {
		panic(err)
	}
}

var (
	// 随机红包
	randLuckyMoney = "rand"
	// 普通红包
	equalLuckyMoney = "equal"
)

// 红包信息
type luckyMoneys struct {
	typ     string // 红包类型
	amount  uint32 // 红包金额
	number  uint32 // 红包个数
	message string // 红包留言
}

// 红包类型转字符串
func luckyMoneysTypeToString(fromID int64, typ string) string {
	if typ == randLuckyMoney {
		return tr(fromID, "lng_new_rand")
	}
	return tr(fromID, "lng_new_equal")
}

// 创建红包
type NewHandler struct {
}

// 消息处理
func (handler *NewHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	// 回复选择红包类型
	data := update.CallbackQuery.Data
	if data == "/new/" {
		r.Clear()
		handler.replyChooseType(bot, update.CallbackQuery)
		return
	}

	// 回复输入红包金额
	info := luckyMoneys{}
	result := reMathType.FindStringSubmatch(data)
	if len(result) == 2 {
		info.typ = result[1]
		handler.replyEnterAmount(bot, r, &info, update)
		return
	}

	// 回复输入红包数量
	result = reMathAmount.FindStringSubmatch(data)
	if len(result) == 3 {
		info.typ = result[1]
		amount, _ := strconv.ParseFloat(result[2], 10)
		info.amount = uint32(amount * 100)
		handler.replyEnterNumber(bot, r, &info, update, true)
		return
	}

	// 回复输入红包留言
	result = reMathNumber.FindStringSubmatch(data)
	if len(result) == 4 {
		info.typ = result[1]
		amount, _ := strconv.ParseFloat(result[2], 10)
		info.amount = uint32(amount * 100)
		number, _ := strconv.Atoi(result[3])
		info.number = uint32(number)
		handler.replyEnterMessage(bot, r, &info, update)
		return
	}

	// 路由到其它处理模块
	newHandler := handler.route(bot, update.CallbackQuery)
	if newHandler == nil {
		return
	}
	newHandler.Handle(bot, r, update)
}

// 消息路由
func (handler *NewHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}

// 返回上级
func backSuperior(data string) string {
	s := strings.Split(data, "/")
	if len(s) <= 2 {
		return "/main/"
	}
	return strings.Join(s[:len(s)-2], "/") + "/"
}

// 生成基本菜单
func makeBaseMenus(fromID int64, data string) *methods.InlineKeyboardMarkup {
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_new_cancel"),
			CallbackData: "/main/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: backSuperior(data),
		},
	}
	return methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
}

// 回复输入选择类型
func (handler *NewHandler) replyChooseType(bot *methods.BotExt, query *types.CallbackQuery) {

	// 生成菜单列表
	data := query.Data
	fromID := query.From.ID
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_new_rand"),
			CallbackData: data + randLuckyMoney + "/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_new_equal"),
			CallbackData: data + equalLuckyMoney + "/",
		},
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: "/main/",
		},
	}

	// 回复请求结果
	reply := tr(fromID, "lng_new_choose_type")
	markup := methods.MakeInlineKeyboardMarkup(menus[:], 2, 1)
	bot.AnswerCallbackQuery(query, "", false, "", 0)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
}

// 处理输入红包金额
func (handler *NewHandler) handleEnterAmount(bot *methods.BotExt, r *history.History,
	info *luckyMoneys, update *types.Update, enterAmount string) {

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID

	// 处理错误
	data := query.Data
	handlerError := func(reply string) {
		r.Pop()
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := makeBaseMenus(fromID, query.Data)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 检查输入金额
	amount, err := strconv.ParseFloat(enterAmount, 10)
	if err != nil || amount < 0.01 {
		handlerError(tr(fromID, "lng_new_set_amount_error"))
		return
	}

	// 检查小数点位数
	s := strings.Split(enterAmount, ".")
	if len(s) == 2 && len(s[1]) > 2 {
		handlerError(tr(fromID, "lng_new_set_amount_error"))
		return
	}

	// 检查帐户余额
	serveCfg := config.GetServe()
	balance := getUserAssetAmount(fromID, serveCfg.Symbol)
	fBalance, _ := strconv.ParseFloat(balance, 10)
	if amount > fBalance {
		reply := tr(fromID, "lng_new_set_amount_no_asset")
		handlerError(fmt.Sprintf(reply, serveCfg.Symbol, balance))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.amount = uint32(amount * 100)
	update.CallbackQuery.Data = data + enterAmount + "/"
	handler.replyEnterNumber(bot, r, info, update, false)
}

// 回复输入红包金额
func (handler *NewHandler) replyEnterAmount(bot *methods.BotExt, r *history.History, info *luckyMoneys,
	update *types.Update) {

	// 处理输入金额
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterAmount(bot, r, info, update, back.Message.Text)
		return
	}

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID
	markup := makeBaseMenus(fromID, query.Data)

	// 回复请求结果
	r.Clear().Push(update)
	amountDesc := tr(fromID, "lng_new_total_amount")
	if info.typ == equalLuckyMoney {
		amountDesc = tr(fromID, "lng_new_unit_amount")
	}

	answer := fmt.Sprintf(tr(fromID, "lng_new_set_amount_answer"), amountDesc)
	bot.AnswerCallbackQuery(query, answer, false, "", 0)

	serveCfg := config.GetServe()
	reply := tr(fromID, "lng_new_set_amount")
	amount := getUserAssetAmount(fromID, serveCfg.Symbol)
	reply = fmt.Sprintf(reply, amountDesc, luckyMoneysTypeToString(fromID, info.typ),
		serveCfg.Symbol, amount)
	bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
}

// 处理输入红包个数
func (handler *NewHandler) handleEnterNumber(bot *methods.BotExt, r *history.History,
	info *luckyMoneys, update *types.Update, enterNumber string) {

	// 生成菜单列表
	query := update.CallbackQuery
	fromID := query.From.ID

	// 处理错误
	handlerError := func(reply string) {
		r.Pop()
		markup := makeBaseMenus(fromID, query.Data)
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		bot.SendMessage(fromID, reply, true, markup)
	}

	// 检查红包数量
	number, err := strconv.ParseUint(enterNumber, 10, 32)
	if err != nil {
		handlerError(tr(fromID, "lng_new_set_number_error"))
		return
	}

	// 检查账户余额
	serveCfg := config.GetServe()
	balance := getUserAssetAmount(fromID, serveCfg.Symbol)
	if info.typ == randLuckyMoney && uint32(number) > info.amount {
		reply := tr(fromID, "lng_new_set_number_not_enough")
		handlerError(fmt.Sprintf(reply, serveCfg.Symbol, balance))
		return
	}

	fBalance, _ := strconv.ParseFloat(balance, 10)
	if info.typ == equalLuckyMoney && (info.amount*uint32(number) > uint32(fBalance*100)) {
		reply := tr(fromID, "lng_new_set_number_not_enough")
		handlerError(fmt.Sprintf(reply, serveCfg.Symbol, balance))
		return
	}

	// 更新下个操作状态
	r.Clear()
	info.number = uint32(number)
	update.CallbackQuery.Data += enterNumber + "/"
	handler.replyEnterMessage(bot, r, info, update)
}

// 回复输入红包数量
func (handler *NewHandler) replyEnterNumber(bot *methods.BotExt, r *history.History, info *luckyMoneys,
	update *types.Update, edit bool) {

	// 处理输入个数
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterNumber(bot, r, info, update, back.Message.Text)
		return
	}

	// 提示输入红包个数
	r.Clear().Push(update)
	query := update.CallbackQuery
	fromID := query.From.ID
	markup := makeBaseMenus(fromID, query.Data)

	amountDesc := tr(fromID, "lng_new_total_amount")
	if info.typ == equalLuckyMoney {
		amountDesc = tr(fromID, "lng_new_unit_amount")
	}

	reply := ""
	serveCfg := config.GetServe()
	reply = tr(fromID, "lng_new_set_number")
	reply = fmt.Sprintf(reply, luckyMoneysTypeToString(fromID, info.typ),
		amountDesc, fmt.Sprintf("%.2f", float64(info.amount)/100.0), serveCfg.Symbol)

	if !edit {
		bot.SendMessage(fromID, reply, true, markup)
	} else {
		bot.EditMessageReplyMarkup(query.Message, reply, true, markup)
	}
	bot.AnswerCallbackQuery(query, tr(fromID, "lng_new_set_number_answer"), false, "", 0)
}

// 处理输入红包留言
func (handler *NewHandler) handleEnterMessage(bot *methods.BotExt, r *history.History,
	info *luckyMoneys, update *types.Update, message string) {

	// 处理错误
	query := update.CallbackQuery
	fromID := query.From.ID
	handlerError := func(reply string) {
		r.Pop()
		bot.AnswerCallbackQuery(query, "", false, "", 0)
		markup := makeBaseMenus(fromID, query.Data)
		bot.SendMessage(fromID, reply, true, markup)
		return
	}

	// 检查留言长度
	serve := config.GetServe()
	if len(message) == 0 || len(message) > serve.MaxMessageLen {
		reply := fmt.Sprintf(tr(fromID, "lng_new_set_message_error"),
			serve.MaxMessageLen)
		handlerError(reply)
		return
	}

	// 处理生成红包
	info.message = message
	data, err := handler.handleGenerateLuckyMoney(fromID, query.From.FirstName, info)
	if err != nil {
		logger.Warnf("Failed to create lucky money, %v", err)
		handlerError(tr(fromID, "lng_new_failed"))
		return
	}

	// 删除已有键盘
	remove := methods.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}
	bot.SendMessage(fromID, tr(fromID, "lng_new_waiting"), false, &remove)

	// 回复红包内容
	r.Clear()
	reply := tr(fromID, "lng_new_created")
	reply = fmt.Sprintf(reply, bot.UserName)
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:              tr(fromID, "lng_send_luckymoney"),
			SwitchInlineQuery: data.SN,
		},
	}
	bot.AnswerCallbackQuery(query, "", false, "", 0)
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
	bot.SendMessageDisableWebPagePreview(fromID, reply, true, markup)
}

// 回复输入红包留言
func (handler *NewHandler) replyEnterMessage(bot *methods.BotExt, r *history.History, info *luckyMoneys,
	update *types.Update) {

	// 处理输入留言
	back, err := r.Back()
	if err == nil && back.Message != nil {
		handler.handleEnterMessage(bot, r, info, update, back.Message.Text)
		return
	}

	// 生成回复键盘
	query := update.CallbackQuery
	fromID := query.From.ID
	menus := [...]methods.KeyboardButton{
		methods.KeyboardButton{
			Text: tr(fromID, "lng_new_benediction"),
		},
	}
	markup := methods.MakeReplyKeyboardMarkup(menus[:], 1)
	markup.OneTimeKeyboard = true

	// 提示输入红包留言
	r.Clear().Push(update)
	amount := tr(fromID, "lng_new_total_amount")
	if info.typ == equalLuckyMoney {
		amount = tr(fromID, "lng_new_unit_amount")
	}
	serveCfg := config.GetServe()
	reply := tr(fromID, "lng_new_set_message")
	reply = fmt.Sprintf(reply, luckyMoneysTypeToString(fromID, info.typ), serveCfg.Symbol,
		amount, fmt.Sprintf("%.2f", float64(info.amount)/100.0), serveCfg.Symbol, info.number)
	bot.SendMessage(fromID, reply, true, markup)
	bot.AnswerCallbackQuery(query, tr(fromID, "lng_new_set_message_answer"), false, "", 0)
}

// 处理生成红包
func (handler *NewHandler) handleGenerateLuckyMoney(userID int64, firstName string,
	info *luckyMoneys) (*models.LuckyMoney, error) {

	// 锁定资金
	amount := info.amount
	if info.typ == equalLuckyMoney {
		amount = info.amount * info.number
	}
	serveCfg := config.GetServe()
	model := models.AccountModel{}
	err := model.LockAccount(userID, serveCfg.Symbol, amount)
	if err != nil {
		return nil, err
	}
	logger.Errorf("Lock account, user_id: %v, asset: %v, amount: %v",
		userID, serveCfg.Symbol, amount)

	// 生成红包
	var luckyMoneyArr []int
	if info.typ == randLuckyMoney {
		luckyMoneyArr, err = algo.Generate(amount, info.number)
		if err != nil {
			logger.Errorf("Failed to generate lucky money, user_id: %v, %v", userID, err)

			// 解锁资金
			if err = model.LockAccount(userID, serveCfg.Symbol, amount); err != nil {
				logger.Errorf("Failed to unlock asset, user_id: %v, asset: %v, amount: %v",
					userID, serveCfg.Symbol, amount)
			}
			return nil, err
		}
	} else {
		luckyMoneyArr = make([]int, 0, info.number)
		for i := 0; i < int(info.number); i++ {
			luckyMoneyArr = append(luckyMoneyArr, int(info.amount))
		}
	}

	// 保存红包信息
	luckyMoney := models.LuckyMoney{
		SenderID:   userID,
		SenderName: firstName,
		Asset:      serveCfg.Symbol,
		Amount:     info.amount,
		Number:     info.number,
		Message:    info.message,
		Lucky:      info.typ == randLuckyMoney,
		Timestamp:  time.Now().UTC().Unix(),
	}
	if info.typ == equalLuckyMoney {
		luckyMoney.Value = info.amount
	}
	luckyMoneyModel := models.LuckyMoneyModel{}
	data, err := luckyMoneyModel.NewLuckyMoney(&luckyMoney, luckyMoneyArr)
	if err != nil {
		logger.Errorf("Failed to new lucky money, user_id: %v, %v", userID, err)

		// 解冻资金
		if err = model.UnlockAccount(userID, serveCfg.Symbol, amount); err != nil {
			logger.Errorf("Failed to unlock asset, user_id: %v, asset: %v, amount: %v",
				userID, serveCfg.Symbol, amount)
		}
		return nil, err
	}
	logger.Errorf("Generate lucky money, id: %v, user_id: %v, asset: %v, amount: %v",
		data.ID, userID, serveCfg.Symbol, amount)

	return data, nil
}
