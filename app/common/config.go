package common

import (
	"github.com/fsnotify/fsnotify"
	"github.com/open_tool/app/utils/logger"
	"github.com/spf13/viper"
	"time"
)

// Configuration 项目配置
type Configuration struct {
	Debug int `yaml:"debug"`
	// 指定服务启动端口，默认为 8090
	Port string `yaml:"port"`
	// gpt apikey
	Token string `yaml:"token"`
	// 请求的 URL 地址
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`
	// 使用模型
	Model string `yaml:"model"`
	// 会话超时时间
	SessionTimeout time.Duration `yaml:"session_timeout" mapstructure:"session_timeout"`
	// 最大问题长度
	MaxQuestionLen int `yaml:"max_question_len" mapstructure:"max_question_len"`
	// 最大答案长度
	MaxAnswerLen int `yaml:"max_answer_len" mapstructure:"max_answer_len"`
	// 最大文本 = 问题 + 回答, 接口限制
	MaxText int `yaml:"max_text" mapstructure:"max_text"`
	// 代理地址
	HttpProxy string `yaml:"http_proxy" mapstructure:"http_proxy"`
}

var Conf *Configuration

// InitConfig 加载配置
func InitConfig(path string) {
	//导入配置文件
	viper.SetConfigType("yaml")
	viper.SetConfigFile(path)
	//读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("读取不到配置文件：" + err.Error())
	}
	err = viper.Unmarshal(&Conf)
	if err != nil {
		logger.Error("解析配置文件失败：" + err.Error())
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		err = viper.Unmarshal(&Conf)
		if err != nil {
			logger.Error("配置文件获取异常" + err.Error())
		}
	})

}

// GetConfigData  返回配置数据方法
func GetConfigData() *Configuration {
	return Conf
}
