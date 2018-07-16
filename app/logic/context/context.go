package context

import (
	"sync"

	"github.com/zhangpanyi/basebot/history"
	"github.com/zhangpanyi/basebot/logger"
)

var once sync.Once
var manager *history.Manager

// 创建记录管理器
func CreateManagerForOnce(bucketNum uint32) {
	once.Do(func() {
		var err error
		manager, err = history.NewManager(bucketNum)
		if err != nil {
			logger.Panicf("Failed to create manager for lucky money, %v", err)
		}
	})
}

// 删除记录
func DelRecord(userID uint32) {
	manager.Del(userID)
}

// 获取记录
func GetRecord(userID uint32) (*history.History, error) {
	return manager.Get(userID)
}
