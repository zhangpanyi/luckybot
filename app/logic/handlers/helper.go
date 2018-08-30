package handlers

import (
	"fmt"
	"math/big"

	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/fmath"
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
	amount := luckyMoney.Amount.String()
	if !luckyMoney.Lucky {
		amount = fmath.Mul(luckyMoney.Amount, big.NewFloat(float64(luckyMoney.Number))).String()
	}
	return fmt.Sprintf(message, luckyMoney.ID, typ, luckyMoney.Number-received, luckyMoney.Number,
		luckyMoney.SenderName, luckyMoney.SenderID,
		amount, luckyMoney.Asset, typ, luckyMoney.Message)
}
