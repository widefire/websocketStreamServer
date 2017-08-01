package webSocketService

import (
	"container/list"
	"encoding/json"
	"errors"
	"events/eStreamerEvent"
	"fmt"
	"logger"
	"mediaTypes/amf"
	"mediaTypes/flv"
	"mediaTypes/mp4"
	"sync"
	"time"
	"wssAPI"

	"github.com/gorilla/websocket"
)

const (
	wsHandler = "websocketHandler"
)

type websocketHandler struct {
	parent       wssAPI.Obj
	conn         *websocket.Conn
	app          string
	streamName   string
	playName     string
	pubName      string
	clientId     string
	isPlaying    bool
	mutexPlaying sync.RWMutex
	waitPlaying  *sync.WaitGroup
	stPlay       playInfo
	isPublish    bool
	mutexPublish sync.RWMutex
	hasSink      bool
	mutexbSink   sync.RWMutex
	hasSource    bool
	mutexbSource sync.RWMutex
	source       wssAPI.Obj
	sourceIdx    int
	lastCmd      int
	mutexWs      sync.Mutex
}

type playInfo struct {
	cache          *list.List
	mutexCache     sync.RWMutex
	audioHeader    *flv.FlvTag
	videoHeader    *flv.FlvTag
	metadata       *flv.FlvTag
	keyFrameWrited bool
	beginTime      uint32
}

func (this *websocketHandler) Init(msg *wssAPI.Msg) (err error) {
	this.conn = msg.Param1.(*websocket.Conn)
	this.app = msg.Param2.(string)
	this.waitPlaying = new(sync.WaitGroup)
	this.lastCmd = WSC_close
	return
}

func (this *websocketHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) Stop(msg *wssAPI.Msg) (err error) {
	this.doClose()
	return
}

func (this *websocketHandler) GetType() string {
	return wsHandler
}

func (this *websocketHandler) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *websocketHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_GetSource_NOTIFY:
		this.hasSink = true
	case wssAPI.MSG_GetSource_Failed:
		this.hasSink = false
		this.sendWsStatus(this.conn, WS_status_error, NETSTREAM_PLAY_FAILED, 0)
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		err = this.appendFlvTag(tag)
	case wssAPI.MSG_PLAY_START:
		this.startPlay()
	case wssAPI.MSG_PLAY_STOP:
		this.stopPlay()
		logger.LOGT("play stop message")
	case wssAPI.MSG_PUBLISH_START:
	case wssAPI.MSG_PUBLISH_STOP:
	}
	return
}

func (this *websocketHandler) appendFlvTag(tag *flv.FlvTag) (err error) {
	if false == this.isPlaying {
		err = errors.New("websocket client not playing")
		logger.LOGE(err.Error())
		return
	}
	tag = tag.Copy()

	//tag.Timestamp -= this.stPlay.beginTime
	//if false == this.stPlay.keyFrameWrited && tag.TagType == flv.FLV_TAG_Video {
	//	if this.stPlay.videoHeader == nil {
	//		this.stPlay.videoHeader = tag
	//	} else {
	//		if (tag.Data[0] >> 4) == 1 {
	//			this.stPlay.keyFrameWrited = true
	//		} else {
	//			return
	//		}
	//	}
	//
	//}

	if this.stPlay.audioHeader == nil && tag.TagType == flv.FLV_TAG_Audio {
		this.stPlay.audioHeader = tag
		this.stPlay.mutexCache.Lock()
		this.stPlay.cache.PushBack(tag)
		this.stPlay.mutexCache.Unlock()
		return
	}
	if this.stPlay.videoHeader == nil && tag.TagType == flv.FLV_TAG_Video {
		this.stPlay.videoHeader = tag
		this.stPlay.mutexCache.Lock()
		this.stPlay.cache.PushBack(tag)
		this.stPlay.mutexCache.Unlock()
		return
	}
	if false == this.stPlay.keyFrameWrited && tag.TagType == flv.FLV_TAG_Video {
		if false == this.stPlay.keyFrameWrited && ((tag.Data[0] >> 4) == 1) {
			this.stPlay.beginTime = tag.Timestamp
			this.stPlay.keyFrameWrited = true
		}
	}
	if false == this.stPlay.keyFrameWrited {
		return
	}

	tag.Timestamp -= this.stPlay.beginTime
	this.stPlay.mutexCache.Lock()
	defer this.stPlay.mutexCache.Unlock()
	this.stPlay.cache.PushBack(tag)

	return
}

