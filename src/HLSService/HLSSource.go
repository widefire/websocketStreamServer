package HLSService

import (
	"container/list"
	"errors"
	"events/eStreamerEvent"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"mediaTypes/ts"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"wssAPI"
)

type hlsTsData struct {
	buf        []byte
	durationMs float64
	idx        int
}

const TsCacheLength = 4
const (
	MasterM3U8 = "master.m3u8"
	videoPref  = "v"
	audioPref  = "a"
)

type HLSSource struct {
	sinkAdded    bool
	inSvrMap     bool
	chValid      bool
	chSvr        chan bool
	streamName   string
	urlPref      string
	clientId     string
	audioHeader  *flv.FlvTag
	videoHeader  *flv.FlvTag
	segIdx       int64
	tsCur        *ts.TsCreater
	audioCur     *audioCache
	tsCache      *list.List
	muxCache     sync.RWMutex
	beginTime    uint32
	waitsChannel *list.List
	muxWaits     sync.RWMutex
}

func (this *HLSSource) Init(msg *wssAPI.Msg) (err error) {
	this.sinkAdded = false
	this.inSvrMap = false
	this.chValid = false
	this.tsCache = list.New()
	this.waitsChannel = list.New()
	this.segIdx = 0
	this.beginTime = 0
	var ok bool
	this.streamName, ok = msg.Param1.(string)
	if false == ok {
		return errors.New("invalid param init hls source")
	}
	this.chSvr, ok = msg.Param2.(chan bool)
	if false == ok {
		return errors.New("invalid param init hls source")
	}
	this.chValid = true

	//create source
	this.clientId = wssAPI.GenerateGUID()
	taskAddSink := &eStreamerEvent.EveAddSink{
		StreamName: this.streamName,
		SinkId:     this.clientId,
		Sinker:     this}

	wssAPI.HandleTask(taskAddSink)

	logger.LOGD("init end")
	if strings.Contains(this.streamName, "/") {
		//this.urlPref="/"+serviceConfig.Route+"/"+this.streamName
		this.urlPref = "/" + serviceConfig.Route + "/" + this.streamName
	}
	return
}

func (this *HLSSource) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *HLSSource) Stop(msg *wssAPI.Msg) (err error) {
	defer func() {
		if err := recover(); err != nil {
			logger.LOGD(err)
		}
	}()
	//从源移除
	if this.sinkAdded {
		logger.LOGD("del sink")
		taskDelSink := &eStreamerEvent.EveDelSink{}
		taskDelSink.StreamName = this.streamName
		taskDelSink.SinkId = this.clientId
		go wssAPI.HandleTask(taskDelSink)
		this.sinkAdded = false
		logger.LOGT("del sinker:" + this.clientId)
	}
	//从service移除
	if this.inSvrMap {
		this.inSvrMap = false
		service.DelSource(this.streamName, this.clientId)
	}
	//清理数据
	if this.chValid {
		close(this.chSvr)
		this.chValid = false
	}

	this.muxWaits.Lock()
	defer this.muxWaits.Unlock()
	if this.waitsChannel != nil {
		for e := this.waitsChannel.Front(); e != nil; e = e.Next() {
			close(e.Value.(chan bool))
		}
		this.waitsChannel = list.New()
	}
	return
}

func (this *HLSSource) GetType() string {
	return ""
}

func (this *HLSSource) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *HLSSource) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_GetSource_NOTIFY:
		//对端还没有监听 所以写不进去
		if this.chValid {
			this.chSvr <- true
			this.inSvrMap = true
			logger.LOGD("get source notify")
		}
		this.sinkAdded = true
		logger.LOGD("get source by notify")
	case wssAPI.MSG_GetSource_Failed:
		this.Stop(nil)
	case wssAPI.MSG_PLAY_START:
		this.sinkAdded = true
		logger.LOGD("get source by start play")
	case wssAPI.MSG_PLAY_STOP:
		//hls 停止就结束移除，不像RTMP等待
		this.Stop(nil)
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		this.AddFlvTag(tag)
	default:
		logger.LOGT(msg.Type)
	}
	return
}

func (this *HLSSource) ServeHTTP(w http.ResponseWriter, req *http.Request, param string) {
	if strings.HasSuffix(param, ".ts") {
		//get ts file
		this.serveTs(w, req, param)
	} else {
		//get m3u8 file
		this.serveM3u8(w, req, param)
	}
}

