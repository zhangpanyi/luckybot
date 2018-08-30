package pusher

import (
	"container/list"
	"sync"

	"github.com/zhangpanyi/basebot/telegram/methods"
	"github.com/zhangpanyi/basebot/telegram/updater"
)

var once sync.Once
var gpusher *msgPusher

// 运行推送器
func ServiceStart(pool *updater.Pool) {
	once.Do(func() {
		gpusher = &msgPusher{
			pool:  pool,
			queue: list.New(),
			cond:  sync.NewCond(&sync.Mutex{}),
		}
		go gpusher.loop()
	})
}

// 推送器
type msgPusher struct {
	queue *list.List
	cond  *sync.Cond
	pool  *updater.Pool
}

// 推送消息
func (m *msgPusher) push(sender *methods.BotExt, receiver int64, text string,
	markdownMode bool, markup *methods.InlineKeyboardMarkup) {

	// 构造消息结构
	msg := message{
		sender:       sender,
		receiver:     receiver,
		text:         text,
		markdownMode: markdownMode,
		markup:       markup,
	}

	// 添加到推送队列
	m.cond.L.Lock()
	isempty := m.queue.Len() == 0
	m.queue.PushBack(&msg)
	if isempty && m.queue.Len() == 1 {
		m.cond.Signal()
	}
	m.cond.L.Unlock()
}

// 事件循环
func (m *msgPusher) loop() {
	for {
		m.cond.L.Lock()
		for m.queue.Len() == 0 {
			m.cond.Wait()
		}
		for m.queue.Len() > 0 {
			element := m.queue.Front()
			msg, ok := element.Value.(*message)
			if ok {
				m.pool.Async(msg.send)
			}
			m.queue.Remove(element)
		}
		m.cond.L.Unlock()
	}
}
