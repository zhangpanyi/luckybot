package luaglue

import (
	"github.com/yuin/gopher-lua"
)

// Lua胶水
type LuaGlue struct {
	closed bool
	state  *lua.LState
}

// 创建实例
func NewLuaGlue() (*LuaGlue, error) {
	state := lua.NewState()
	state.PreloadModule("http", HttpLoader)
	if err := state.DoFile("scripts/main.lua"); err != nil {
		return nil, err
	}
	glue := LuaGlue{state: state}
	return &glue, nil
}

// 释放资源
func (glue *LuaGlue) Close() {
	glue.state.Close()
	glue.closed = true
}

// 时钟事件
func (glue *LuaGlue) OnTick(delaytime float64) {
	fn := glue.state.GetGlobal("on_tick")
	if fn == nil {
		return
	}

	glue.state.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, lua.LNumber(delaytime))
}

// 账户是否有效
func (glue *LuaGlue) ValidAccount(account string) bool {
	fn := glue.state.GetGlobal("valid_account")
	if fn == nil {
		return false
	}

	glue.state.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, lua.LString(account))

	ret := glue.state.Get(-1)
	defer glue.state.Pop(1)
	if ret.Type() != lua.LTBool {
		return false
	}

	val := ret.(lua.LBool)
	return bool(val)
}

// 接收提现请求
func (glue *LuaGlue) OnWithdraw(to, symbol, amount, id string) {
	fn := glue.state.GetGlobal("on_withdraw")
	if fn == nil {
		return
	}

	future := newFuture(glue.state, id)
	glue.state.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, lua.LString(to), lua.LString(symbol), lua.LString(amount), future)
}

// 交易是否有效
func (glue *LuaGlue) ValidTransaction(from, to, symbol, amount, txid string) bool {
	fn := glue.state.GetGlobal("valid_transaction")
	if fn == nil {
		return false
	}

	glue.state.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, lua.LString(from), lua.LString(to), lua.LString(symbol), lua.LString(amount), lua.LString(txid))

	ret := glue.state.Get(-1)
	defer glue.state.Pop(1)
	if ret.Type() != lua.LTBool {
		return false
	}

	val := ret.(lua.LBool)
	return bool(val)
}