func (this *HLSSource) serveTs(w http.ResponseWriter, req *http.Request, param string) {
	if strings.HasPrefix(param, videoPref) {
		this.serveVideo(w, req, param)
	} else {
		logger.LOGE("no audio now")
	}
}

func (this *HLSSource) serveM3u8(w http.ResponseWriter, req *http.Request, param string) {

	if MasterM3U8 == param {
		this.serveMaster(w, req, param)
		return
	}
}

func (this *HLSSource) createVideoM3U8(tsCacheCopy *list.List) (strOut string) {
	//max duration
	maxDuration := 0

	if tsCacheCopy.Len() == TsCacheLength {
		tsCacheCopy.Remove(tsCacheCopy.Front())
	}
	for e := tsCacheCopy.Front(); e != nil; e = e.Next() {
		dura := int(e.Value.(*hlsTsData).durationMs / 1000.0)
		if dura > maxDuration {
			maxDuration = dura
		}
	}
	if maxDuration == 0 {
		maxDuration = 1
	}
	//sequence
	sequence := tsCacheCopy.Front().Value.(*hlsTsData).idx

	strOut = "#EXTM3U\n"
	strOut += "#EXT-X-TARGETDURATION:" + strconv.Itoa(maxDuration) + "\n"
	strOut += "#EXT-X-VERSION:3\n"
	strOut += "#EXT-X-MEDIA-SEQUENCE:" + strconv.Itoa(sequence) + "\n"
	//strOut += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	strOut += "#EXT-X-INDEPENDENT-SEGMENTS\n"
	//last two ts？？
	//if tsCacheCopy.Len()<=TsCacheLength{
	if true {
		for e := tsCacheCopy.Front(); e != nil; e = e.Next() {
			tmp := e.Value.(*hlsTsData)
			strOut += fmt.Sprintf("#EXTINF:%f,\n", tmp.durationMs/1000.0)
			//strOut += this.urlPref+"/"+strconv.Itoa(tmp.idx) + ".ts" + "\n"
			strOut += "v" + strconv.Itoa(tmp.idx) + ".ts" + "\n"
		}
	} else {
		e := tsCacheCopy.Front()
		for i := 0; i < tsCacheCopy.Len()-2; i++ {
			e = e.Next()
		}
		for ; e != nil; e = e.Next() {
			tmp := e.Value.(*hlsTsData)
			strOut += fmt.Sprintf("#EXTINF:%f,\n", tmp.durationMs/1000.0)
			//strOut += this.urlPref+"/"+strconv.Itoa(tmp.idx) + ".ts" + "\n"
			strOut += "v" + strconv.Itoa(tmp.idx) + ".ts" + "\n"
		}
	}
	//strOut += "#EXT-X-ENDLIST\n"
	return strOut
}

func (this *HLSSource) serveMaster(w http.ResponseWriter, req *http.Request, param string) {

	this.muxCache.RLock()
	tsCacheCopy := list.New()
	for e := this.tsCache.Front(); e != nil; e = e.Next() {
		tsCacheCopy.PushBack(e.Value)
	}
	this.muxCache.RUnlock()
	if tsCacheCopy.Len() > 0 {
		w.Header().Set("Content-Type", "Application/vnd.apple.mpegurl")
		strOut := this.createVideoM3U8(tsCacheCopy)
		w.Write([]byte(strOut))
	} else {
		//wait for new
		chWait := make(chan bool, 1)
		this.muxWaits.Lock()
		this.waitsChannel.PushBack(chWait)
		this.muxWaits.Unlock()
		select {
		case ret, ok := <-chWait:
			if false == ok || false == ret {
				w.WriteHeader(404)
				logger.LOGE("no data now")
				return
			} else {
				this.muxCache.RLock()
				tsCacheCopy := list.New()
				for e := this.tsCache.Front(); e != nil; e = e.Next() {
					tsCacheCopy.PushBack(e.Value)
				}
				this.muxCache.RUnlock()
				strOut := this.createVideoM3U8(tsCacheCopy)
				w.Header().Set("Content-Type", "Application/vnd.apple.mpegurl")
				w.Write([]byte(strOut))
			}
		case <-time.After(time.Minute):
			w.WriteHeader(404)
			logger.LOGE("time out")
			return
		}
	}
}

