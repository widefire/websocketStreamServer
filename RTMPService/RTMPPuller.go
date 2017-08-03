package RTMPService

import (
	"container/list"
	"errors"
	"events/eLiveListCtrl"
	"events/eRTMPEvent"
	"events/eStreamerEvent"
	"fmt"
	"logger"
	"math/rand"
	"mediaTypes/flv"
	"net"
	"strconv"
	"sync"
	"time"
	"wssAPI"
)

type RTMPPuller struct {
	rtmp       RTMP
	parent     wssAPI.Obj
	src        wssAPI.Obj
	pullParams *eRTMPEvent.EvePullRTMPStream
	waitRead   *sync.WaitGroup
	reading    bool
	srcId      int64
	chValid    bool
	metaDatas  *list.List
}

func PullRTMPLive(task *eRTMPEvent.EvePullRTMPStream) {
	puller := &RTMPPuller{}
	msg := &wssAPI.Msg{}
	msg.Param1 = task
	puller.Init(msg)
	puller.Start(nil)
}

func (this *RTMPPuller) Init(msg *wssAPI.Msg) (err error) {
	this.pullParams = msg.Param1.(*eRTMPEvent.EvePullRTMPStream).Copy()
	this.initRTMPLink()
	this.waitRead = new(sync.WaitGroup)
	this.chValid = true
	this.metaDatas = list.New()
	return
}

func (this *RTMPPuller) closeCh() {
	if this.chValid {
		this.chValid = false
		logger.LOGD("close ch")
		close(this.pullParams.Src)
	}
}

func (this *RTMPPuller) initRTMPLink() {
	this.rtmp.Link.Protocol = this.pullParams.Protocol
	this.rtmp.Link.App = this.pullParams.App
	this.rtmp.Link.Path = this.pullParams.StreamName
	this.rtmp.Link.TcUrl = this.pullParams.Protocol + "://" +
		this.pullParams.Address + ":" +
		strconv.Itoa(this.pullParams.Port) + "/" +
		this.pullParams.App
	//if len(this.pullParams.Instance)>0{
	//	this.rtmp.Link.TcUrl+="/"+this.pullParams.Instance
	//}
	logger.LOGD(this.rtmp.Link.TcUrl)
}

func (this *RTMPPuller) Start(msg *wssAPI.Msg) (err error) {
	defer func() {
		if err != nil {
			logger.LOGE("start failed")
			this.closeCh()
			if nil != this.rtmp.Conn {
				this.rtmp.Conn.Close()
				this.rtmp.Conn = nil
			}
		}
	}()
	//start pull
	//connect
	addr := this.pullParams.Address + ":" + strconv.Itoa(this.pullParams.Port)

	conn, err := net.Dial("tcp", addr)
	logger.LOGT(addr)
	if err != nil {
		logger.LOGE("connect failed:" + err.Error())
		return
	}
	this.rtmp.Init(conn)
	//just simple handshake
	err = this.handleShake()
	if err != nil {
		logger.LOGE("handle shake failed")
		return
	}
	//start read thread
	this.rtmp.BytesIn = 3073
	go this.threadRead()
	//play
	err = this.play()
	if err != nil {
		logger.LOGE("play failed")
		return
	}
	return
}

func (this *RTMPPuller) Stop(msg *wssAPI.Msg) (err error) {
	//stop pull
	logger.LOGT("stop puller")
	this.reading = false
	this.waitRead.Wait()

	if wssAPI.InterfaceValid(this.rtmp.Conn) {
		this.rtmp.Conn.Close()
		this.rtmp.Conn = nil
	}
	//del src
	if wssAPI.InterfaceValid(this.src) {
		taskDelSrc := &eStreamerEvent.EveDelSource{}
		taskDelSrc.StreamName = this.pullParams.SourceName
		taskDelSrc.Id = this.srcId
		err = wssAPI.HandleTask(taskDelSrc)
		if err != nil {
			logger.LOGE(err.Error())
		}
		this.src = nil
	}

	return
}

