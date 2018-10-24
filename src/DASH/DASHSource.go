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
	"fmt"
	"strings"
	"mediaTypes/h264"
)


type DASHSource struct {
	clientId string
	chSvr chan bool
	chValid bool
	streamName string
	sinkAdded bool
	inSvr bool

	slicer *dashSlicer.DASHSlicer
	mediaReceiver *FMP4Cache
	appendedAACHeader bool
	appendedKeyFrame bool
	audioHeader *flv.FlvTag
	videoHeader *flv.FlvTag
}

func (this *DASHSource)serveHTTP(reqType,param string,w http.ResponseWriter,req *http.Request)  {
	switch reqType {
	case MPD_PREFIX:
		this.serveMPD(param,w,req)
	case Video_PREFIX:
		//logger.LOGD(param)
		this.serveVideo(param,w,req)
	case Audio_PREFIX:
		//logger.LOGD(param)
		this.serveAudio(param,w,req)
	}
}

func (this *DASHSource)serveMPD(param string,w http.ResponseWriter,req *http.Request)  {

	if nil==this.slicer{
		w.WriteHeader(404)
		return
	}
	mpd,err:= this.slicer.GetMPD()
	if err!=nil{
		logger.LOGE(err.Error())
		return
	}
	//mpd,err=wssAPI.ReadFileAll("mpd/taotao.mpd")
	//if err!=nil{
	//	logger.LOGE(err.Error())
	//	return
	//}
	w.Header().Set("Content-Type","application/dash+xml")
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Write(mpd)

}

func (this *DASHSource)serveVideo(param string,w http.ResponseWriter,req *http.Request){
	var data []byte
	var err error
	if strings.Contains(param,"_init_"){
		data,err=this.mediaReceiver.GetVideoHeader()
	}else{
		id:=int64(0)
		fmt.Sscanf(param,"video_video0_%d_mp4.m4s",&id)
		data,err=this.mediaReceiver.GetVideoSegment(id)
	}
	if err!=nil{
		w.WriteHeader(404)
		return
	}

	w.Write(data)
}

func (this *DASHSource)serveAudio(param string,w http.ResponseWriter,req *http.Request){
	var data []byte
	var err error
	if strings.Contains(param,"_init_"){
		data,err=this.mediaReceiver.GetAudioHeader()
	}else{
		id:=int64(0)
		fmt.Sscanf(param,"audio_audio0_%d_mp4.m4s",&id)
		data,err=this.mediaReceiver.GetAudioSegment(id)
	}
	if err!=nil{
		w.WriteHeader(404)
		return
	}
	w.Write(data)
}

func (this *DASHSource) Init(msg *wssAPI.Msg) (err error) {

	//this.slicer=dashSlicer.NEWSlicer(true,2000,10000,5)

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

func (this *DASHSource)createSlicer() (err error){
	var fps int
	if nil!=this.videoHeader {
		if this.videoHeader.Data[0] == 0x17 && this.videoHeader.Data[1] == 0 {
			avc, err := H264.DecodeAVC(this.videoHeader.Data[5:])
			if err != nil {
				logger.LOGE(err.Error())
				return err
			}
			for e := avc.SPS.Front(); e != nil; e = e.Next() {
				_,_,fps=h264.ParseSPS(e.Value.([]byte))
				break;
			}
		}
		this.mediaReceiver = NewFMP4Cache(5)
		this.slicer, err = dashSlicer.NEWSlicer(fps, 1000, 1000, 1000, 9000, 5, this.mediaReceiver)
		if err != nil {
			logger.LOGE(err.Error())
			return err
		}
	}else{
		err=errors.New("invalid video  header")
		return
	}

	if nil!=this.audioHeader{
		this.slicer.AddAACFrame(this.audioHeader.Data[2:],int64(this.audioHeader.Timestamp))
	}
	tag:=this.videoHeader.Copy()
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
		this.slicer.AddH264Nals(nal,int64(tag.Timestamp))
	}
	for e := avc.PPS.Front(); e != nil; e = e.Next() {
		nal := make([]byte, 3+len(e.Value.([]byte)))
		nal[0] = 0
		nal[1] = 0
		nal[2] = 1
		copy(nal[3:], e.Value.([]byte))
		this.slicer.AddH264Nals(nal,int64(tag.Timestamp))
	}
	return
}

func (this *DASHSource)addFlvTag(tag *flv.FlvTag)  {
	switch tag.TagType {
	case flv.FLV_TAG_Audio:
		if nil==this.audioHeader{
			this.audioHeader=tag.Copy()
			return
		}else if this.slicer==nil {
			this.createSlicer()
		}
		if false==this.appendedAACHeader{
			logger.LOGD("AAC")
			this.slicer.AddAACFrame(tag.Data[2:],int64(tag.Timestamp))
			this.appendedAACHeader=true
		}else{
			if this.appendedKeyFrame{
				this.slicer.AddAACFrame(tag.Data[2:],int64(tag.Timestamp))
			}
		}
	case flv.FLV_TAG_Video:
		if nil==this.videoHeader{
			this.videoHeader=tag.Copy()
			return
		}else if nil==this.slicer {
			this.createSlicer()
		}
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
			//this.slicer.AddH264Nals(nal,int64(tag.Timestamp))
			cur += size
		}
		compositionTime:=int(tag.Data[2])<<16
		compositionTime|=int(tag.Data[3])<<8
		compositionTime|=int(tag.Data[4])<<0
		this.slicer.AddH264Frame(tag.Data[5:],int64(tag.Timestamp),compositionTime)
	}
}
