package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// 前景色
	fgBlack  = 30
	fgRed    = 31
	fgGreen  = 32
	fgYellow = 33
	fgBlue   = 34
	fgPurple = 35
	fgCyan   = 36
	fgGray   = 37
	// 背景色
	bgBlack  = 40
	bgRed    = 41
	bgGreen  = 42
	bgYellow = 43
	bgBlue   = 44
	bgPurple = 45
	bgCyan   = 46
	bgGray   = 47
)

var colorMap = map[string]int{
	"green":  bgGreen,
	"white":  bgGray,
	"yellow": bgYellow,
	"red":    bgRed,

	"GET":     bgBlue,
	"POST":    bgCyan,
	"PUT":     bgYellow,
	"DELETE":  bgRed,
	"PATCH":   bgGreen,
	"HEAD":    bgPurple,
	"OPTIONS": bgGray,
}

// ColorByMethod return color by http Method
func ColorByMethod(method string) string {
	color, err := colorMap[method]
	if err {
		color = colorMap["white"]
	}
	return fmt.Sprintf("\033[%dm %s \033[0m", color, method)
}

// ColorByStatus return color by http code
func ColorByStatus(code int) string {
	var color int
	switch {
	case code >= 200 && code < 300:
		color = colorMap["green"]
	case code >= 300 && code < 400:
		color = colorMap["white"]
	case code >= 400 && code < 500:
		color = colorMap["yellow"]
	default:
		color = colorMap["red"]
	}
	return fmt.Sprintf("\033[%dm %d \033[0m", color, code)
}

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped

		// Stop timer
		end := time.Now()
		timeSub := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		//bodySize := c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}

		statusMsg := ColorByStatus(statusCode)

		methodMsg := ColorByMethod(method)

		logrus.Infof("[GIN] %s |%s| %d | %s | %s | %s",
			start.Format("2006-01-02 15:04:06"),
			statusMsg,
			timeSub,
			clientIP,
			methodMsg,
			path,
		)

	}

}
