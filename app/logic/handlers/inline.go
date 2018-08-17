package handlers

import (
	"fmt"
	"strconv"

	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/location"
	"github.com/zhangpanyi/luckymoney/app/storage"
)

// 显示红包信息
func ShowLuckyMoney(bot *methods.BotExt, query *types.InlineQuery) {
	if query.Query == "list" {
		replyLuckyMoneyList(bot, query)
		return
	}
	replyLuckyMoneyInfo(bot, query)
}

// 回复空信息
func replyNone(bot *methods.BotExt, query *types.InlineQuery) {
	result := make([]methods.InlineQueryResult, 0)
	bot.AnswerInlineQuery(query, 0, 0, result)
}

// 回复红包列表
func replyLuckyMoneyList(bot *methods.BotExt, query *types.InlineQuery) {
	// 筛选查询
	offset, err := strconv.Atoi(query.Offset)
	if len(query.Offset) > 0 && err != nil {
		replyNone(bot, query)
		return
	}
	if offset > 0 {
		replyNone(bot, query)
		return
	}

	// 查询信息
	handler := storage.LuckyMoneyStorage{}
	ids, err := handler.AllLuckyMoney(query.From.ID, uint(offset), 5)
	if err != nil {
		replyNone(bot, query)
		return
	}

	// 查询红包信息
	result := make([]methods.InlineQueryResult, 0)
	for i := len(ids) - 1; i >= 0; i-- {
		luckyMoney, received, err := handler.GetLuckyMoney(ids[i])
		if err != nil || luckyMoney.Received == luckyMoney.Amount {
			continue
		}
		result = append(result, makeLuckyMoneyInfo(luckyMoney, received, i))
	}

	// 生成红包信息
	bot.AnswerInlineQuery(query, int32(offset+len(result)), 0, result)
}

// 回复红包信息
func replyLuckyMoneyInfo(bot *methods.BotExt, query *types.InlineQuery) {
	// 筛选查询
	offset, err := strconv.Atoi(query.Offset)
	if len(query.Offset) > 0 && err != nil {
		replyNone(bot, query)
		return
	}
	if offset > 0 {
		replyNone(bot, query)
		return
	}

	// 查询信息
	handler := storage.LuckyMoneyStorage{}
	id, err := handler.GetLuckyMoneyIDBySN(query.Query)
	if err != nil {
		replyNone(bot, query)
		return
	}
	luckyMoney, received, err := handler.GetLuckyMoney(id)
	if err != nil || luckyMoney.Received == luckyMoney.Amount {
		replyNone(bot, query)
		return
	}

	// 生成红包信息
	result := make([]methods.InlineQueryResult, 0)
	result = append(result, makeLuckyMoneyInfo(luckyMoney, received, 0))
	bot.AnswerInlineQuery(query, int32(offset+len(result)), 0, result)
}

// 生成红包信息
func makeLuckyMoneyInfo(luckyMoney *storage.LuckyMoney, received uint32, idx int) methods.InlineQueryResult {
	result := methods.InlineQueryResultArticle{}
	result.ID = strconv.Itoa(idx)
	result.Title = location.Format(luckyMoney.Timestamp)
	result.InputMessageContent = &methods.InputTextMessageContent{
		MessageText:           location.Format(luckyMoney.Timestamp),
		ParseMode:             methods.ParseModeMarkdown,
		DisableWebPagePreview: true,
	}

	tag := equalLuckyMoney
	if luckyMoney.Lucky {
		tag = randLuckyMoney
	}

	serveCfg := config.GetServe()
	reply := tr(0, "lng_luckymoney_info")
	result.Description = fmt.Sprintf(reply,
		luckyMoneysTypeToString(0, tag),
		luckyMoney.ID,
		fmt.Sprintf("%.2f", float64(luckyMoney.Amount-luckyMoney.Received)/100.0),
		fmt.Sprintf("%.2f", float64(luckyMoney.Amount)/100.0),
		serveCfg.Symbol,
		received,
		luckyMoney.Number,
	)
	return &result
}
