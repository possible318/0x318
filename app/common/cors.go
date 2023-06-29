package common

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func InitCors(engine *gin.Engine) {
	corsCnf := cors.New(cors.Config{
		//准许跨域请求网站，多个使用，分开，限制使用
		AllowOrigins: []string{"*"},
		//允许跨域得远点网站，可以直接retrun true就可以了
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		//准许使用得请求方式
		AllowMethods: []string{"PUT", "PATCH", "POST", "DELETE", "GET"},
		//准许使用的请求头
		AllowHeaders: []string{"Origin", "Authorization", "Content-Type"},
		//凭证共享，确定共享
		AllowCredentials: true,
		//显示得请求头
		ExposeHeaders: []string{"Content-Type"},
		//超时设定
		MaxAge: time.Hour * 24,
		//AllowWildcard:          false,
		//AllowBrowserExtensions: false,
		//AllowWebSockets:        false,
		//AllowFiles:             false,
		//AllowAllOrigins:        false,

	})

	engine.Use(corsCnf)
}
