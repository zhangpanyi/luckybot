package luaglue

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

// 加载模块
func HttpLoader(state *lua.LState) int {
	mod := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"get":  get,
		"post": post,
	})
	state.Push(mod)
	return 1
}

// GET方法
func get(state *lua.LState) int {
	url := state.CheckString(-1)
	resp, err := http.Get(url)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
	} else {
		state.Push(formatResp(state, resp))
		state.Push(lua.LNil)
	}
	return 2
}

// POST方法
func post(state *lua.LState) int {
	url := state.CheckString(-3)
	contentType := state.CheckString(-2)
	body := state.CheckString(-1)
	resp, err := http.Post(url, contentType, bytes.NewReader([]byte(body)))
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
	} else {
		state.Push(formatResp(state, resp))
		state.Push(lua.LNil)
	}
	return 2
}

// 格式化响应
func formatResp(state *lua.LState, resp *http.Response) *lua.LTable {
	// 读取Header
	header := state.NewTable()
	for k, v := range resp.Header {
		array := state.NewTable()
		for i := 0; i < len(v); i++ {
			state.RawSetInt(array, i+1, lua.LString(v[i]))
		}
		state.SetField(header, k, array)
	}

	table := state.NewTable()
	state.SetField(table, "header", header)
	state.SetField(table, "status_code", lua.LString(strconv.Itoa(resp.StatusCode)))
	state.SetField(table, "content_length", lua.LNumber(float64(resp.ContentLength)))

	// 读取Body数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		state.SetField(table, "body", lua.LNil)
	} else {
		resp.Body.Close()
		state.SetField(table, "body", lua.LString(body))
	}
	return table
}
