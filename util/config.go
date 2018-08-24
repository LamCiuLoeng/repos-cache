package util

import (
	// 自动从.env文件中读取环境变量，方便开发
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
)

// NewConfig 用于从环境变量中提取和保存项目配置信息
func NewConfig() *viper.Viper {
	config := viper.New()
	config.AutomaticEnv()
	return config
}
