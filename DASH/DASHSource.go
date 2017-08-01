package DASH

import (
	"net/http"
	"wssAPI"
	"errors"
	"events/eStreamerEvent"
	"logger"
	"mediaTypes/flv"
	"time"
)

type DASHSource struct {
	clientId string
	chSvr chan bool
	chValid bool
	streamName string
	sinkAdded bool
	inSvr bool

	audioHeader *flv.FlvTag
	videoHeader *flv.FlvTag
	mpd *mpdCreater
}

func (this *DASHSource)serveHTTP(reqType,param string,w http.ResponseWriter,req *http.Request)  {
	switch reqType {
	case MPD_PREFIX:
		this.serveMPD(param,w,req)
	case Video_PREFIX:
	case Audio_PREFIX:
	}
}

func (this *DASHSource)serveMPD(param string,w http.ResponseWriter,req *http.Request)  {
	if nil==this.videoHeader&&nil==this.audioHeader {
		time.Sleep(time.Second*5)
	}
	if nil==this.videoHeader&&nil==this.audioHeader {
		w.WriteHeader(400)
		logger.LOGE("no media data")
		return
	}

	if this.mpd==nil{
		this.mpd=&mpdCreater{}
		this.mpd.init(this.videoHeader,this.audioHeader)
	}

	bufXML:=this.mpd.GetXML(this.clientId,0)
	w.Header().Set("Content-Type",http.DetectContentType(bufXML))
	w.Write(bufXML)
}

func (this *DASHSource) Init(msg *wssAPI.Msg) (err error) {
	var ok bool
	this.streamName,ok=msg.Param1.(string)
	if false==ok{
		return errors.New("invalid param init DASH source")
	}
	this.chSvr,ok=msg.Param2.(chan bool)
	if false==ok{
		this.chValid=false
		return errors.New("invalid param init hls source")
	}
	this.chValid=true
	this.clientId=wssAPI.GenerateGUID()
	taskAddSink := &eStreamerEvent.EveAddSink{
		StreamName: this.streamName,
		SinkId:     this.clientId,
		Sinker:     this}

	wssAPI.HandleTask(taskAddSink)
	return
}

func (this *DASHSource) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *DASHSource) Stop(msg *wssAPI.Msg) (err error) {
	if this.sinkAdded{
		logger.LOGD("del sink:"+ this.clientId)
		taskDelSink := &eStreamerEvent.EveDelSink{}
		taskDelSink.StreamName = this.streamName
		taskDelSink.SinkId = this.clientId
		go wssAPI.HandleTask(taskDelSink)
		this.sinkAdded = false
		logger.LOGT("del sinker:" + this.clientId)
	}
	if this.inSvr{
		service.Del(this.streamName,this.clientId)
	}
	if this.chValid{
		close(this.chSvr)
		this.chValid=false
	}
	return
}

func (this *DASHSource) GetType() string {
	return ""
}

func (this *DASHSource) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *DASHSource) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_GetSource_NOTIFY:
		if this.chValid {
			this.chSvr<-true
			this.inSvr=true
			close(this.chSvr)
			this.chValid=false
		}
		this.sinkAdded=true
	case wssAPI.MSG_GetSource_Failed:
		this.Stop(nil)
	case wssAPI.MSG_PLAY_START:
		this.sinkAdded=true
	case wssAPI.MSG_PLAY_STOP:
		this.Stop(nil)
	case wssAPI.MSG_FLV_TAG:
		this.addFlvTag(msg.Param1.(*flv.FlvTag))
	}
	return
}

func (this *DASHSource)addFlvTag(tag *flv.FlvTag)  {
	if this.videoHeader==nil&&tag.TagType==flv.FLV_TAG_Video {
		this.videoHeader=tag
	}

	if this.audioHeader==nil&&tag.TagType==flv.FLV_TAG_Audio{
		this.audioHeader=tag
	}
}
