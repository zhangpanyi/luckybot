package luaglue

import (
	"encoding/json"
	"math/big"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// 加载模块
func JsonLoader(state *lua.LState) int {
	mod := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"dump":  dump,
		"parse": parse,
	})
	state.Push(mod)
	return 1
}

// 判断整数
func isInteger(num lua.LNumber) bool {
	val := big.NewFloat(float64(num))
	return len(strings.Split(val.String(), ".")) == 1
}

// 判断数组
func isPureArray(table *lua.LTable) bool {
	ret := true
	lastidx := 0
	table.ForEach(func(k lua.LValue, v lua.LValue) {
		if !ret {
			return
		}

		if k.Type() != lua.LTNumber {
			ret = false
			return
		}

		if !isInteger(k.(lua.LNumber)) {
			ret = false
			return
		}

		idx := int(k.(lua.LNumber))
		if lastidx+1 != idx {
			ret = false
			return
		}
		lastidx++
	})
	return ret
}

func dumpArray(table *lua.LTable) []interface{} {
	arr := make([]interface{}, 0)
	table.ForEach(func(k lua.LValue, v lua.LValue) {
		switch v.Type() {
		case lua.LTNil:
			arr = append(arr, nil)
		case lua.LTBool:
			arr = append(arr, bool(v.(lua.LBool)))
		case lua.LTString:
			arr = append(arr, string(v.(lua.LString)))
		case lua.LTNumber:
			arr = append(arr, float64(v.(lua.LNumber)))
		case lua.LTTable:
			table = v.(*lua.LTable)
			if isPureArray(table) {
				arr = append(arr, dumpArray(table))
			} else {
				arr = append(arr, dumpObject(table))
			}
		}
	})
	return arr
}

func dumpObject(table *lua.LTable) map[string]interface{} {
	hash := make(map[string]interface{})
	table.ForEach(func(k lua.LValue, v lua.LValue) {
		if k.Type() == lua.LTString {
			switch v.Type() {
			case lua.LTNil:
				hash[string(k.(lua.LString))] = nil
			case lua.LTBool:
				hash[string(k.(lua.LString))] = bool(v.(lua.LBool))
			case lua.LTString:
				hash[string(k.(lua.LString))] = string(v.(lua.LString))
			case lua.LTNumber:
				hash[string(k.(lua.LString))] = float64(v.(lua.LNumber))
			case lua.LTTable:
				newTable := v.(*lua.LTable)
				if isPureArray(newTable) {
					hash[string(k.(lua.LString))] = dumpArray(newTable)
				} else {
					hash[string(k.(lua.LString))] = dumpObject(newTable)
				}
			}
		}
	})
	return hash
}

func dump(state *lua.LState) int {
	var err error
	var jsb []byte
	table := state.CheckTable(-1)
	if isPureArray(table) {
		arr := dumpArray(table)
		jsb, err = json.Marshal(&arr)
	} else {
		hash := dumpObject(table)
		jsb, err = json.Marshal(&hash)
	}

	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
	} else {
		state.Push(lua.LString(string(jsb)))
		state.Push(lua.LNil)
	}
	return 2
}

func parseArray(state *lua.LState, arr []interface{}) *lua.LTable {
	table := state.NewTable()
	for i := 0; i < len(arr); i++ {
		switch arr[i].(type) {
		case nil:
			table.Append(lua.LString("null"))
		case bool:
			table.Append(lua.LBool(arr[i].(bool)))
		case string:
			table.Append(lua.LString(arr[i].(string)))
		case float64:
			table.Append(lua.LNumber(arr[i].(float64)))
		case []interface{}:
			newArr := arr[i].([]interface{})
			table.Append(parseArray(state, newArr))
		case map[string]interface{}:
			dict := arr[i].(map[string]interface{})
			table.Append(parseObject(state, dict))
		}
	}
	return table
}

func parseObject(state *lua.LState, dict map[string]interface{}) *lua.LTable {
	table := state.NewTable()
	for k, v := range dict {
		switch v.(type) {
		case nil:
			state.SetField(table, k, lua.LNil)
		case bool:
			state.SetField(table, k, lua.LBool(v.(bool)))
		case string:
			state.SetField(table, k, lua.LString(v.(string)))
		case float64:
			state.SetField(table, k, lua.LNumber(v.(float64)))
		case []interface{}:
			newArr := v.([]interface{})
			state.SetField(table, k, parseArray(state, newArr))
		case map[string]interface{}:
			dict := v.(map[string]interface{})
			state.SetField(table, k, parseObject(state, dict))
		}
	}
	return table
}

func parse(state *lua.LState) int {
	var err error
	var table *lua.LTable
	jstring := state.CheckString(-1)
	if len(jstring) >= 2 && jstring[0] == '[' && jstring[len(jstring)-1] == ']' {
		var arr []interface{}
		err = json.Unmarshal([]byte(jstring), &arr)
		if err == nil {
			table = parseArray(state, arr)
		}
	} else {
		var dict map[string]interface{}
		err = json.Unmarshal([]byte(jstring), &dict)
		if err == nil {
			table = parseObject(state, dict)
		}
	}

	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
	} else {
		state.Push(table)
		state.Push(lua.LNil)
	}
	return 2
}