func (this *HLSSource) serveVideo(w http.ResponseWriter, req *http.Request, param string) {
	strIdx := strings.TrimPrefix(param, "v")
	strIdx = strings.TrimSuffix(strIdx, ".ts")
	idx, _ := strconv.Atoi(strIdx)
	this.muxCache.RLock()
	defer this.muxCache.RUnlock()
	for e := this.tsCache.Front(); e != nil; e = e.Next() {
		tsData := e.Value.(*hlsTsData)
		if tsData.idx == idx {
			w.Write(tsData.buf)
			return
		}
	}
}

func (this *HLSSource) AddFlvTag(tag *flv.FlvTag) {
	if this.audioHeader == nil && tag.TagType == flv.FLV_TAG_Audio {
		this.audioHeader = tag.Copy()
		return
	}
	if this.videoHeader == nil && tag.TagType == flv.FLV_TAG_Video {
		this.videoHeader = tag.Copy()
		return
	}

	//if idr,new slice
	if tag.TagType == flv.FLV_TAG_Video && tag.Data[0] == 0x17 && tag.Data[1] == 1 {
		this.createNewTSSegment(tag)
	} else {
		this.appendTag(tag)
	}
}

func (this *HLSSource) createNewTSSegment(keyframe *flv.FlvTag) {

	if this.tsCur == nil {
		this.tsCur = &ts.TsCreater{}
		if this.audioHeader != nil {
			this.tsCur.AddTag(this.audioHeader)

			this.audioCur = &audioCache{}
			this.audioCur.Init(this.audioHeader)
		}
		if this.videoHeader != nil {
			this.tsCur.AddTag(this.videoHeader)
		}
		this.tsCur.AddTag(keyframe)

	} else {
		//flush data
		if this.tsCur.GetDuration() < 10000 {
			this.appendTag(keyframe)
			return
		}
		data := this.tsCur.FlushTsList()
		this.muxCache.Lock()
		defer this.muxCache.Unlock()
		if this.tsCache.Len() > TsCacheLength {
			this.tsCache.Remove(this.tsCache.Front())
		}
		tsdata := &hlsTsData{}
		tsdata.durationMs = float64(this.tsCur.GetDuration())
		tsdata.buf = make([]byte, ts.TS_length*data.Len())
		ptr := 0
		for e := data.Front(); e != nil; e = e.Next() {
			copy(tsdata.buf[ptr:], e.Value.([]byte))
			ptr += ts.TS_length
		}
		tsdata.idx = int(this.segIdx & 0xffffffff)
		this.segIdx++
		this.tsCache.PushBack(tsdata)
		this.muxWaits.Lock()
		if this.waitsChannel.Len() > 0 {
			for e := this.waitsChannel.Front(); e != nil; e = e.Next() {
				e.Value.(chan bool) <- true
			}
			this.waitsChannel = list.New()
		}
		this.muxWaits.Unlock()

		if this.segIdx < 10 {
			//if true{
			wssAPI.CreateDirectory("audio")
			fileName := "audio/" + strconv.Itoa(int(this.segIdx-1)) + ".ts"
			logger.LOGD(tsdata.durationMs)
			fp, err := os.Create(fileName)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
			defer fp.Close()
			fp.Write(tsdata.buf)

			if this.audioHeader != nil {

				aacData := this.audioCur.Flush()
				fpaac, _ := os.Create("aac.aac")
				defer fpaac.Close()
				fpaac.Write(aacData)
				this.audioCur.Init(this.audioHeader)
			}
		}

		this.tsCur.Reset()
		//		this.tsCur = &ts.TsCreater{}
		//		if this.videoHeader != nil {
		//			this.tsCur.AddTag(this.videoHeader)
		//		}
		this.tsCur.AddTag(keyframe)

	}
}

func (this *HLSSource) appendTag(tag *flv.FlvTag) {
	if this.tsCur != nil {
		if this.beginTime == 0 && tag.Timestamp > 0 {
			this.beginTime = tag.Timestamp
		}
		tagIn := tag.Copy()
		//tagIn.Timestamp-=this.beginTime
		this.tsCur.AddTag(tagIn)
	}
	if flv.FLV_TAG_Audio == tag.TagType && this.audioCur != nil {
		this.audioCur.AddTag(tag)
	}
}
