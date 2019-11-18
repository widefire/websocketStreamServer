package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/widefire/websocketStreamServer/RTSP"
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
	s := strings.Split("-10:07:00", "-")
	log.Info(s[0])
	RTSP.ParseSMPTE("smpte-25=10:07:00-10:07:33:05.01")
}
