package config

import (
	"log"
	"os"
)

// 定义一个全局的日志记录器
var logger *log.Logger

func init() {
	// 初始化全局日志记录器，写入标准输出，并禁用默认日志设置
	logger = log.New(os.Stdout, "", 0)
}
