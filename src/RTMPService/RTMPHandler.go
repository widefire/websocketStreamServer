package RTMPService

import (
	"container/list"
	"errors"
	"events/eStreamerEvent"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"strings"
	"sync"
	"wssAPI"
)

type RTMPHandler struct {
	parent       wssAPI.Obj
	mutexStatus  sync.RWMutex
	rtmpInstance *RTMP
	source       wssAPI.Obj
	sinke        wssAPI.Obj
	srcAdded     bool
	sinkAdded    bool
	streamName   string
	clientId     string
	playInfo     RTMPPlayInfo
	app          string
	player       rtmpPlayer
	publisher    rtmpPublisher
	srcId        int64
}
type RTMPPlayInfo struct {
	playReset      bool
	playing        bool //true for thread send playing data
	waitPlaying    *sync.WaitGroup
	mutexCache     sync.RWMutex
	cache          *list.List
	audioHeader    *flv.FlvTag
	videoHeader    *flv.FlvTag
	metadata       *flv.FlvTag
	keyFrameWrited bool
	beginTime      uint32
	startTime      float32
	duration       float32
	reset          bool
}

func (this *RTMPHandler) Init(msg *wssAPI.Msg) (err error) {
	this.rtmpInstance = msg.Param1.(*RTMP)
	msgInit := &wssAPI.Msg{}
	msgInit.Param1 = this.rtmpInstance
	this.player.Init(msgInit)
	this.publisher.Init(msgInit)
	return
}

func (this *RTMPHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPHandler) Stop(msg *wssAPI.Msg) (err error) {
	if this.srcAdded {
		taskDelSrc := &eStreamerEvent.EveDelSource{}
		taskDelSrc.StreamName = this.streamName
		taskDelSrc.Id = this.srcId
		wssAPI.HandleTask(taskDelSrc)
		logger.LOGT("del source:" + this.streamName)
		this.srcAdded = false
	}
	if this.sinkAdded {
		taskDelSink := &eStreamerEvent.EveDelSink{}
		taskDelSink.StreamName = this.streamName
		taskDelSink.SinkId = this.clientId
		wssAPI.HandleTask(taskDelSink)
		this.sinkAdded = false
		logger.LOGT("del sinker:" + this.clientId)
	}
	this.player.Stop(msg)
	this.publisher.Stop(msg)
	return
}

func (this *RTMPHandler) GetType() string {
	return rtmpTypeHandler
}

func (this *RTMPHandler) HandleTask(task wssAPI.Task) (err error) {
	if task.Receiver() != this.GetType() {
		logger.LOGW(fmt.Sprintf("invalid task receiver in rtmpHandler :%s", task.Receiver()))
		return errors.New("invalid task")
	}
	return
}

