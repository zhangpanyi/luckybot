package handlers

import (
	"fmt"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/tg-lucky-money/app/config"
)

// 存款
type DepositHandler struct {
}

// 消息处理
func (*DepositHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	serveCfg := config.GetServe()
	fromID := update.CallbackQuery.From.ID
	reply := fmt.Sprintf(tr(fromID, "lng_deposit_say"), serveCfg.Symbol, serveCfg.Address, fromID)
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: "/main/",
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
	bot.AnswerCallbackQuery(update.CallbackQuery, "", false, "", 0)
	bot.EditMessageReplyMarkup(update.CallbackQuery.Message, reply, true, markup)
}

// 消息路由
func (*DepositHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}
