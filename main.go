package main

import (
	"net/http"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/vrecan/death"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/basebot/telegram/updater"
	"github.com/zhangpanyi/luckymoney/app/config"
	"github.com/zhangpanyi/luckymoney/app/future"
	"github.com/zhangpanyi/luckymoney/app/inspector"
	"github.com/zhangpanyi/luckymoney/app/logic"
	"github.com/zhangpanyi/luckymoney/app/logic/context"
	"github.com/zhangpanyi/luckymoney/app/logic/scriptengine"
	"github.com/zhangpanyi/luckymoney/app/poller"
	"github.com/zhangpanyi/luckymoney/app/storage"
)

func main() {
	// 加载配置文件
	config.LoadConfig("server.yml")

	// 初始化日志库
	serveCfg := config.GetServe()
	logger.CreateLoggerOnce(logger.DebugLevel, logger.InfoLevel)

	// 连接到数据库
	err := storage.Connect(serveCfg.BolTDBPath)
	if err != nil {
		logger.Panic(err)
	}

	// 状态上下文管理
	context.CreateManagerOnce(16)

	// 创建Future管理器
	future.NewFutureManagerOnce()

	// 创建Lua脚本引擎
	scriptengine.NewScriptEngineOnce()

	// 创建机器人轮询器
	poller := poll.NewPoller(serveCfg.APIWebsite)
	bot, err := poller.StartPoll(serveCfg.Token, logic.NewUpdate)
	if err != nil {
		logger.Panic(err)
	}
	logger.Infof("Lucky money bot id is: %d", bot.ID)

	// 启动红包检查器
	pool := updater.NewPool(64)
	inspector.StartChecking(bot, pool)

	// 启动HTTP服务器
	router := mux.NewRouter()
	addr := serveCfg.Host + ":" + strconv.Itoa(serveCfg.Port)
	go func() {
		s := &http.Server{
			Addr:    addr,
			Handler: router,
		}
		if err = s.ListenAndServe(); err != nil {
			logger.Panicf("Failed to listen and serve, %v, %v", addr, err)
		}
	}()
	logger.Infof("Lucky money server started")

	// 捕捉退出信号
	d := death.NewDeath(syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL,
		syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM)
	d.WaitForDeathWithFunc(func() {
		storage.Close()
		logger.Info("Lucky money server stoped")
	})
}