func (this *RTMPPuller) handleShake() (err error) {
	randomSize := 1528
	//send c0
	conn := this.rtmp.Conn
	c0 := make([]byte, 1)
	c0[0] = 3
	_, err = wssAPI.TcpWriteTimeDuration(conn, c0, time.Duration(serviceConfig.TimeoutSec)*time.Second)
	if err != nil {
		logger.LOGE("send c0 failed")
		return
	}
	//send c1
	c1 := make([]byte, randomSize+4+4)
	for idx := 8; idx < len(c1); idx++ {
		c1[idx] = byte(rand.Intn(255))
	}
	_, err = wssAPI.TcpWriteTimeDuration(conn, c1, time.Duration(serviceConfig.TimeoutSec)*time.Second)
	if err != nil {
		logger.LOGE("send c1 failed")
		return
	}
	//read s0
	s0, err := wssAPI.TcpReadTimeDuration(conn, 1, time.Duration(serviceConfig.TimeoutSec)*time.Second)
	if err != nil {
		logger.LOGE("read s0 failed")
		return
	}
	logger.LOGT(s0)
	//read s1
	s1, err := wssAPI.TcpReadTimeDuration(conn, randomSize+8, time.Duration(serviceConfig.TimeoutSec)*time.Second)
	if err != nil {
		logger.LOGE("read s1 failed")
		return
	}
	//send c2
	_, err = wssAPI.TcpWriteTimeDuration(conn, s1, time.Duration(serviceConfig.TimeoutSec)*time.Second)
	if err != nil {
		logger.LOGE("send c2 failed")
		return
	}
	//read s2
	s2, err := wssAPI.TcpReadTimeDuration(conn, randomSize+8, time.Duration(serviceConfig.TimeoutSec)*time.Second)
	if err != nil {
		logger.LOGE("read s2 failed")
		return
	}
	for idx := 0; idx < len(s2); idx++ {
		if c1[idx] != s2[idx] {
			logger.LOGE("invalid s2")
			return errors.New("invalid s2")
		}
	}
	logger.LOGT("handleshake ok")
	return
}

func (this *RTMPPuller) GetType() string {
	return rtmpTypePuller
}

func (this *RTMPPuller) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *RTMPPuller) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_SourceClosed_Force:
		logger.LOGT("rtmp puller data sink closed")
		this.src = nil
		this.reading = false
	default:
		logger.LOGE(msg.Type + " not processed")
	}
	return
}

func (this *RTMPPuller) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}

func (this *RTMPPuller) HandleControl(pkt *RTMPPacket) (err error) {
	ctype, err := AMF0DecodeInt16(pkt.Body)
	if err != nil {
		return
	}
	switch ctype {
	case RTMP_CTRL_streamBegin:
		streamId, _ := AMF0DecodeInt32(pkt.Body[2:])
		logger.LOGT(fmt.Sprintf("stream begin:%d", streamId))
	case RTMP_CTRL_streamEof:
		streamId, _ := AMF0DecodeInt32(pkt.Body[2:])
		logger.LOGT(fmt.Sprintf("stream eof:%d", streamId))
		err = errors.New("stream eof ")
	case RTMP_CTRL_streamDry:
		streamId, _ := AMF0DecodeInt32(pkt.Body[2:])
		logger.LOGT(fmt.Sprintf("stream dry:%d", streamId))
	case RTMP_CTRL_setBufferLength:
		streamId, _ := AMF0DecodeInt32(pkt.Body[2:])
		buffMS, _ := AMF0DecodeInt32(pkt.Body[6:])
		this.rtmp.buffMS = uint32(buffMS)
		this.rtmp.StreamId = uint32(streamId)
		//logger.LOGI(fmt.Sprintf("set buffer length --streamid:%d--buffer length:%d", this.StreamId, this.buffMS))
	case RTMP_CTRL_streamIsRecorded:
		streamId, _ := AMF0DecodeInt32(pkt.Body[2:])
		logger.LOGT(fmt.Sprintf("stream %d is recorded", streamId))
	case RTMP_CTRL_pingRequest:
		timestamp, _ := AMF0DecodeInt32(pkt.Body[2:])
		this.rtmp.pingResponse(timestamp)
		logger.LOGT(fmt.Sprintf("ping :%d", timestamp))
	case RTMP_CTRL_pingResponse:
		timestamp, _ := AMF0DecodeInt32(pkt.Body[2:])
		logger.LOGF(fmt.Sprintf("pong :%d", timestamp))
	case RTMP_CTRL_streamBufferEmpty:
		//logger.LOGT(fmt.Sprintf("buffer empty"))
	case RTMP_CTRL_streamBufferReady:
		//logger.LOGT(fmt.Sprintf("buffer ready"))
	default:
		logger.LOGI(fmt.Sprintf("rtmp control type:%d not processed", ctype))
	}
	return
}

