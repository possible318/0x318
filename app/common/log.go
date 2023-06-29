package common

//
//func InitLog(engine *gin.Engine) {
//
//	// Disable Console Color, you don't need console color when writing the logs to file.
//	gin.ForceConsoleColor()
//
//	gin.DebugPrintRouteFunc = func(
//		httpMethod,
//		absolutePath,
//		handlerName string,
//		nuHandlers int) {
//		log.Printf("%v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
//	}
//
//	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
//		// 你的自定义格式
//		return fmt.Sprintf("[%s]- %s \"%s %s %s %d %s \"%s\" %s\"\n",
//			param.TimeStamp.Format(time.DateTime),
//			param.ClientIP,
//
//			param.Method,
//			param.Path,
//			param.Request.Proto,
//			param.StatusCode,
//			param.Latency,
//			param.Request.UserAgent(),
//			param.ErrorMessage,
//		)
//	}))
//	engine.Use(gin.Recovery())
//}
