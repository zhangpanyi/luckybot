package poll

import (
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/updater"
)

// 轮询器
type Poller struct {
	apiaccess string
}

// 创建轮询器
func NewPoller(apiaccess string) *Poller {
	poller := new(Poller)
	poller.apiaccess = apiaccess
	return poller
}

// 开始轮询
func (poller *Poller) StartPoll(token string, handler updater.Handler) (*methods.BotExt, error) {
	bot, err := methods.GetMe(poller.apiaccess, token)
	if err != nil {
		return nil, err
	}
	err = methods.DelWebhook(poller.apiaccess, token)
	if err != nil {
		return nil, err
	}
	go poller.startPoll(bot, handler)
	return bot, nil

}

func (poller *Poller) startPoll(bot *methods.BotExt, handler updater.Handler) {
	var offset uint32
	for {
		updates, err := bot.GetUpdates(5, offset)
		if err != nil {
			logger.Infof("Failed to get updates, %v", err)
			continue
		}

		for i := 0; i < len(updates); i++ {
			go handler(bot, updates[i])
			offset = uint32(updates[i].UpdateID + 1)
		}
	}
}
