package main

import (
	"github.com/zhangpanyi/basebot/logger"
	"github.com/zhangpanyi/luckybot/app/config"
	"github.com/zhangpanyi/luckybot/app/logic/scriptengine"
)

func main() {
	// 加载配置文件
	config.LoadConfig("server.yml")

	// 初始化日志库
	//serveCfg := config.GetServe()
	logger.CreateLoggerOnce(logger.DebugLevel, logger.InfoLevel)

	// // 连接到数据库
	// err := storage.Connect(serveCfg.BolTDBPath)
	// if err != nil {
	// 	logger.Panic(err)
	// }

	// // 状态上下文管理
	// context.CreateManagerOnce(16)

	// // 创建Future管理器
	// future.NewFutureManagerOnce()

	// 创建Lua脚本引擎
	scriptengine.NewScriptEngineOnce()

	// // 创建机器人轮询器
	// poller := poll.NewPoller(serveCfg.APIWebsite)
	// bot, err := poller.StartPoll(serveCfg.Token, logic.NewUpdate)
	// if err != nil {
	// 	logger.Panic(err)
	// }
	// botext.SetBot(bot)
	// logger.Infof("Lucky money bot id: %d", bot.ID)

	// // 启动红包检查器
	// pool := updater.NewPool(64)
	// monitor.StartChecking(bot, pool)

	// // 运行推送服务
	// pusher.ServiceStart(pool)

	// // 启动HTTP服务器
	// router := mux.NewRouter()
	// admin.InitRoute(router)
	// router.HandleFunc("/deposit", deposit.HandleDeposit)
	// addr := serveCfg.Host + ":" + strconv.Itoa(serveCfg.Port)
	// go func() {
	// 	s := &http.Server{
	// 		Addr:    addr,
	// 		Handler: router,
	// 	}
	// 	if err = s.ListenAndServe(); err != nil {
	// 		logger.Panicf("Failed to listen and serve, %v, %v", addr, err)
	// 	}
	// }()
	// logger.Infof("Lucky money server started")

	// // 捕捉退出信号
	// d := death.NewDeath(syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL,
	// 	syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM)
	// d.WaitForDeathWithFunc(func() {
	// 	storage.Close()
	// 	logger.Infof("Lucky money server stoped")
	// })
}
