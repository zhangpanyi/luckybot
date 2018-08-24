package scriptengine

import (
	"sync"

	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/luckymoney/app/luaglue"
)

var once sync.Once
var Engine *luaglue.LuaGlue

// 创建脚本引擎
func NewScriptEngineOnce() {
	once.Do(func() {
		var err error
		Engine, err = luaglue.NewLuaGlue()
		if err != nil {
			logger.Panic(err)
		}
	})
}

// // 事件循环
// func (glue *LuaGlue) eventLoop() {
// 	lasttime := time.Now()
// 	duration := 100 * time.Millisecond
// 	timer := time.NewTimer(duration)
// 	for {
// 		select {
// 		case <-timer.C:
// 			now := time.Now()
// 			glue.OnTick(now.Sub(lasttime).Seconds())
// 			lasttime = now
// 			timer.Reset(duration)
// 		}

// 		if glue.closed {
// 			break
// 		}
// 	}
// }
