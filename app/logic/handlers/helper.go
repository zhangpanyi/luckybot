package handlers

import (
	"fmt"
	"math/big"

	"luckybot/app/fmath"
	"luckybot/app/logic/handlers/utils"
	"luckybot/app/storage/models"
)

// 语言翻译
func tr(userID int64, key string) string {
	return utils.Tr(userID, key)
}

// 生成红包基本信息
func makeBaseMessage(luckyMoney *models.LuckyMoney, received uint32) string {
	tag := equalLuckyMoney
	if luckyMoney.Lucky {
		tag = randLuckyMoney
	}
	message := tr(luckyMoney.SenderID, "lng_luckymoney_info")
	typ := luckyMoneysTypeToString(luckyMoney.SenderID, tag)
	amount := luckyMoney.Amount.String()
	if !luckyMoney.Lucky {
		amount = fmath.Mul(luckyMoney.Amount, big.NewFloat(float64(luckyMoney.Number))).String()
	}
	return fmt.Sprintf(message, luckyMoney.ID, typ, luckyMoney.Number-received, luckyMoney.Number,
		luckyMoney.SenderName, luckyMoney.SenderID,
		amount, luckyMoney.Asset, typ, luckyMoney.Message)
}
