package pusher

import (
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
)

// 消息内容
type message struct {
	sender       *methods.BotExt               // 发送者
	receiver     int64                         // 接收者
	text         string                        // 文本
	markdownMode bool                          // MarkDown渲染
	markup       *methods.InlineKeyboardMarkup // Reply Markup
}

// 发送消息
func (msg *message) send() {
	_, err := msg.sender.SendMessage(msg.receiver, msg.text,
		msg.markdownMode, msg.markup)
	if err != nil {
		logger.Warnf("Failed to push message, %v", err)
	}
}
