package handlers

import (
	"fmt"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
)

// 分享机器人
type ShareBotHandler struct {
}

// 消息处理
func (*ShareBotHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	fromID := update.CallbackQuery.From.ID
	reply := fmt.Sprintf(tr(fromID, "lng_share_say"), bot.UserName,
		fromID, bot.UserName, fromID)
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_back_superior"),
			CallbackData: "/main/",
		},
	}
	markup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)
	bot.EditMessageReplyMarkup(update.CallbackQuery.Message, reply, true, markup)
}

// 消息路由
func (*ShareBotHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}
