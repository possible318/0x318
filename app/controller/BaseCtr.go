package controller

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

type BaseCtr struct {
	*gin.Context
}

// Response 返回数据
func (c BaseCtr) Response(code int, msg string, data any, expires int) {
	if expires > 0 {
		c.Context.Header("Cache-Control", "public, max-age="+strconv.Itoa(expires))
	} else {
		c.Context.Header("Cache-Control", "no-cache")
	}
	c.JSON(code, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})

}
