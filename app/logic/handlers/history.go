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
	"github.com/zhangpanyi/luckybot/app/config"
	"github.com/zhangpanyi/luckybot/app/location"
	"github.com/zhangpanyi/luckybot/app/storage/models"
)

// 每页条目
const PageLimit = 5

// 匹配历史页数
var reMathHistoryPage *regexp.Regexp

func init() {
	var err error
	reMathHistoryPage, err = regexp.Compile("^/history/(|(\\d+)/)$")
	if err != nil {
		panic(err)
	}
}

// 生成历史内容
func MakeHistoryMessage(fromID int64, version *models.Version) string {
	switch version.Reason {
	case models.ReasonGive:
		// 发放红包
		message := tr(fromID, "lng_history_give")
		return fmt.Sprintf(message, *version.RefLuckyMoneyID,
			version.Locked.String(), version.Symbol)
	case models.ReasonReceive:
		// 领取红包
		message := tr(fromID, "lng_history_receive")
		return fmt.Sprintf(message, *version.RefUserName,
			*version.RefUserID, *version.RefLuckyMoneyID, version.Balance.String(), version.Symbol)
	case models.ReasonSystem:
		// 系统发放
		message := tr(fromID, "lng_history_system")
		return fmt.Sprintf(message, version.Balance.String(), version.Symbol)
	case models.ReasonGiveBack:
		// 退还红包
		message := tr(fromID, "lng_history_giveback")
		return fmt.Sprintf(message, *version.RefLuckyMoneyID,
			version.Locked.Abs(version.Locked).String(), version.Symbol)
	case models.ReasonDeposit:
		// 充值代币
		message := tr(fromID, "lng_history_deposit")
		return fmt.Sprintf(message, version.Balance.String(), version.Symbol,
			*version.RefBlockHeight, *version.RefTxID)
	case models.ReasonWithdraw:
		// 正在提现
		serverCfg := config.GetServe()
		message := tr(fromID, "lng_history_withdraw")
		return fmt.Sprintf(message, version.Locked.String(), version.Symbol, serverCfg.Name,
			*version.RefAddress, version.Fee.String(), version.Symbol)
	case models.ReasonWithdrawFailure:
		// 提现失败
		serverCfg := config.GetServe()
		message := tr(fromID, "lng_history_withdraw_failure")
		return fmt.Sprintf(message, version.Locked.Abs(version.Locked).String(), version.Symbol,
			serverCfg.Name, *version.RefAddress)
	case models.ReasonWithdrawSuccess:
		// 提现成功
		serverCfg := config.GetServe()
		message := tr(fromID, "lng_history_withdraw_success")
		return fmt.Sprintf(message, version.Locked.Abs(version.Locked).String(), version.Symbol,
			serverCfg.Name, *version.RefAddress, *version.RefTxID)
	}
	return ""
}

// 历史记录
type HistoryHandler struct {
}

// 消息处理
func (handler *HistoryHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	data := update.CallbackQuery.Data
	result := reMathHistoryPage.FindStringSubmatch(data)
	if len(result) == 3 {
		page, err := strconv.Atoi(result[2])
		if err != nil {
			handler.replyHistory(bot, 1, update.CallbackQuery)
		} else {
			handler.replyHistory(bot, page, update.CallbackQuery)
		}
	}
}

// 消息路由
func (handler *HistoryHandler) route(bot *methods.BotExt, query *types.CallbackQuery) Handler {
	return nil
}

// 生成菜单列表
func (handler *HistoryHandler) makeMenuList(fromID int64, page, pagesum int) *methods.InlineKeyboardMarkup {
	privpage := page - 1
	if privpage < 1 {
		privpage = 1
	}
	nextpage := page + 1
	if nextpage > pagesum {
		nextpage = pagesum
	}
	priv := fmt.Sprintf("/history/%d/", privpage)
	next := fmt.Sprintf("/history/%d/", nextpage)
	menus := [...]methods.InlineKeyboardButton{
		methods.InlineKeyboardButton{Text: tr(fromID, "lng_previous_page"), CallbackData: priv},
		methods.InlineKeyboardButton{Text: tr(fromID, "lng_next_page"), CallbackData: next},
		methods.InlineKeyboardButton{Text: tr(fromID, "lng_back_superior"), CallbackData: "/main/"},
	}
	return methods.MakeInlineKeyboardMarkupAuto(menus[:], 2)
}

// 生成回复内容
func (handler *HistoryHandler) makeReplyContent(fromID int64, array []*models.Version, page, pagesum uint) string {
	header := fmt.Sprintf("%s (*%d*/%d)\n\n", tr(fromID, "lng_history"), page, pagesum)
	if len(array) > 0 {
		lines := make([]string, 0, len(array))
		for _, version := range array {
			date := location.Format(version.Timestamp)
			lines = append(lines, fmt.Sprintf("`%s` %s", date, MakeHistoryMessage(fromID, version)))
		}
		return header + strings.Join(lines, "\n\n")
	}
	return header + tr(fromID, "lng_priv_history_no_op")
}

// 回复历史记录
func (handler *HistoryHandler) replyHistory(bot *methods.BotExt, page int, query *types.CallbackQuery) {
	// 检查页数
	if page < 1 {
		page = 1
	}

	// 查询历史
	fromID := query.From.ID
	model := models.AccountVersionModel{}
	history, sum, err := model.GetVersions(fromID, uint((page-1)*PageLimit), PageLimit, true)
	if err != nil {
		logger.Warnf("Failed to query user history, %v", err)
	}
	pagesum := sum / PageLimit
	if sum%PageLimit > 0 {
		pagesum++
	}

	// 回复内容
	if len(history) > 0 {
		bot.AnswerCallbackQuery(query, "", false, "", 0)
	} else {
		reply := tr(fromID, "lng_history_no_op")
		bot.AnswerCallbackQuery(query, reply, false, "", 0)
		return
	}
	reply := handler.makeReplyContent(fromID, history, uint(page), uint(pagesum))
	bot.EditMessageReplyMarkup(query.Message, reply, true, handler.makeMenuList(fromID, page, pagesum))
}