func (this *RTMPPuller) threadRead() {
	this.reading = true
	this.waitRead.Add(1)
	defer func() {
		this.waitRead.Done()
		this.Stop(nil)
		this.closeCh()
		logger.LOGT("stop read,close conn")
	}()
	for this.reading {
		packet, err := this.readRTMPPkt()
		if err != nil {
			logger.LOGE(err.Error())
			this.reading = false
			return
		}
		switch packet.MessageTypeId {
		case RTMP_PACKET_TYPE_CHUNK_SIZE:
			this.rtmp.RecvChunkSize, err = AMF0DecodeInt32(packet.Body)
			logger.LOGT(fmt.Sprintf("chunk size:%d", this.rtmp.RecvChunkSize))
		case RTMP_PACKET_TYPE_CONTROL:
			err = this.rtmp.HandleControl(packet)
		case RTMP_PACKET_TYPE_BYTES_READ_REPORT:
			logger.LOGT("bytes read report")
		case RTMP_PACKET_TYPE_SERVER_BW:
			this.rtmp.AcknowledgementWindowSize, err = AMF0DecodeInt32(packet.Body)
			logger.LOGT(fmt.Sprintf("acknowledgment size %d", this.rtmp.TargetBW))
		case RTMP_PACKET_TYPE_CLIENT_BW:
			this.rtmp.SelfBW, err = AMF0DecodeInt32(packet.Body)
			this.rtmp.LimitType = uint32(packet.Body[4])
			logger.LOGT(fmt.Sprintf("peer band width %d %d ", this.rtmp.SelfBW, this.rtmp.LimitType))
		case RTMP_PACKET_TYPE_FLEX_MESSAGE:
			err = this.handleInvoke(packet)
		case RTMP_PACKET_TYPE_INVOKE:
			err = this.handleInvoke(packet)
		case RTMP_PACKET_TYPE_AUDIO:
			err = this.sendFlvToSrc(packet)
		case RTMP_PACKET_TYPE_VIDEO:
			err = this.sendFlvToSrc(packet)
		case RTMP_PACKET_TYPE_INFO:
			this.metaDatas.PushBack(packet.Copy())
		case RTMP_PACKET_TYPE_FLASH_VIDEO:
			err = this.processAggregation(packet)
		default:
			logger.LOGW(fmt.Sprintf("rtmp packet type %d not processed", packet.MessageTypeId))
		}
		if err != nil {
			this.reading = false
		}
	}
}

