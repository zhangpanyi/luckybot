package luaglue

import (
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zhangpanyi/luckymoney/app/future"
)

// 创建Future
func newFuture(state *lua.LState, id string) *lua.LTable {
	tabel := state.NewTable()
	state.SetField(tabel, "id", lua.LString(id))
	state.SetFuncs(tabel, map[string]lua.LGFunction{
		"set_result": set_result,
	})
	return tabel
}

// 设置结果
func set_result(state *lua.LState) int {
	self := state.CheckTable(1)
	id := state.GetField(self, "id").(lua.LString)
	result := state.Get(2)
	switch result.Type() {
	case lua.LTNil:
		future.Manager.SetResult(string(id), nil)
	case lua.LTString:
		reason := string(result.(lua.LString))
		future.Manager.SetResult(string(id), errors.New(reason))
	default:
		future.Manager.SetResult(string(id), errors.New(fmt.Sprintf("%v", result)))
	}
	return 0
}