func (this *websocketHandler) processWSMessage(data []byte) (err error) {
	if nil == data || len(data) < 4 {
		this.Stop(nil)
		return
	}
	msgType := int(data[0])
	switch msgType {
	case WS_pkt_audio:
	case WS_pkt_video:
	case WS_pkt_control:
		logger.LOGD("recv control data:")
		logger.LOGD(data)
		return this.controlMsg(data[1:])
	default:
		err = errors.New(fmt.Sprintf("msg type %d not supported", msgType))
		logger.LOGW("invalid binary data")
		return
	}
	return
}

func (this *websocketHandler) controlMsg(data []byte) (err error) {
	if nil == data || len(data) < 4 {
		return errors.New("invalid msg")
	}
	ctrlType, err := amf.AMF0DecodeInt24(data)
	if err != nil {
		logger.LOGE("get ctrl type failed")
		return
	}
	logger.LOGT(ctrlType)
	switch ctrlType {
	case WSC_play:
		return this.ctrlPlay(data[3:])
	case WSC_play2:
		return this.ctrlPlay2(data[3:])
	case WSC_resume:
		return this.ctrlResume(data[3:])
	case WSC_pause:
		return this.ctrlPause(data[3:])
	case WSC_seek:
		return this.ctrlSeek(data[3:])
	case WSC_close:
		return this.ctrlClose(data[3:])
	case WSC_stop:
		return this.ctrlStop(data[3:])
	case WSC_publish:
		return this.ctrlPublish(data[3:])
	case WSC_onMetaData:
		return this.ctrlOnMetadata(data[3:])
	default:
		logger.LOGE("unknowd websocket control type")
		return errors.New("invalid ctrl msg type")
	}
	return
}

func (this *websocketHandler) sendSlice(slice *mp4.FMP4Slice) (err error) {
	dataSend := make([]byte, len(slice.Data)+1)
	dataSend[0] = byte(slice.Type)
	copy(dataSend[1:], slice.Data)
	return this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
}

func (this *websocketHandler) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}

func (this *websocketHandler) addSource(streamName string) (id int, src wssAPI.Obj, err error) {
	taskAddSrc := &eStreamerEvent.EveAddSource{StreamName: streamName}
	taskAddSrc.RemoteIp = this.conn.RemoteAddr()
	err = wssAPI.HandleTask(taskAddSrc)
	if err != nil {
		logger.LOGE("add source " + streamName + " failed")
		return
	}
	this.hasSource = true
	return
}

func (this *websocketHandler) delSource(streamName string, id int) (err error) {
	taskDelSrc := &eStreamerEvent.EveDelSource{StreamName: streamName, Id: int64(id)}
	err = wssAPI.HandleTask(taskDelSrc)
	this.hasSource = false
	if err != nil {
		logger.LOGE("del source " + streamName + " failed:" + err.Error())
		return
	}
	return
}

func (this *websocketHandler) addSink(streamName, clientId string, sinker wssAPI.Obj) (err error) {
	taskAddsink := &eStreamerEvent.EveAddSink{StreamName: streamName, SinkId: clientId, Sinker: sinker}
	err = wssAPI.HandleTask(taskAddsink)
	if err != nil {
		logger.LOGE(fmt.Sprintf("add sink %s %s failed :%s", streamName, clientId, err.Error()))
		return
	}
	this.hasSink = taskAddsink.Added
	return
}

func (this *websocketHandler) delSink(streamName, clientId string) (err error) {
	taskDelSink := &eStreamerEvent.EveDelSink{StreamName: streamName, SinkId: clientId}
	err = wssAPI.HandleTask(taskDelSink)
	this.hasSink = false
	if err != nil {
		logger.LOGE(fmt.Sprintf("del sink %s %s failed:\n%s", streamName, clientId, err.Error()))
	}
	logger.LOGE("del sinker")
	return
}

