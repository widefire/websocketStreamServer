package RTSPService

import (
	"encoding/json"
	"errors"
	"logger"
	"net"
	"strconv"
	"wssAPI"
)

type RTSPService struct {
}

type RTSPConfig struct {
	Port       int `json:"port"`
	TimeoutSec int `json:"timeoutSec"`
}

var service *RTSPService
var serviceConfig RTSPConfig

func (this *RTSPService) Init(msg *wssAPI.Msg) (err error) {
	if nil == msg || nil == msg.Param1 {
		logger.LOGE("init rtsp server failed")
		return errors.New("invalid param init rtsp server")
	}
	fileName, ok := msg.Param1.(string)
	if false == ok {
		logger.LOGE("bad param init rtsp server")
		return errors.New("invalid param init rtsp server")
	}
	err = this.loadConfigFile(fileName)
	if err != nil {
		logger.LOGE("load rtsp config failed:" + err.Error())
		return
	}
	return
}

func (this *RTSPService) loadConfigFile(fileName string) (err error) {
	data, err := wssAPI.ReadFileAll(fileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &serviceConfig)
	if err != nil {
		return
	}
	if serviceConfig.TimeoutSec <= 0 {
		serviceConfig.TimeoutSec = 60
	}

	if serviceConfig.Port == 0 {
		serviceConfig.Port = 554
	}
	return
}

func (this *RTSPService) Start(msg *wssAPI.Msg) (err error) {
	logger.LOGT("start RTSP server")
	strPort := ":" + strconv.Itoa(serviceConfig.Port)
	tcp, err := net.ResolveTCPAddr("tcp4", strPort)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	listener, err := net.ListenTCP("tcp4", tcp)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	go this.rtspLoop(listener)
	return
}

func (this *RTSPService) rtspLoop(listener *net.TCPListener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.LOGE(err.Error())
			continue
		}
		go this.handleConnect(conn)
	}
}

func (this *RTSPService) handleConnect(conn net.Conn) {
	defer func() {
		conn.Close()
	}()
	handler := &RTSPHandler{}
	handler.conn = conn
	handler.Init(nil)
	for {
		data, err := ReadPacket(conn, handler.tcpTimeout)
		if err != nil {
			logger.LOGE("read rtsp failed")
			logger.LOGE(err.Error())
			handler.handlePacket(nil)
			return
		}
		logger.LOGT(string(data))
		err = handler.handlePacket(data)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
	}
}

func (this *RTSPService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTSPService) GetType() string {
	return wssAPI.OBJ_RTSPServer
}

func (this *RTSPService) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *RTSPService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}
