package handlers

import (
	"fmt"

	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/storage/models"
)

// 语言翻译
func tr(userID int64, key string) string {
	return config.GetLanguge().Value("zh_CN", key)
}

// 生成红包基本信息
func makeBaseMessage(luckyMoney *models.LuckyMoney, received uint32) string {
	tag := equalLuckyMoney
	if luckyMoney.Lucky {
		tag = randLuckyMoney
	}
	message := tr(luckyMoney.SenderID, "lng_luckymoney_info")
	typ := luckyMoneysTypeToString(luckyMoney.SenderID, tag)
	amount := fmt.Sprintf("%.2f", float64(luckyMoney.Amount)/100.0)
	if !luckyMoney.Lucky {
		amount = fmt.Sprintf("%.2f", float64(luckyMoney.Amount*luckyMoney.Number)/100.0)
	}
	return fmt.Sprintf(message, typ, luckyMoney.Number-received, luckyMoney.Number,
		luckyMoney.SenderName, luckyMoney.SenderID,
		amount, luckyMoney.Asset, typ, luckyMoney.Message)
}
