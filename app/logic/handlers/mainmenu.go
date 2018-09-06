package handlers

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckybot/app/config"
	"github.com/zhangpanyi/luckybot/app/storage"
	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 消息处理器
type Handler interface {
	route(*methods.BotExt, *types.CallbackQuery) Handler
	Handle(*methods.BotExt, *history.History, *types.Update)
}

// 主菜单
type MainMenuHandler struct {
}

// 消息处理
func (handler *MainMenuHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	if bot == nil || r == nil {
		return
	}

	// 处理消息
	if update.Message != nil {
		// 是否由子菜单处理
		var callback *types.Update
		r.Foreach(func(idx int, element *types.Update) bool {
			if element.CallbackQuery != nil {
				callback = element
				return false
			}
			return true
		})

		// 子菜单处理请求
		if update.Message.Text != "/start" && callback != nil {
			newHandler := handler.route(bot, callback.CallbackQuery)
			if newHandler == nil {
				r.Clear()
				return
			}
			newHandler.Handle(bot, r.Push(update), callback)
			return
		}

		// 发送菜单列表
		r.Clear()
		reply, menus := handler.replyMessage(update.Message.From.ID)
		markup := methods.MakeInlineKeyboardMarkup(menus, 2, 2, 2, 1)
		bot.SendMessage(update.Message.Chat.ID, reply, true, markup)
		return
	}

	if update.CallbackQuery == nil {
		return
	}

	// 回复主菜单
	if update.CallbackQuery.Data == "/main/" {
		r.Clear()
		bot.AnswerCallbackQuery(update.CallbackQuery, "", false, "", 0)
		reply, menus := handler.replyMessage(update.CallbackQuery.From.ID)
		markup := methods.MakeInlineKeyboardMarkup(menus, 2, 2, 2, 1)
		bot.EditMessageReplyMarkup(update.CallbackQuery.Message, reply, true, markup)
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
func (handler *MainMenuHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	// 创建红包
	if strings.HasPrefix(query.Data, "/new/") {
		return new(NewHandler)
	}

	// 使用说明
	if strings.HasPrefix(query.Data, "/usage/") {
		return new(UsageHandler)
	}

	// 机器人评分
	if strings.HasPrefix(query.Data, "/rate/") {
		return new(RateBotHandler)
	}

	// 分享机器人
	if strings.HasPrefix(query.Data, "/share/") {
		return new(ShareBotHandler)
	}

	// 操作历史记录
	if strings.HasPrefix(query.Data, "/history/") {
		return new(HistoryHandler)
	}

	// 存款操作
	if strings.HasPrefix(query.Data, "/deposit/") {
		return new(DepositHandler)
	}

	// 提现操作
	if strings.HasPrefix(query.Data, "/withdraw/") {
		return new(WithdrawHandler)
	}
	return nil
}

// 获取用户资产数量
func getUserBalance(userID int64, asset string) (*big.Float, *big.Float) {
	model := models.AccountModel{}
	account, err := model.GetAccount(userID, asset)
	if err != nil {
		if err != storage.ErrNoBucket && err != models.ErrNoSuchTypeAccount {
			logger.Warnf("Failed to get user asset, %v, %v, %v", userID, asset, err)
		}
		return big.NewFloat(0), big.NewFloat(0)
	}
	return account.Amount, account.Locked
}

// 获取回复消息
func (handler *MainMenuHandler) replyMessage(userID int64) (string, []methods.InlineKeyboardButton) {
	// 获取资产信息
	serveCfg := config.GetServe()
	amount, locked := getUserBalance(userID, serveCfg.Symbol)

	// 生成菜单列表
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{Text: tr(userID, "lng_new_lucky_money"), CallbackData: "/new/"},
		methods.InlineKeyboardButton{Text: tr(userID, "lng_history"), CallbackData: "/history/"},
		methods.InlineKeyboardButton{Text: tr(userID, "lng_deposit"), CallbackData: "/deposit/"},
		methods.InlineKeyboardButton{Text: tr(userID, "lng_withdraw"), CallbackData: "/withdraw/"},
		methods.InlineKeyboardButton{Text: tr(userID, "lng_rate"), CallbackData: "/rate/"},
		methods.InlineKeyboardButton{Text: tr(userID, "lng_share"), CallbackData: "/share/"},
		methods.InlineKeyboardButton{Text: tr(userID, "lng_help"), CallbackData: "/usage/"},
	}
	reply := fmt.Sprintf(tr(userID, "lng_welcome"), serveCfg.Name, serveCfg.Symbol,
		amount.String(), serveCfg.Symbol, locked.String(), serveCfg.Symbol)
	return reply, menus[:]
}
