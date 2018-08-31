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

	err := state.Get(-1)
	txid := state.Get(-2)
	if err.Type() == lua.LTNil {
		if txid.Type() == lua.LTString {
			future.Manager.SetResult(string(id), string(txid.(lua.LString)), nil)
		}
	} else {
		future.Manager.SetResult(string(id), "", errors.New(fmt.Sprintf("%v", err)))
	}
	return 0
}
