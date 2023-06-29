package main

import (
	"github.com/gin-gonic/gin"
	"github.com/open_tool/app/common"
	"github.com/open_tool/app/router"
	"github.com/open_tool/app/utils/logger"
)

// 路由初始化
var engine = gin.Default()

func init() {
	gin.SetMode(gin.ReleaseMode)

	//读取yaml中得配置信息
	common.InitConfig("config.yaml")

	//初始化数据库连接
	common.InitDB("./data/account.db")

	// 初始化跨域
	common.InitCors(engine)
	//加载路由
	router.InitRouter(engine)

}

func main() {
	// 关闭debug模式
	configs := common.GetConfigData()
	if configs.Debug == 1 {
		gin.SetMode(gin.DebugMode)
	}
	port := configs.Port

	logger.Info("Server is running at http://0.0.0.0:" + port)
	err := engine.Run("0.0.0.0:" + port)
	if err != nil {
		logger.Error("Failed to start server:" + err.Error())
	}
}
