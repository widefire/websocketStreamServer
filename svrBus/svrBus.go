package svrBus

import (
	"HLSService"
	"RTMPService"
	"backend"
	"encoding/json"
	"errors"
	"logger"
	"os"
	"runtime"
	"streamer"
	"strings"
	"sync"
	"time"
	"webSocketService"
	"wssAPI"
	"HTTPMUX"
	"DASH"
	"RTSPService"
)

type busConfig struct {
	RTMPConfigName          string `json:"RTMP"`
	WebSocketConfigName     string `json:"WebSocket"`
	BackendConfigName       string `json:"Backend"`
	LogPath                 string `json:"LogPath"`
	StreamManagerConfigName string `json:"Streamer"`
	HLSConfigName           string `json:"HLS"`
	DASHConfigName 			string `json:"DASH,omitempty"`
	RTSPConfigName string `json:"RTSP,omitempty"`
}

type SvrBus struct {
	mutexServices sync.RWMutex
	services      map[string]wssAPI.Obj
}

var service *SvrBus

func init() {
	service = &SvrBus{}
	wssAPI.SetBus(service)
}

func Start() {
	service.Init(nil)
	service.Start(nil)
}

func (this *SvrBus) Init(msg *wssAPI.Msg) (err error) {
	this.services = make(map[string]wssAPI.Obj)
	err = this.loadConfig()
	if err != nil {
		logger.LOGE("svr bus load config failed")
		return
	}
	return
}

func (this *SvrBus) loadConfig() (err error) {
	var configName string
	if len(os.Args) > 1 {
		configName = os.Args[1]
	} else {
		logger.LOGW("use default :config.json")
		configName = "config.json"
	}
	data, err := wssAPI.ReadFileAll(configName)
	if err != nil {
		logger.LOGE("load config file failed:" + err.Error())
		return
	}
	cfg := &busConfig{}
	err = json.Unmarshal(data, cfg)

	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	if len(cfg.LogPath) > 0 {
		this.createLogFile(cfg.LogPath)
	}

	if true {
		livingSvr := &streamer.StreamerService{}
		msg := &wssAPI.Msg{}
		if len(cfg.StreamManagerConfigName) > 0 {
			msg.Param1 = cfg.StreamManagerConfigName
		}
		err = livingSvr.Init(msg)
		if err != nil {
			logger.LOGE(err.Error())
		} else {
			this.mutexServices.Lock()
			this.services[livingSvr.GetType()] = livingSvr
			this.mutexServices.Unlock()
		}
	}

	if len(cfg.RTMPConfigName) > 0 {
		rtmpSvr := &RTMPService.RTMPService{}
		msg := &wssAPI.Msg{}
		msg.Param1 = cfg.RTMPConfigName
		err = rtmpSvr.Init(msg)
		if err != nil {
			logger.LOGE(err.Error())
		} else {
			this.mutexServices.Lock()
			this.services[rtmpSvr.GetType()] = rtmpSvr
			this.mutexServices.Unlock()
		}
	}

	if len(cfg.WebSocketConfigName) > 0 {
		webSocketSvr := &webSocketService.WebSocketService{}
		msg := &wssAPI.Msg{}
		msg.Param1 = cfg.WebSocketConfigName
		err = webSocketSvr.Init(msg)
		if err != nil {
			logger.LOGE(err.Error())
		} else {
			this.mutexServices.Lock()
			this.services[webSocketSvr.GetType()] = webSocketSvr
			this.mutexServices.Unlock()
		}
	}

	if len(cfg.BackendConfigName) > 0 {
		backendSvr := &backend.BackendService{}
		msg := &wssAPI.Msg{}
		msg.Param1 = cfg.BackendConfigName
		err = backendSvr.Init(msg)
		if err != nil {
			logger.LOGE(err.Error())
		} else {
			this.mutexServices.Lock()
			this.services[backendSvr.GetType()] = backendSvr
			this.mutexServices.Unlock()
		}
	}

	if len(cfg.RTSPConfigName) > 0 {
		rtspSvr := &RTSPService.RTSPService{}
		msg := &wssAPI.Msg{}
		msg.Param1 = cfg.RTSPConfigName
		err = rtspSvr.Init(msg)
		if err != nil {
			logger.LOGE(err.Error())
		} else {
			this.mutexServices.Lock()
			this.services[rtspSvr.GetType()] = rtspSvr
			this.mutexServices.Unlock()
		}
	}

	if len(cfg.HLSConfigName) > 0 {
		hls := &HLSService.HLSService{}
		msg := &wssAPI.Msg{Param1: cfg.HLSConfigName}
		err = hls.Init(msg)
		if err != nil {
			logger.LOGE(err.Error())
		} else {
			this.mutexServices.Lock()
			this.services[hls.GetType()] = hls
			this.mutexServices.Unlock()
		}
	}

	if len(cfg.DASHConfigName)>0{
		dash:=&DASH.DASHService{}
		msg:=&wssAPI.Msg{Param1:cfg.DASHConfigName}
		err=dash.Init(msg)
		if err!=nil{
			logger.LOGE(err.Error())
		}else{
			this.mutexServices.Lock()
			this.services[dash.GetType()]=dash
			this.mutexServices.Unlock()
		}
	}
	return
}

