package main

import (
	"net/http"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/vrecan/death"
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/tg-lucky-money/app/config"
	"github.com/zhangpanyi/tg-lucky-money/app/logic"
	"github.com/zhangpanyi/tg-lucky-money/app/logic/context"
	"github.com/zhangpanyi/tg-lucky-money/app/poller"
	"github.com/zhangpanyi/tg-lucky-money/app/storage"
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

	// 消息上下文管理
	context.CreateManagerForOnce(16)

	// 创建轮询器
	poller := poll.NewPoller(serveCfg.APIWebsite)
	bot, err := poller.StartPoll(serveCfg.Token, logic.NewUpdate)
	if err != nil {
		logger.Panic(err)
	}
	logger.Infof("Lucky money bot id is: %d", bot.ID)

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
