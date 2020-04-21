package utils

import (
	"fmt"

	"luckybot/app/config"
	"luckybot/app/storage/models"
)

// 语言翻译
func Tr(userID int64, key string) string {
	return config.GetLanguge().Value("zh_CN", key)
}

// 生成历史内容
func MakeHistoryMessage(fromID int64, version *models.Version) string {
	switch version.Reason {
	case models.ReasonGive:
		// 发放红包
		message := Tr(fromID, "lng_history_give")
		return fmt.Sprintf(message, *version.RefLuckyMoneyID,
			version.Locked.String(), version.Symbol)
	case models.ReasonReceive:
		// 领取红包
		message := Tr(fromID, "lng_history_receive")
		return fmt.Sprintf(message, *version.RefUserName,
			*version.RefUserID, *version.RefLuckyMoneyID, version.Balance.String(), version.Symbol)
	case models.ReasonSystem:
		// 系统发放
		message := Tr(fromID, "lng_history_system")
		return fmt.Sprintf(message, version.Balance.String(), version.Symbol)
	case models.ReasonGiveBack:
		// 退还红包
		message := Tr(fromID, "lng_history_giveback")
		return fmt.Sprintf(message, *version.RefLuckyMoneyID,
			version.Locked.Abs(version.Locked).String(), version.Symbol)
	case models.ReasonDeposit:
		// 充值代币
		message := Tr(fromID, "lng_history_deposit")
		return fmt.Sprintf(message, version.Balance.String(), version.Symbol,
			*version.RefBlockHeight, *version.RefTxID)
	case models.ReasonWithdraw:
		// 正在提现
		serverCfg := config.GetServe()
		message := Tr(fromID, "lng_history_withdraw")
		return fmt.Sprintf(message, version.Locked.String(), version.Symbol, serverCfg.Name,
			*version.RefAddress, version.Fee.String(), version.Symbol)
	case models.ReasonWithdrawFailure:
		// 提现失败
		serverCfg := config.GetServe()
		message := Tr(fromID, "lng_history_withdraw_failure")
		return fmt.Sprintf(message, version.Locked.Abs(version.Locked).String(), version.Symbol,
			serverCfg.Name, *version.RefAddress)
	case models.ReasonWithdrawSuccess:
		// 提现成功
		serverCfg := config.GetServe()
		message := Tr(fromID, "lng_history_withdraw_success")
		return fmt.Sprintf(message, version.Locked.Abs(version.Locked).String(), version.Symbol,
			serverCfg.Name, *version.RefAddress, *version.RefTxID)
	}
	return ""
}