func (this *RTMPPuller) sendFlvToSrc(pkt *RTMPPacket) (err error) {
	if wssAPI.InterfaceIsNil(this.src) && pkt.MessageTypeId != flv.FLV_TAG_ScriptData {

		this.CreatePlaySRC()
	}
	if wssAPI.InterfaceValid(this.src) {
		if this.metaDatas.Len() > 0 {
			for e := this.metaDatas.Front(); e != nil; e = e.Next() {
				metaDataPkt := e.Value.(*RTMPPacket).ToFLVTag()
				msg := &wssAPI.Msg{Type: wssAPI.MSG_FLV_TAG, Param1: metaDataPkt}
				err = this.src.ProcessMessage(msg)
				if err != nil {
					logger.LOGE(err.Error())
					this.Stop(nil)
				}
			}
			this.metaDatas = list.New()
		}
		msg := &wssAPI.Msg{}
		msg.Type = wssAPI.MSG_FLV_TAG
		msg.Param1 = pkt.ToFLVTag()
		err = this.src.ProcessMessage(msg)
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

func (this *RTMPPuller) processAggregation(pkt *RTMPPacket) (err error) {
	cur := 0
	firstAggTime := uint32(0xffffffff)
	for cur < len(pkt.Body) {
		flvPkt := &flv.FlvTag{}
		flvPkt.StreamID = 0
		flvPkt.Timestamp = 0
		flvPkt.TagType = pkt.Body[cur]
		pktLength, _ := AMF0DecodeInt24(pkt.Body[cur+1 : cur+4])
		TimeStamp, _ := AMF0DecodeInt24(pkt.Body[cur+4 : cur+7])
		TimeStampExtended := uint32(pkt.Body[7])
		TimeStamp |= (TimeStampExtended << 24)
		if 0xffffffff == firstAggTime {
			firstAggTime = TimeStamp
		}
		flvPkt.Timestamp = pkt.TimeStamp + TimeStamp - firstAggTime
		flvPkt.Data = make([]byte, pktLength)

		copy(flvPkt.Data, pkt.Body[cur+11:cur+11+int(pktLength)])
		cur += 11 + int(pktLength) + 4
		msg := &wssAPI.Msg{}
		msg.Type = wssAPI.MSG_FLV_TAG
		msg.Param1 = flvPkt
		err = this.src.ProcessMessage(msg)
		if err != nil {
			logger.LOGE(fmt.Sprintf("send aggregation pkts failed"))
			return
		}
	}

	return
}

func (this *RTMPPuller) readRTMPPkt() (packet *RTMPPacket, err error) {
	err = this.rtmp.Conn.SetReadDeadline(time.Now().Add(time.Duration(serviceConfig.TimeoutSec) * time.Second))
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer this.rtmp.Conn.SetReadDeadline(time.Time{})
	packet, err = this.rtmp.ReadPacket()
	return
}

func (this *RTMPPuller) play() (err error) {
	//connect
	err = this.rtmp.Connect(false)
	if err != nil {
		logger.LOGE("rtmp connect to play failed")
		return
	}
	return
}

func (this *RTMPPuller) getProp(pkt *RTMPPacket) (amfobj *AMF0Object, err error) {
	if RTMP_PACKET_TYPE_FLEX_MESSAGE == pkt.MessageTypeId {
		amfobj, err = AMF0DecodeObj(pkt.Body[1:])
	} else {
		amfobj, err = AMF0DecodeObj(pkt.Body)
	}
	if err != nil {
		logger.LOGE("recved invalid amf0 object")
		return
	}
	if amfobj.Props.Len() == 0 {
		logger.LOGT(pkt.Body)
		logger.LOGT(string(pkt.Body))
		return nil, errors.New("no props")
	}
	return
}

func (this *RTMPPuller) handleInvoke(pkt *RTMPPacket) (err error) {

	amfobj, err := this.getProp(pkt)
	if err != nil || amfobj == nil {
		logger.LOGE("decode amf failed")
		return
	}
	methodProp := amfobj.Props.Front().Value.(*AMF0Property)
	switch methodProp.Value.StrValue {
	case "_result":
		err = this.handleRTMPResult(amfobj)
	case "onBWDone":
		this.rtmp.SendCheckBW()
		this.rtmp.SendReleaseStream()
		this.rtmp.SendFCPublish()
	case "_onbwcheck":
		err = this.rtmp.SendCheckBWResult(amfobj.AMF0GetPropByIndex(1).Value.NumValue)
	case "onFCPublish":
	case "_error":
	case "_onbwdone":
	case "onStatus":
		code := ""
		level := ""
		desc := ""
		if amfobj.Props.Len() >= 4 {
			objStatus := amfobj.AMF0GetPropByIndex(3).Value.ObjValue
			for e := objStatus.Props.Front(); e != nil; e = e.Next() {
				prop := e.Value.(*AMF0Property)
				switch prop.Name {
				case "code":
					code = prop.Value.StrValue
				case "level":
					level = prop.Value.StrValue
				case "description":
					desc = prop.Value.StrValue
				}
			}
		}
		logger.LOGT(level)
		logger.LOGT(desc)
		//close
		if code == "NetStream.Failed" || code == "NetStream.Play.Failed" ||
			code == "NetStream.Play.StreamNotFound" || code == "NetConnection.Connect.InvalidApp" ||
			code == "NetStream.Publish.Rejected" || code == "NetStream.Publish.Denied" ||
			code == "NetConnection.Connect.Rejected" {
			return errors.New("error and close")
		}
		//stop play
		if code == "NetStream.Play.Complete" || code == "NetStream.Play.Stop" ||
			code == "NetStream.Play.UnpublishNotify" {
			return errors.New("stop play")
		}
		//start play
		if code == "NetStream.Play.Start" || code == "NetStream.Play.PublishNotify" {
			//this.CreatePlaySRC()
			logger.LOGW("start by media data")
		}
	default:
		logger.LOGW(fmt.Sprintf("method %s not processed", methodProp.Value.StrValue))
	}

	return
}

func (this *RTMPPuller) handleRTMPResult(amfobj *AMF0Object) (err error) {
	idx := int32(amfobj.AMF0GetPropByIndex(1).Value.NumValue)
	this.rtmp.mutexMethod.Lock()
	methodRet, ok := this.rtmp.methodCache[idx]
	if ok == true {
		delete(this.rtmp.methodCache, idx)
	}
	this.rtmp.mutexMethod.Unlock()
	if !ok {
		logger.LOGW("method not found")
		return
	}
	switch methodRet {
	case "connect":
		err = this.rtmp.AcknowledgementBW()
		if err != nil {
			logger.LOGE("acknowledgementBW failed")
			return
		}
		err = this.rtmp.CreateStream()
		if err != nil {
			logger.LOGE("createStream failed")
			return
		}
	case "createStream":
		this.rtmp.StreamId = uint32(amfobj.AMF0GetPropByIndex(3).Value.NumValue)
		err = this.rtmp.SendPlay()
		if err != nil {
			logger.LOGE("send play failed")
			return
		}
	case "releaseStream":
	case "FCPublish":
	default:
		logger.LOGE(fmt.Sprintf("%s result not processed", methodRet))
	}
	return
}

func (this *RTMPPuller) CreatePlaySRC() {
	if this.src == nil {
		taskGet := &eStreamerEvent.EveGetSource{}
		taskGet.StreamName = this.pullParams.SourceName
		err := wssAPI.HandleTask(taskGet)
		if err != nil {
			logger.LOGE(err.Error())
		}
		logger.LOGD(this.pullParams.Address)
		if wssAPI.InterfaceValid(taskGet.SrcObj) && taskGet.HasProducer {
			//已经被其他人抢先了
			logger.LOGD("some other pulled this stream:"+taskGet.StreamName, this.pullParams.Address)
			logger.LOGD(taskGet.HasProducer)
			if this.chValid {
				this.pullParams.Src <- taskGet.SrcObj
			}
			this.srcId = 0
			this.reading = false
			return
		}
		taskAdd := &eStreamerEvent.EveAddSource{}
		taskAdd.Producer = this
		taskAdd.StreamName = this.pullParams.SourceName
		taskAdd.RemoteIp = this.rtmp.Conn.RemoteAddr()
		err = wssAPI.HandleTask(taskAdd)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		if wssAPI.InterfaceIsNil(taskAdd.SrcObj) {
			this.closeCh()
			this.reading = false
			return
		}
		this.src = taskAdd.SrcObj
		this.srcId = taskAdd.Id
		this.pullParams.Src <- this.src
		go this.checkPlayerCounts()
		logger.LOGT("add src ok..")
		return
	}
}

func (this *RTMPPuller) checkPlayerCounts() {
	for this.reading && wssAPI.InterfaceValid(this.src) {
		time.Sleep(time.Duration(2) * time.Minute)
		eve := &eLiveListCtrl.EveGetLivePlayerCount{LiveName: this.pullParams.SourceName}

		err := wssAPI.HandleTask(eve)
		if err != nil {
			logger.LOGD(err.Error())
			continue
		}
		if 1 > eve.Count {
			logger.LOGI("no player for this puller ,close itself")
			this.reading = false
			return
		}
	}
}
