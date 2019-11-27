package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	// 设置日志格式为json格式
	log.SetFormatter(&log.JSONFormatter{})

	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)

	// 设置日志级别为warn以上
	log.SetLevel(log.WarnLevel)
}
func main() {
	log.Error("Helloerror")
	log.Warn("hello warn")
	s := strings.Split("- 1210471895651749945 2 IN IP4 127.0.0.1", " ")
	log.Info(s[0])

}
