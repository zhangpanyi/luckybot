package handlers

import (
	"fmt"
	"strings"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/types"
	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/storage"
	"github.com/zhangpanyi/luckymoney/app/storage/models"
)

// 领取红包
type ReceiveHandler struct {
}

// 消息处理
func (handler *ReceiveHandler) Handle(bot *methods.BotExt, r *history.History, update *types.Update) {
	if bot == nil || r == nil {
		return
	}
	handler.handleReceiveLuckyMoney(bot, update.CallbackQuery)
}

// 处理红包错误
func (handler *ReceiveHandler) handleReceiveError(bot *methods.BotExt, query *types.CallbackQuery,
	id uint64, err error) {

	// 没有红包
	fromID := query.From.ID
	if err == storage.ErrNoBucket {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_invalid_id"), false, "", 0)
		return
	}

	// 没有激活
	if err == models.ErrNotActivated {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_not_activated"), false, "", 0)
		return
	}

	// 领完了
	if err == models.ErrNothingLeft {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_nothing_left"), false, "", 0)
		return
	}

	// 重复领取
	if err == models.ErrRepeatReceive {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_repeat_receive"), false, "", 0)
		return
	}

	// 红包过期
	if err == models.ErrLuckyMoneydExpired {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_expired"), false, "", 0)
		return
	}

	logger.Errorf("Failed to receive lucky money, id: %d, user_id: %d, %v",
		id, fromID, err)
	bot.AnswerCallbackQuery(query, tr(0, "lng_chat_receive_error"), false, "", 0)
}

// 处理领取红包
func (handler *ReceiveHandler) handleReceiveLuckyMoney(bot *methods.BotExt, query *types.CallbackQuery) {
	// 是否过期
	fromID := query.From.ID
	if query.Data == "expired" {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_expired_say"), false, "", 0)
		return
	}

	// 是否结束
	if query.Data == "removed" {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_nothing_left"), false, "", 0)
		return
	}

	// 获取红包ID
	model := models.LuckyMoneyModel{}
	id, err := model.GetLuckyMoneyIDBySN(query.Data)
	if err != nil {
		bot.AnswerCallbackQuery(query, tr(fromID, "lng_chat_invalid_id"), false, "", 0)
		return
	}

	// 执行领取红包
	value, _, err := model.ReceiveLuckyMoney(id, fromID, query.From.FirstName)
	if err != nil {
		handler.handleReceiveError(bot, query, id, err)
		return
	}
	logger.Warnf("Receive lucky money, id: %d, user_id: %d, value: %d", id, fromID, value)

	// 获取红包信息
	luckyMoney, received, err := model.GetLuckyMoney(id)
	if err != nil {
		logger.Errorf("Failed to get lucky money, %v", err)
		bot.AnswerCallbackQuery(query, tr(0, "lng_chat_receive_error"), false, "", 0)
		return
	}

	// 获取领取记录
	size := 0
	users := make([]string, 0)
	history, err := model.GetHistory(id)
	if err != nil {
		logger.Errorf("Failed to get lucky money history, %v", err)
	}
	serveCfg := config.GetServe()
	for i := 0; i < len(history); i++ {
		user := history[i].User
		message := tr(fromID, "lng_chat_receive_history")
		amount := fmt.Sprintf("%.2f", float64(history[i].Value)/100)
		message = fmt.Sprintf(message, user.FirstName, user.UserID, amount, luckyMoney.Asset)

		size += len(message)
		if size > serveCfg.MaxHistoryTextLen {
			users = append(users, "...")
			break
		}
		users = append(users, message)
	}

	// 更新资产信息
	accountModel := models.AccountModel{}
	_, toAccount, err := accountModel.TransferFromLockAccount(luckyMoney.SenderID, fromID,
		luckyMoney.Asset, uint32(value))
	if err != nil {
		logger.Fatalf("Failed to transfer from lock account, from: %d, to: %d, asset: %s, amount: %d, %v",
			fromID, fromID, luckyMoney.Asset, value, err)
		return
	}

	// 插入账户记录
	versionModel := models.AccountVersionModel{}
	versionModel.InsertVersion(fromID, &models.Version{
		Symbol:          luckyMoney.Asset,
		Balance:         int32(value),
		Amount:          toAccount.Amount,
		Reason:          models.ReasonReceive,
		RefLuckyMoneyID: &luckyMoney.ID,
		RefUserID:       &luckyMoney.SenderID,
		RefUserName:     &luckyMoney.SenderName,
	})

	// 发送领取通知
	alert := tr(0, "lng_chat_receive_success")
	alert = fmt.Sprintf(alert, fmt.Sprintf("%.2f", float64(value)/100), luckyMoney.Asset, bot.UserName)
	bot.AnswerCallbackQuery(query, alert, true, "", 0)

	// 更新按钮信息
	menus := make([]methods.InlineKeyboardButton, 0)
	if received == luckyMoney.Number {
		menus = append(menus, methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_chat_finished"),
			CallbackData: "removed",
		})
	} else {
		menus = append(menus, methods.InlineKeyboardButton{
			Text:         tr(fromID, "lng_chat_receive"),
			CallbackData: luckyMoney.SN,
		})
	}
	replyMarkup := methods.MakeInlineKeyboardMarkupAuto(menus[:], 1)

	// 手气结果统计
	settle := ""
	if received == luckyMoney.Number {
		best, worst, err := model.GetBestAndWorst(id)
		if err == nil && luckyMoney.Number > 1 && luckyMoney.Lucky {
			settle = tr(fromID, "lng_chat_receive_settle")
			bestValue := fmt.Sprintf("%.2f", float64(best.Value)/100.0)
			worstValue := fmt.Sprintf("%.2f", float64(worst.Value)/100.0)
			settle = fmt.Sprintf(settle,
				best.User.FirstName, best.User.UserID, bestValue, luckyMoney.Asset,
				worst.User.FirstName, worst.User.UserID, worstValue, luckyMoney.Asset)
		}
	}

	// 更新红包信息
	message := makeBaseMessage(luckyMoney, received)
	if len(users) > 0 {
		message = fmt.Sprintf(tr(fromID, "lng_chat_receive_format"), message, strings.Join(users, ","), settle)
	}
	bot.EditReplyMarkupByInlineMessageID(*query.InlineMessageID, message, true, replyMarkup)
}
