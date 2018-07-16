package config

import (
	"encoding/json"
	"errors"
	"sync"
)

// 语言包配置
type Languges struct {
	lock sync.RWMutex
	priv map[string]privLanguges
}

type privLanguges map[string]string

// 创建语言包
func NewLanguges() *Languges {
	return &Languges{
		priv: make(map[string]privLanguges),
	}
}

// 获取配置
func (l *Languges) Value(code string, key string) string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	lang, ok := l.priv[code]
	if !ok {
		return ""
	}

	val, ok := lang[key]
	if !ok {
		return ""
	}
	return val
}

// 解析数据
func (l *Languges) parse(data []byte) error {
	lang := make(privLanguges)
	err := json.Unmarshal(data, &lang)
	if err != nil {
		return err
	}

	code, ok := lang["lng_language_code"]
	if !ok {
		return errors.New("not found lng_language_code")
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	l.priv[code] = lang
	return nil
}
