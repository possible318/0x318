package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var LevelDebug = logrus.DebugLevel

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

func init() {

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	//  设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
	logrus.SetOutput(os.Stdout) //设置输出类型
	file, _ := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	writers := []io.Writer{
		file,
		os.Stdout}
	//  同时写文件和屏幕
	fileAndStdoutWriter := io.MultiWriter(writers...)
	logrus.SetOutput(fileAndStdoutWriter)
}

func SetDebug() {
	logrus.SetLevel(LevelDebug)
}

func Color(color int, msg string) string {
	return fmt.Sprintf("\033[1;%dm %s \033[0m", color, msg)
}

func Info(msg string) {
	logrus.Info(Color(fgBlue, msg))
}

func Warn(msg string) {
	logrus.Warn(Color(bgYellow, msg))
}

func Error(msg string) {
	logrus.Error(Color(bgRed, msg))
}