func (this *playInfo) reset() {
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
	this.cache = list.New()
	this.audioHeader = nil
	this.videoHeader = nil
	this.metadata = nil
	this.keyFrameWrited = false
	this.beginTime = 0
}

func (this *playInfo) addInitPkts() {
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
	if this.audioHeader != nil {
		this.cache.PushBack(this.audioHeader)
	}
	if this.videoHeader != nil {
		this.cache.PushBack(this.videoHeader)
	}
	if this.metadata != nil {
		this.cache.PushBack(this.metadata)
	}
}

func (this *websocketHandler) startPlay() {
	this.stPlay.reset()
	this.isPlaying = true
	go this.threadPlay()
}

func (this *websocketHandler) threadPlay() {
	this.isPlaying = true
	this.waitPlaying.Add(1)
	defer func() {
		this.waitPlaying.Done()
		this.stPlay.reset()
	}()
	fmp4Creater := &mp4.FMP4Creater{}
	for true == this.isPlaying {
		this.stPlay.mutexCache.Lock()
		if this.stPlay.cache == nil || this.stPlay.cache.Len() == 0 {
			this.stPlay.mutexCache.Unlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}
		tag := this.stPlay.cache.Front().Value.(*flv.FlvTag)
		this.stPlay.cache.Remove(this.stPlay.cache.Front())
		this.stPlay.mutexCache.Unlock()
		if WSC_pause == this.lastCmd {
			continue
		}
		if tag.TagType == flv.FLV_TAG_ScriptData {
			err := this.sendWsControl(this.conn, WSC_onMetaData, tag.Data)
			if err != nil {
				logger.LOGE(err.Error())
				this.isPlaying = false
			}
			continue
		}
		slice := fmp4Creater.AddFlvTag(tag)
		if slice != nil {
			err := this.sendFmp4Slice(slice)
			if err != nil {
				logger.LOGE(err.Error())
				this.isPlaying = false
			}
		}
	}
}

func (this *websocketHandler) sendFmp4Slice(slice *mp4.FMP4Slice) (err error) {
	this.mutexWs.Lock()
	defer this.mutexWs.Unlock()
	dataSend := make([]byte, len(slice.Data)+1)
	dataSend[0] = byte(slice.Type)
	copy(dataSend[1:], slice.Data)
	err = this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
	return
}

func (this *websocketHandler) stopPlay() {
	this.isPlaying = false
	this.waitPlaying.Wait()
	this.stPlay.reset()
	this.sendWsStatus(this.conn, WS_status_status, NETSTREAM_PLAY_STOP, 0)
}

func (this *websocketHandler) stopPublish() {
	logger.LOGE("stop publish not code")
}

func (this *websocketHandler) sendWsControl(conn *websocket.Conn, ctrlType int, data []byte) (err error) {
	this.mutexWs.Lock()
	defer this.mutexWs.Unlock()
	dataSend := make([]byte, len(data)+4)
	dataSend[0] = WS_pkt_control
	dataSend[1] = byte((ctrlType >> 16) & 0xff)
	dataSend[2] = byte((ctrlType >> 8) & 0xff)
	dataSend[3] = byte((ctrlType >> 0) & 0xff)
	copy(dataSend[4:], data)
	return conn.WriteMessage(websocket.BinaryMessage, dataSend)
}

func (this *websocketHandler) sendWsStatus(conn *websocket.Conn, level, code string, req int) (err error) {
	this.mutexWs.Lock()
	defer this.mutexWs.Unlock()
	st := &stResult{Level: level, Code: code, Req: req}
	dataJson, err := json.Marshal(st)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	dataSend := make([]byte, len(dataJson)+4)
	dataSend[0] = WS_pkt_control
	dataSend[1] = 0
	dataSend[2] = 0
	dataSend[3] = 0
	copy(dataSend[4:], dataJson)
	err = conn.WriteMessage(websocket.BinaryMessage, dataSend)
	return
}
