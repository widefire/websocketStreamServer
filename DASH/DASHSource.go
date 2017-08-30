package DASH

import (
	"net/http"
	"wssAPI"
	"errors"
	"events/eStreamerEvent"
	"logger"
	"mediaTypes/flv"
	"github.com/panda-media/muxer-fmp4/dashSlicer"
	"github.com/panda-media/muxer-fmp4/codec/H264"

	"os"
)

type DASHSource struct {
	clientId string
	chSvr chan bool
	chValid bool
	streamName string
	sinkAdded bool
	inSvr bool

	slicer *dashSlicer.DASHSlicer
	appendedAACHeader bool
	appendedKeyFrame bool
}

func (this *DASHSource)serveHTTP(reqType,param string,w http.ResponseWriter,req *http.Request)  {
	switch reqType {
	case MPD_PREFIX:
		this.serveMPD(param,w,req)
	case Video_PREFIX:
		logger.LOGD(param)
		this.serveVideo(param,w,req)
	case Audio_PREFIX:
		logger.LOGD(param)
		this.serveAudio(param,w,req)
	}
}

func (this *DASHSource)serveMPD(param string,w http.ResponseWriter,req *http.Request)  {

	mpd,err:= this.slicer.GetLastedMPD()
	if err!=nil{
		logger.LOGE(err.Error())
		return
	}
	w.Header().Set("Content-Type","application/dash+xml")
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Write(mpd)

}

func (this *DASHSource)serveVideo(param string,w http.ResponseWriter,req *http.Request){
	data,err:=this.slicer.GetVideoData(param)
	if err!=nil{
		logger.LOGE(err.Error())
		return
	}
	if data!=nil{
		w.Write(data)
		wssAPI.CreateDirectory("dashMedia")
		fileName:="dashMedia/"+param
		fp,_:=os.Create(fileName)
		defer fp.Close()
		fp.Write(data)
	}
}

func (this *DASHSource)serveAudio(param string,w http.ResponseWriter,req *http.Request){
	data,err:=this.slicer.GetAudioData(param)
	if err!=nil{
		logger.LOGE(err.Error())
		return
	}
	if data!=nil{
		w.Write(data)
		wssAPI.CreateDirectory("dashMedia")
		fileName:="dashMedia/"+param
		fp,_:=os.Create(fileName)
		defer fp.Close()
		fp.Write(data)
	}
}

func (this *DASHSource) Init(msg *wssAPI.Msg) (err error) {

	this.slicer=dashSlicer.NEWSlicer(true,2000,10000,5)

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
	switch tag.TagType {
	case flv.FLV_TAG_Audio:
		if false==this.appendedAACHeader{
			this.slicer.AddAACFrame(tag.Data[2:])
			this.appendedAACHeader=true
		}else{
			if this.appendedKeyFrame{
				this.slicer.AddAACFrame(tag.Data[2:])
			}
		}
	case flv.FLV_TAG_Video:
		if tag.Data[0]==0x17&&tag.Data[1]==0{
			logger.LOGD("AVC")
			avc,err:=H264.DecodeAVC(tag.Data[5:])
			if err!=nil{
				logger.LOGE(err.Error())
				return
			}
			for e := avc.SPS.Front(); e != nil; e = e.Next() {
				nal := make([]byte, 3+len(e.Value.([]byte)))
				nal[0] = 0
				nal[1] = 0
				nal[2] = 1
				copy(nal[3:], e.Value.([]byte))
				this.slicer.AddH264Nals(nal)
			}
			for e := avc.PPS.Front(); e != nil; e = e.Next() {
				nal := make([]byte, 3+len(e.Value.([]byte)))
				nal[0] = 0
				nal[1] = 0
				nal[2] = 1
				copy(nal[3:], e.Value.([]byte))
				this.slicer.AddH264Nals(nal)
			}
		}else {
			cur := 5
			for cur < len(tag.Data) {
				size := int(tag.Data[cur]) << 24
				size |= int(tag.Data[cur+1]) << 16
				size |= int(tag.Data[cur+2]) << 8
				size |= int(tag.Data[cur+3]) << 0
				cur += 4
				nal := make([]byte, 3+size)
				nal[0] = 0
				nal[1] = 0
				nal[2] = 1
				copy(nal[3:], tag.Data[cur:cur+size])
				if false==this.appendedKeyFrame{
					if tag.Data[cur]&0x1f==H264.NAL_IDR_SLICE{
						this.appendedKeyFrame=true
					}else{
						cur+=size
						continue
					}
				}
				this.slicer.AddH264Nals(nal)
				cur += size
			}
		}
	}
}
