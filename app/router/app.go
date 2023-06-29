package router

import (
	"github.com/gin-gonic/gin"
	"github.com/open_tool/app/controller"
)

func InitRouter(engine *gin.Engine) {
	// 加载模板

	engine.LoadHTMLGlob("app/views/**/*")

	engine.GET("/", controller.AppCtr{}.IndexPage)
	engine.Any("/stream", controller.AppCtr{}.Stream)

	engine.Any("/rpc/ping.action", controller.AppCtr{}.Ping)
	engine.Any("/rpc/obtainTicket.action", controller.AppCtr{}.ObtainTicket)

	// 构建keyPool
	engine.GET("/auth/index", controller.AppCtr{}.KeyPoolPage)
	engine.GET("/auth/make", controller.AppCtr{}.MakeFakePool)

	// MjPrompt
	engine.GET("/midjourney", controller.AppCtr{}.MjPromptIndex)
	engine.POST("/midjourney/make", controller.AppCtr{}.MjPromptMake)

	// Text2Sql
	engine.GET("/text2sql", controller.AppCtr{}.Text2SqlIndex)
	engine.POST("/text2sql/make", controller.AppCtr{}.Text2SqlTrans)
}
