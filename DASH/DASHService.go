package DASH

import (
	"wssAPI"
	"logger"
	"errors"
	"strconv"
	"HTTPMUX"
	"net/http"
	"encoding/json"
	"strings"
	"sync"
	"time"
)
//http://addr/DASH/streamName/mpd/id.mpd
//http://addr/DASH/streamName/video/init.mp4 or number.m4s
//http://addr/DASH/streamName/audio/init.mp4 or number.m4s
const (
	MPD_PREFIX="mpd"
	Video_PREFIX="video"
	Audio_PREFIX="audio"
)


type DASHService struct {
	sources map[string]*DASHSource
	muxSource sync.RWMutex
}

type DASHConfig struct {
	Port int `json:"Port"`
	Route string `json:"Route"`
}

var service *DASHService
var serviceConfig DASHConfig

func (this *DASHService) Init(msg *wssAPI.Msg) (err error) {
	defer func() {
		if nil!=err{
			logger.LOGE(err.Error())
		}
	}()
	if nil==msg||nil==msg.Param1{
		err=errors.New("invalid param")
		return
	}
	fileName:=msg.Param1.(string)
	err=this.loadConfigFile(fileName)
	if nil!=err{
		return
	}
	strPort := ":" + strconv.Itoa(serviceConfig.Port)
	HTTPMUX.AddRoute(strPort,serviceConfig.Route,this.ServeHTTP)
	this.sources=make(map[string]*DASHSource)
	return
}

func (this *DASHService)ServeHTTP(w http.ResponseWriter,req *http.Request)  {
	streamName,reqType,param,err:=this.parseURL(req.URL.String())
	if err!=nil{
		logger.LOGE(err.Error())
		return
	}
	this.muxSource.RLock()
	source,exist:=this.sources[streamName]
	this.muxSource.RUnlock()
	if false==exist{

		this.createSource(streamName)

		this.muxSource.RLock()
		source,exist=this.sources[streamName]
		this.muxSource.RUnlock()

		if false==exist{
			w.WriteHeader(400)
			w.Write([]byte("bad request"))
		}
	}
	source.serveHTTP(reqType,param,w,req)
}
func (this *DASHService) createSource(streamName string) {
	ch:=make(chan bool,1)
	msg:=&wssAPI.Msg{Param1:streamName,Param2:ch}
	source:=&DASHSource{}
	err:=source.Init(msg)
	if nil!=err{
		logger.LOGE(err.Error())
		return
	}
	select {
	case ret,ok:=<-ch:
		if false==ok||false==ret{
			source.Stop(nil)
		}else{
			this.muxSource.Lock()
			defer this.muxSource.Unlock()
			_,exist:=this.sources[streamName]
			if exist {
				logger.LOGD("competition:" + streamName)
				return
			}else{
				this.sources[streamName]=source
			}
		}
		case <-time.After(time.Minute):
		source.Stop(nil)
		return
	}
}

func (this *DASHService)loadConfigFile(fileName string) (err error) {
	buf,err:=wssAPI.ReadFileAll(fileName)
	if err!=nil{
		return
	}
	err=json.Unmarshal(buf,&serviceConfig)
	return err
}

func (this *DASHService) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *DASHService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *DASHService) GetType() string {
	return wssAPI.OBJ_DASHServer
}

func (this *DASHService) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *DASHService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *DASHService)parseURL(url string)(streamName,reqType,param string,err error)  {
	url=strings.TrimPrefix(url,serviceConfig.Route)
	url=strings.TrimSuffix(url,"/")
	subs:=strings.Split(url,"/")
	if len(subs)<3{
		err=errors.New("invalid request :"+url)
		return
	}
	param=subs[len(subs)-1]
	url=strings.TrimSuffix(url,param)
	url=strings.TrimSuffix(url,"/")
	reqType=subs[len(subs)-2]
	url=strings.TrimSuffix(url,reqType)
	streamName=strings.TrimSuffix(url,"/")

	logger.LOGD(streamName,reqType,param)
	return 
}

func (this *DASHService)Add(name string,src *DASHSource) (err error) {
	this.muxSource.Lock()
	defer this.muxSource.Unlock()
	_,exist:=this.sources[name]
	if exist{
		err=errors.New("source existed")
		return
	}
	this.sources[name]=src
	return
}

func (this *DASHService)Del(name,id string)  {
	this.muxSource.Lock()
	defer this.muxSource.RUnlock()
	src,exist:=this.sources[name]
	if exist&&src.clientId==id{
		delete(this.sources,name)
	}
}