func (this *RTMPHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {

	if msg == nil {
		return errors.New("nil message")
	}
	switch msg.Type {
	case wssAPI.MSG_GetSource_NOTIFY:
		this.sinkAdded = true
	case wssAPI.MSG_GetSource_Failed:
		//发送404
		this.rtmpInstance.CmdStatus("error", "NetStream.Play.StreamNotFound",
			"paly failed", this.streamName, 0, RTMP_channel_Invoke)
	case wssAPI.MSG_SourceClosed_Force:
		this.srcAdded = false
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		err = this.player.appendFlvTag(tag)

	case wssAPI.MSG_PLAY_START:
		this.player.startPlay()
		return
	case wssAPI.MSG_PLAY_STOP:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		this.sourceInvalid()
		return
	case wssAPI.MSG_PUBLISH_START:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		if err != nil {
			logger.LOGE("start publish failed")
			return
		}
		if false == this.publisher.startPublish() {
			logger.LOGE("start publish falied")
			if true == this.srcAdded {
				taskDelSrc := &eStreamerEvent.EveDelSource{}
				taskDelSrc.StreamName = this.streamName
				taskDelSrc.Id = this.srcId
				wssAPI.HandleTask(taskDelSrc)
			}
		}
		return
	case wssAPI.MSG_PUBLISH_STOP:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		if err != nil {
			logger.LOGE("stop publish failed")
			return
		}
		this.publisher.stopPublish()
		return
	default:
		logger.LOGW(fmt.Sprintf("msg type: %s not processed", msg.Type))
		return
	}
	return
}

func (this *RTMPHandler) sourceInvalid() {
	logger.LOGT("stop play,keep sink")
	this.player.stopPlay()
}

func (this *RTMPHandler) HandleRTMPPacket(packet *RTMPPacket) (err error) {
	if nil == packet {
		this.Stop(nil)
		return
	}
	switch packet.MessageTypeId {
	case RTMP_PACKET_TYPE_CHUNK_SIZE:
		this.rtmpInstance.RecvChunkSize, err = AMF0DecodeInt32(packet.Body)
		logger.LOGT(fmt.Sprintf("chunk size:%d", this.rtmpInstance.RecvChunkSize))
	case RTMP_PACKET_TYPE_CONTROL:
		err = this.rtmpInstance.HandleControl(packet)
	case RTMP_PACKET_TYPE_BYTES_READ_REPORT:
		//		logger.LOGT(packet.TimeStamp)
		//		logger.LOGT(packet.Body)
		//		logger.LOGT("bytes read repost")
	case RTMP_PACKET_TYPE_SERVER_BW:
		this.rtmpInstance.TargetBW, err = AMF0DecodeInt32(packet.Body)
		logger.LOGT(fmt.Sprintf("确认窗口大小 %d", this.rtmpInstance.TargetBW))
	case RTMP_PACKET_TYPE_CLIENT_BW:
		this.rtmpInstance.SelfBW, err = AMF0DecodeInt32(packet.Body)
		this.rtmpInstance.LimitType = uint32(packet.Body[4])
		logger.LOGT(fmt.Sprintf("设置对端宽带 %d %d ", this.rtmpInstance.SelfBW, this.rtmpInstance.LimitType))
	case RTMP_PACKET_TYPE_FLEX_MESSAGE:
		err = this.handleInvoke(packet)
	case RTMP_PACKET_TYPE_INVOKE:
		err = this.handleInvoke(packet)
	case RTMP_PACKET_TYPE_AUDIO:
		return this.sendFlvToSrc(packet)
	case RTMP_PACKET_TYPE_VIDEO:
		return this.sendFlvToSrc(packet)
	case RTMP_PACKET_TYPE_INFO:
		return this.sendFlvToSrc(packet)
	default:
		logger.LOGW(fmt.Sprintf("rtmp packet type %d not processed", packet.MessageTypeId))
	}
	return
}

func (this *RTMPHandler) sendFlvToSrc(pkt *RTMPPacket) (err error) {
	if this.publisher.isPublishing() && wssAPI.InterfaceValid(this.source) {
		msg := &wssAPI.Msg{}
		msg.Type = wssAPI.MSG_FLV_TAG
		msg.Param1 = pkt.ToFLVTag()
		err = this.source.ProcessMessage(msg)
		if err != nil {
			logger.LOGE(err.Error())
			this.Stop(nil)
		}
		return
	} else {
		logger.LOGE("bad status")
	}
	return
}

func (this *RTMPHandler) handleInvoke(packet *RTMPPacket) (err error) {
	var amfobj *AMF0Object
	if RTMP_PACKET_TYPE_FLEX_MESSAGE == packet.MessageTypeId {
		amfobj, err = AMF0DecodeObj(packet.Body[1:])
	} else {
		amfobj, err = AMF0DecodeObj(packet.Body)
	}
	if err != nil {
		logger.LOGE("recved invalid amf0 object")
		return
	}
	if amfobj.Props.Len() == 0 {
		logger.LOGT(packet.Body)
		logger.LOGT(string(packet.Body))
		return
	}

	method := amfobj.Props.Front().Value.(*AMF0Property)

	switch method.Value.StrValue {
	case "connect":
		cmdObj := amfobj.AMF0GetPropByIndex(2)
		if cmdObj != nil {
			this.app = cmdObj.Value.ObjValue.AMF0GetPropByName("app").Value.StrValue
			if strings.HasSuffix(this.app, "/") {
				this.app = strings.TrimSuffix(this.app, "/")
			}
		}
		if this.app != serviceConfig.LivePath {
			logger.LOGE(this.app)
			logger.LOGE(serviceConfig.LivePath)
			logger.LOGW("path wrong")
		}
		err = this.rtmpInstance.AcknowledgementBW()
		if err != nil {
			return
		}
		err = this.rtmpInstance.SetPeerBW()
		if err != nil {
			return
		}
		//err = this.rtmpInstance.SetChunkSize(RTMP_better_chunk_size)
		//		if err != nil {
		//			return
		//		}
		err = this.rtmpInstance.OnBWDone()
		if err != nil {
			return
		}
		err = this.rtmpInstance.ConnectResult(amfobj)
		if err != nil {
			return
		}
	case "_checkbw":
		err = this.rtmpInstance.OnBWCheck()
	case "_result":
		this.handle_result(amfobj)
	case "releaseStream":
		//		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		//		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
		//			fmt.Sprintf("Method not found (%s).", "releaseStream"), idx)
	case "FCPublish":
		//		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		//		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
		//			fmt.Sprintf("Method not found (%s).", "FCPublish"), idx)
	case "createStream":
		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		err = this.rtmpInstance.CmdNumberResult(idx, 1.0)
	case "publish":
		//check prop
		if amfobj.Props.Len() < 4 {
			logger.LOGE("invalid props length")
			err = errors.New("invalid amf obj for publish")
			return
		}

		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		//check status
		if true == this.publisher.isPublishing() {
			logger.LOGE("publish on bad status ")
			idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
			err = this.rtmpInstance.CmdError("error", "NetStream.Publish.Denied",
				fmt.Sprintf("can not publish (%s).", "publish"), idx)
			return
		}
		//add to source
		this.streamName = this.app + "/" + amfobj.AMF0GetPropByIndex(3).Value.StrValue
		taskAddSrc := &eStreamerEvent.EveAddSource{}
		taskAddSrc.Producer = this
		taskAddSrc.StreamName = this.streamName
		taskAddSrc.RemoteIp = this.rtmpInstance.Conn.RemoteAddr()
		err = wssAPI.HandleTask(taskAddSrc)
		if err != nil {
			logger.LOGE("add source failed:" + err.Error())
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Publish.BadName",
				fmt.Sprintf("publish %s.", this.streamName), "", 0, RTMP_channel_Invoke)
			this.streamName = ""
			return err
		}
		this.source = taskAddSrc.SrcObj
		this.srcId = taskAddSrc.Id
		if this.source == nil {
			logger.LOGE("add source failed:")
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Publish.BadName",
				fmt.Sprintf("publish %s.", this.streamName), "", 0, RTMP_channel_Invoke)
			this.streamName = ""
			return errors.New("bad name")
		}
		this.srcAdded = true
		this.rtmpInstance.Link.Path = amfobj.AMF0GetPropByIndex(2).Value.StrValue
		if false == this.publisher.startPublish() {
			logger.LOGE("start publish failed:" + this.streamName)
			//streamer.DelSource(this.streamName)
			taskDelSrc := &eStreamerEvent.EveDelSource{}
			taskDelSrc.StreamName = this.streamName
			taskDelSrc.Id = this.srcId
			wssAPI.HandleTask(taskDelSrc)
			return
		}
	case "FCUnpublish":
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
	case "deleteStream":
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
	//do nothing now
	case "play":
		this.streamName = this.app + "/" + amfobj.AMF0GetPropByIndex(3).Value.StrValue
		this.rtmpInstance.Link.Path = this.streamName
		startTime := -2
		duration := -1
		reset := false
		this.playInfo.startTime = -2
		this.playInfo.duration = -1
		this.playInfo.reset = false
		if amfobj.Props.Len() >= 5 {
			this.playInfo.startTime = float32(amfobj.AMF0GetPropByIndex(4).Value.NumValue)
		}
		if amfobj.Props.Len() >= 6 {
			this.playInfo.duration = float32(amfobj.AMF0GetPropByIndex(5).Value.NumValue)
			if this.playInfo.duration < 0 {
				this.playInfo.duration = -1
			}
		}
		if amfobj.Props.Len() >= 7 {
			this.playInfo.reset = amfobj.AMF0GetPropByIndex(6).Value.BoolValue
		}

		//check player status,if playing,error
		if false == this.player.setPlayParams(this.streamName, startTime, duration, reset) {
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Play.Failed",
				"paly failed", this.streamName, 0, RTMP_channel_Invoke)

			return nil
		}
		err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}

		if true == this.playInfo.playReset {
			err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Reset",
				fmt.Sprintf("Playing and resetting %s", this.rtmpInstance.Link.Path),
				this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		}

		err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Start",
			fmt.Sprintf("Started playing %s", this.rtmpInstance.Link.Path), this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}

		this.clientId = wssAPI.GenerateGUID()
		taskAddSink := &eStreamerEvent.EveAddSink{}
		taskAddSink.StreamName = this.streamName
		taskAddSink.SinkId = this.clientId
		taskAddSink.Sinker = this
		err = wssAPI.HandleTask(taskAddSink)
		if err != nil {
			//404
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Play.StreamNotFound",
				"paly failed", this.streamName, 0, RTMP_channel_Invoke)
			return nil
		}
		this.sinkAdded = taskAddSink.Added
	case "_error":
		amfobj.Dump()
	case "closeStream":
		amfobj.Dump()
	default:
		logger.LOGW(fmt.Sprintf("rtmp method <%s> not processed", method.Value.StrValue))
	}
	return
}

func (this *RTMPHandler) handle_result(amfobj *AMF0Object) {
	transactionId := int32(amfobj.AMF0GetPropByIndex(1).Value.NumValue)
	resultMethod := this.rtmpInstance.methodCache[transactionId]
	switch resultMethod {
	case "_onbwcheck":
	default:
		logger.LOGW("result of " + resultMethod + " not processed")
	}
}

func (this *RTMPHandler) startPublishing() (err error) {
	err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return nil
	}
	err = this.rtmpInstance.CmdStatus("status", "NetStream.Publish.Start",
		fmt.Sprintf("publish %s", this.rtmpInstance.Link.Path), "", 0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
		return nil
	}
	this.publisher.startPublish()
	return
}

func (this *RTMPHandler) isPlaying() bool {
	return this.player.IsPlaying()
}

func (this *RTMPHandler) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}
