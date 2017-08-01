package RTMPService

import (
	"encoding/json"
	"errors"
	"events/eRTMPEvent"
	"fmt"
	"logger"
	"net"
	"strconv"
	"strings"
	"time"
	"wssAPI"
)

const (
	rtmpTypeHandler  = "rtmpHandler"
	rtmpTypePuller   = "rtmpPuller"
	livePathDefault  = "live"
	timeoutDefault   = 3000
	rtmpCacheDefault = 1000
)

type RTMPService struct {
	listener *net.TCPListener
	parent   wssAPI.Obj
}

func init() {

}

type RTMPConfig struct {
	Port       int    `json:"Port"`
	TimeoutSec int    `json:"TimeoutSec"`
	LivePath   string `json:"LivePath"`
	CacheCount int    `json:"CacheCount"`
}

var service *RTMPService
var serviceConfig RTMPConfig

func (this *RTMPService) Init(msg *wssAPI.Msg) (err error) {
	if nil == msg || nil == msg.Param1 {
		logger.LOGE("invalid param init rtmp server")
		return errors.New("init rtmp service failed")
	}
	fileName := msg.Param1.(string)
	err = this.loadConfigFile(fileName)
	if err != nil {
		logger.LOGE(err.Error())
		return errors.New("init rtmp service failed")
	}
	service = this
	return
}

func (this *RTMPService) Start(msg *wssAPI.Msg) (err error) {
	logger.LOGT("start rtmp service")
	strPort := ":" + strconv.Itoa(serviceConfig.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", strPort)
	if nil != err {
		logger.LOGE(err.Error())
		return
	}
	this.listener, err = net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	go this.rtmpLoop()
	return
}

func (this *RTMPService) Stop(msg *wssAPI.Msg) (err error) {
	this.listener.Close()
	return
}

func (this *RTMPService) GetType() string {
	return wssAPI.OBJ_RTMPServer
}

func (this *RTMPService) HandleTask(task wssAPI.Task) (err error) {
	if task.Receiver() != this.GetType() {
		return errors.New("not my task")
	}
	switch task.Type() {
	case eRTMPEvent.PullRTMPStream:
		taskPull, ok := task.(*eRTMPEvent.EvePullRTMPStream)
		if false == ok {
			return errors.New("invalid param to pull rtmp stream")
		}
		taskPull.Protocol = strings.ToLower(taskPull.Protocol)
		switch taskPull.Protocol {
		case "rtmp":
			PullRTMPLive(taskPull)
		default:
			logger.LOGE(fmt.Sprintf("fmt %s not support now", taskPull.Protocol))
			close(taskPull.Src)
			return errors.New("fmt not support")
		}
		return
	default:
		return errors.New(fmt.Sprintf("task %s not prossed", task.Type()))
	}
	return
}

func (this *RTMPService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPService) loadConfigFile(fileName string) (err error) {
	data, err := wssAPI.ReadFileAll(fileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &serviceConfig)
	if err != nil {
		return
	}

	if serviceConfig.TimeoutSec == 0 {
		serviceConfig.TimeoutSec = timeoutDefault
	}

	if len(serviceConfig.LivePath) == 0 {
		serviceConfig.LivePath = livePathDefault
	}
	if serviceConfig.CacheCount == 0 {
		serviceConfig.CacheCount = rtmpCacheDefault
	}
	strPort := ""
	if serviceConfig.Port != 1935 {
		strPort = strconv.Itoa(serviceConfig.Port)
	}
	logger.LOGI("rtmp://address:" + strPort + "/" + serviceConfig.LivePath + "/streamName")
	logger.LOGI("rtmp timeout: " + strconv.Itoa(serviceConfig.TimeoutSec) + " s")
	return
}

func (this *RTMPService) rtmpLoop() {
	for {
		conn, err := this.listener.Accept()
		if err != nil {
			logger.LOGW(err.Error())
			continue
		}
		go this.handleConnect(conn)
	}
}

func (this *RTMPService) handleConnect(conn net.Conn) {
	var err error
	defer conn.Close()
	defer logger.LOGT("close connect>>>")
	err = rtmpHandleshake(conn)
	if err != nil {
		logger.LOGE("rtmp handle shake failed")
		return
	}

	rtmp := &RTMP{}
	rtmp.Init(conn)

	msgInit := &wssAPI.Msg{}
	msgInit.Param1 = rtmp

	handler := &RTMPHandler{}
	err = handler.Init(msgInit)
	if err != nil {
		logger.LOGE("rtmp handler init failed")
		return
	}
	logger.LOGT("new connect:" + conn.RemoteAddr().String())
	for {
		var packet *RTMPPacket
		packet, err = this.readPacket(rtmp, handler.isPlaying())
		if err != nil {
			handler.HandleRTMPPacket(nil)
			return
		}
		err = handler.HandleRTMPPacket(packet)
		if err != nil {
			return
		}
	}
}

func (this *RTMPService) readPacket(rtmp *RTMP, playing bool) (packet *RTMPPacket, err error) {
	if false == playing {
		err = rtmp.Conn.SetReadDeadline(time.Now().Add(time.Duration(serviceConfig.TimeoutSec) * time.Second))
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		defer rtmp.Conn.SetReadDeadline(time.Time{})
	}
	packet, err = rtmp.ReadPacket()
	return
}

func (this *RTMPService) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}
