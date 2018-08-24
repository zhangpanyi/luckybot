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
