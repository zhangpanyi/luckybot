package handlers

import (
	"fmt"
	"strconv"

	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/location"
	"github.com/zhangpanyi/luckymoney/app/storage/models"
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
	bot.AnswerInlineQuery(query, nil, 1, result)
}

// 回复红包列表
func replyLuckyMoneyList(bot *methods.BotExt, query *types.InlineQuery) {
	// 筛选查询
	offset, err := strconv.Atoi(query.Offset)
	if len(query.Offset) != 0 && err != nil {
		replyNone(bot, query)
		return
	}

	// 筛选红包
	model := models.LuckyMoneyModel{}
	ids, err := model.FilterLuckyMoney(query.From.ID, uint(offset), 5, true)
	if err != nil || len(ids) == 0 {
		replyNone(bot, query)
		return
	}

	// 查询红包信息
	result := make([]methods.InlineQueryResult, 0)
	for i := 0; i < len(ids); i++ {
		luckyMoney, received, err := model.GetLuckyMoney(ids[i])
		if err != nil || luckyMoney.Received == luckyMoney.Amount {
			continue
		}
		result = append(result, makeLuckyMoneyInfo(luckyMoney, received, i))
	}

	// 生成红包信息
	nextOffet := int32(offset + len(result))
	bot.AnswerInlineQuery(query, &nextOffet, 1, result)
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
	model := models.LuckyMoneyModel{}
	id, err := model.GetLuckyMoneyIDBySN(query.Query)
	if err != nil {
		replyNone(bot, query)
		return
	}
	luckyMoney, received, err := model.GetLuckyMoney(id)
	if err != nil || luckyMoney.Received == luckyMoney.Amount {
		replyNone(bot, query)
		return
	}

	// 生成红包信息
	result := make([]methods.InlineQueryResult, 0)
	result = append(result, makeLuckyMoneyInfo(luckyMoney, received, 0))
	bot.AnswerInlineQuery(query, nil, 1, result)
}

// 生成红包信息
func makeLuckyMoneyInfo(luckyMoney *models.LuckyMoney, received uint32, idx int) methods.InlineQueryResult {
	serveCfg := config.GetServe()
	result := methods.InlineQueryResultArticle{}
	result.ID = strconv.Itoa(idx)
	result.Title = location.Format(luckyMoney.Timestamp)
	result.InputMessageContent = &methods.InputTextMessageContent{
		MessageText:           location.Format(luckyMoney.Timestamp),
		ParseMode:             methods.ParseModeMarkdown,
		DisableWebPagePreview: true,
	}
	if len(serveCfg.ThumbURL) > 0 {
		result.ThumbWidth = 64
		result.ThumbHeight = 64
		result.ThumbURL = serveCfg.ThumbURL
	}

	tag := equalLuckyMoney
	if luckyMoney.Lucky {
		tag = randLuckyMoney
	}

	reply := tr(0, "lng_luckymoney_info")
	result.Description = fmt.Sprintf(reply,
		luckyMoneysTypeToString(0, tag),
		fmt.Sprintf("%.2f", float64(luckyMoney.Amount-luckyMoney.Received)/100.0),
		fmt.Sprintf("%.2f", float64(luckyMoney.Amount)/100.0),
		serveCfg.Symbol,
		luckyMoney.Number-received,
		luckyMoney.Number,
	)
	return &result
}