func (this *SvrBus) createLogFile(logPath string) {
	if strings.HasSuffix(logPath, "/") {
		logPath = strings.TrimSuffix(logPath, "/")
	}
	dir := logPath + time.Now().Format("/2006/01/02/")
	bResult, _ := wssAPI.CheckDirectory(dir)

	if false == bResult {
		_, err := wssAPI.CreateDirectory(dir)
		if err != nil {
			logger.LOGE("create log file failed:", err.Error())
			return
		}
	}
	fullName := dir + time.Now().Format("2006-01-02_15.04") + ".log"
	fp, err := os.OpenFile(fullName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, os.ModePerm)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	logger.SetOutput(fp)
	//avoid one log file too big
	go func() {
		logFileTick := time.Tick(time.Hour * 72)
		for {
			select {
			case <-logFileTick:
				fullName := dir + time.Now().Format("2006-01-02_15:04") + ".log"
				newLogFile, _ := os.OpenFile(fullName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, os.ModePerm)
				if newLogFile != nil {
					logger.SetOutput(newLogFile)
					fp.Close()
					fp = newLogFile
				}
			}
		}
	}()
}

func (this *SvrBus) Start(msg *wssAPI.Msg) (err error) {
	//if false {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//}
	HTTPMUX.Start()
	this.mutexServices.RLock()
	defer this.mutexServices.RUnlock()
	for k, v := range this.services {
		//v.SetParent(this)
		err = v.Start(nil)
		if err != nil {
			logger.LOGE("start " + k + " failed:" + err.Error())
			continue
		}
		logger.LOGI("start " + k + " successed")
	}

	return
}

func (this *SvrBus) Stop(msg *wssAPI.Msg) (err error) {
	this.mutexServices.RLock()
	defer this.mutexServices.RUnlock()
	for _, v := range this.services {
		err = v.Stop(nil)
	}
	return
}

func (this *SvrBus) GetType() string {
	return wssAPI.OBJ_ServerBus
}

func (this *SvrBus) HandleTask(task wssAPI.Task) (err error) {
	this.mutexServices.RLock()
	defer this.mutexServices.RUnlock()
	handler, exist := this.services[task.Receiver()]
	if exist == false {
		return errors.New("invalid task")
	}
	return handler.HandleTask(task)
}

func (this *SvrBus) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return nil
}

func (this *SvrBus) SetParent(arent wssAPI.Obj) {

}

func AddSvr(svr wssAPI.Obj) {
	logger.LOGE("add svr")
}
