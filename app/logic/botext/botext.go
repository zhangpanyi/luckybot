package botext

import (
	"sync"

	"github.com/zhangpanyi/basebot/telegram/methods"
)

// 机器人实例
var once sync.Once
var botInstance *methods.BotExt

// 获取机器人
func GetBot() *methods.BotExt {
	return botInstance
}

// 设置机器人
func SetBot(bot *methods.BotExt) {
	once.Do(func() {
		botInstance = bot
	})
}
