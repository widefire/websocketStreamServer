package RTSPService

import (
	"container/list"
	"errors"
	"events/eStreamerEvent"
	"logger"
	"mediaTypes/amf"
	"mediaTypes/flv"
	"mediaTypes/h264"
	"net"
	//	"strconv"
	"strings"
	"sync"
	"time"
	"wssAPI"
)

type RTSPHandler struct {
	conn        net.Conn
	mutexConn   sync.Mutex
	session     string
	streamName  string
	sinkAdded   bool
	sinkRunning bool
	audioHeader *flv.FlvTag
	videoHeader *flv.FlvTag
	isPlaying   bool
	waitPlaying *sync.WaitGroup
	videoCache  *list.List
	mutexVideo  sync.RWMutex
	audioCache  *list.List
	mutexAudio  sync.RWMutex
	tracks      map[string]*trackInfo
	mutexTracks sync.RWMutex
	tcpTimeout  bool //just for vlc(live555) no heart beat
}

type trackInfo struct {
	unicast      bool
	transPort    string
	firstSeq     int
	seq          int
	mark         bool
	RTPStartTime uint32
	trackId      string
	ssrc         uint32
	clockRate    uint32
	byteSend     int64
	pktSend      int64
	RTPChannel   int
	RTPCliPort   int
	RTPSvrPort   int
	RTCPChannel  int
	RTCPCliPort  int
	RTCPSvrPort  int
	RTPCliConn   *net.UDPConn //用来向客户端发送数据
	RTCPCliConn  *net.UDPConn //
	RTPSvrConn   *net.UDPConn //接收客户端的数据
	RTCPSvrConn  *net.UDPConn //
}

func (this *trackInfo) reset() {
	this.unicast = false
	this.transPort = "udp"
	this.firstSeq = 0
	this.seq = 0
	this.RTPStartTime = 0
	this.trackId = ""
	this.ssrc = 0
	this.clockRate = 0
	this.byteSend = 0
	this.pktSend = 0
	this.RTPChannel = 0
	this.RTPCliPort = 0
	this.RTPSvrPort = 0
	this.RTCPChannel = 0
	this.RTCPCliPort = 0
	this.RTCPSvrPort = 0
	if nil != this.RTPCliConn {
		this.RTPCliConn.Close()
	}
	if nil != this.RTCPCliConn {
		this.RTCPCliConn.Close()
	}
	if nil != this.RTPSvrConn {
		this.RTPSvrConn.Close()

	}
	if nil != this.RTCPSvrConn {
		this.RTCPSvrConn.Close()
	}
}

func (this *RTSPHandler) Init(msg *wssAPI.Msg) (err error) {
	this.session = wssAPI.GenerateGUID()
	this.sinkAdded = false
	this.tracks = make(map[string]*trackInfo)
	this.waitPlaying = new(sync.WaitGroup)
	this.tcpTimeout = true
	return
}

func (this *RTSPHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTSPHandler) Stop(msg *wssAPI.Msg) (err error) {
	this.isPlaying = false
	this.waitPlaying.Wait()
	this.delSink()
	return
}

func (this *RTSPHandler) GetType() string {
	return "RTSPHandler"
}

func (this *RTSPHandler) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *RTSPHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_FLV_TAG:
		return this.appendFlvTag(msg)
	case wssAPI.MSG_PLAY_START:
		//这个状态下，rtsp 可以play
		this.sinkRunning = true
	case wssAPI.MSG_PLAY_STOP:
		//如果在play,停止
		this.sinkRunning = false
	default:
		logger.LOGE("msg not processed")
	}
	return
}

func (this *RTSPHandler) appendFlvTag(msg *wssAPI.Msg) (err error) {
	tag := msg.Param1.(*flv.FlvTag)

	if this.audioHeader == nil && tag.TagType == flv.FLV_TAG_Audio {
		this.audioHeader = tag.Copy()
		return
	}
	if this.videoHeader == nil && tag.TagType == flv.FLV_TAG_Video {
		this.videoHeader = tag.Copy()
		return
	}

	if tag.TagType == flv.FLV_TAG_Video {
		this.mutexVideo.Lock()
		if this.videoCache == nil || this.videoCache.Len() > 0xff {
			this.videoCache = list.New()
		}
		this.videoCache.PushBack(tag.Copy())
		this.mutexVideo.Unlock()
	}

	if flv.FLV_TAG_Audio == tag.TagType {
		this.mutexAudio.Lock()
		if this.audioCache == nil || this.audioCache.Len() > 0xff {
			this.audioCache = list.New()
		}
		this.audioCache.PushBack(tag.Copy())
		this.mutexAudio.Unlock()
	}

	return
}

func (this *RTSPHandler) handlePacket(data []byte) (err error) {
	//连接关闭
	if nil == data {
		return this.Stop(nil)
	}
	if '$' == data[0] {
		return this.handleRTPRTCP(data[1:])
	}

	return this.handleRTSP(data)
}

func (this *RTSPHandler) handleRTPRTCP(data []byte) (err error) {
	logger.LOGT("RTP RTCP not processed")
	return
}

func (this *RTSPHandler) handleRTSP(data []byte) (err error) {
	lines := strings.Split(string(data), RTSP_EL)

	//取出方法
	if len(lines) < 1 {
		err = errors.New("rtsp cmd invalid")
		logger.LOGE(err.Error())
		return err
	}
	cmd := ""
	{
		strs := strings.Split(lines[0], " ")
		if len(strs) < 1 {
			logger.LOGE("invalid cmd")
			return errors.New("invalid cmd")
		}
		cmd = strs[0]
	}
	//处理每个方法
	switch cmd {
	case RTSP_METHOD_OPTIONS:
		return this.serveOptions(lines)
	case RTSP_METHOD_DESCRIBE:
		return this.serveDescribe(lines)
	case RTSP_METHOD_SETUP:
		return this.serveSetup(lines)
	case RTSP_METHOD_PLAY:
		return this.servePlay(lines)
	case RTSP_METHOD_PAUSE:
		return this.servePause(lines)
	default:
		logger.LOGE("method " + cmd + " not support now")
		return this.sendErrorReply(lines, 551)
	}
	return
}

func (this *RTSPHandler) send(data []byte) (err error) {
	this.mutexConn.Lock()
	defer this.mutexConn.Unlock()
	_, err = wssAPI.TcpWriteTimeOut(this.conn, data, serviceConfig.TimeoutSec)
	return
}

func (this *RTSPHandler) addSink() bool {
	if true == this.sinkAdded {
		logger.LOGE("sink not deleted")
		return false
	}
	taskAddSink := &eStreamerEvent.EveAddSink{}
	taskAddSink.StreamName = this.streamName
	taskAddSink.SinkId = this.session
	taskAddSink.Sinker = this
	err := wssAPI.HandleTask(taskAddSink)
	if err != nil {
		logger.LOGE(err.Error())
		return false
	}
	this.sinkAdded = true
	return true
}

func (this *RTSPHandler) delSink() {
	if false == this.sinkAdded {
		return
	}
	taskDelSink := &eStreamerEvent.EveDelSink{}
	taskDelSink.StreamName = this.streamName
	taskDelSink.SinkId = this.session
	wssAPI.HandleTask(taskDelSink)
	this.sinkAdded = false
}

func (this *RTSPHandler) threadPlay() {
	this.isPlaying = true
	this.mutexTracks.RLock()
	defer this.mutexTracks.RUnlock()
	this.waitPlaying.Add(1)
	defer func() {
		logger.LOGT("set play to false")
		this.isPlaying = false
		for _, v := range this.tracks {
			v.reset()
		}
		this.waitPlaying.Done()
	}()
	chList := list.New()
	for _, v := range this.tracks {
		ch := make(chan int)
		if v.transPort == "udp" {
			go this.threadUdp(ch, v)
		} else {
			go this.threadTCP(ch, v)
		}
		chList.PushBack(ch)
	}

	for v := chList.Front(); v != nil; v = v.Next() {
		ch := v.Value.(chan int)
		logger.LOGT(<-ch)
	}
	logger.LOGT("all ch end")
}

func (this *RTSPHandler) threadUdp(ch chan int, track *trackInfo) {
	waitRTP := new(sync.WaitGroup)
	waitRTCP := new(sync.WaitGroup)
	defer func() {
		waitRTP.Wait()
		waitRTCP.Wait()
		close(ch)
		logger.LOGD("thread udp end")
	}()
	//监听RTP和RTCP服务端端口，等待穿透数据
	waitRTP.Add(1)
	waitRTCP.Add(1)
	chRtpCli := make(chan *net.UDPAddr)
	chRtcpCli := make(chan *net.UDPAddr)
	defer func() {
		close(chRtpCli)
		close(chRtcpCli)
	}()
	go listenRTP(track, this, waitRTP, chRtpCli)
	go listenRTCP(track, this, waitRTCP, chRtcpCli)
	//rtp
	select {
	case addr := <-chRtpCli:
		var err error
		track.RTPCliConn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		defer track.RTPCliConn.Close()
	case <-time.After(5 * time.Second):
		logger.LOGE("no rtp net data recved")
		return
	}
	//rtcp
	select {
	case addr := <-chRtcpCli:
		var err error
		track.RTCPCliConn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		defer track.RTCPCliConn.Close()
	case <-time.After(5 * time.Second):
		logger.LOGE("no rtcp net data recved")
		return
	}
	logger.LOGT("get cli conn")
	defer func() {
		//关闭连接，以触发关闭监听
		logger.LOGD("close rtp rtcp server conn")
		track.RTPSvrConn.Close()
		track.RTCPSvrConn.Close()
	}()
	//发送数据
	beginSend := false
	if track.trackId == ctrl_track_video {

		//清空之前累计的亢余数据
		this.mutexVideo.Lock()
		this.videoCache = list.New()
		this.mutexVideo.Unlock()
	} else if track.trackId == ctrl_track_audio {
		this.mutexAudio.Lock()
		this.audioCache = list.New()
		this.mutexAudio.Unlock()
	}
	beginTime := uint32(0)
	audioBeginTime := uint32(0)
	logger.LOGT(track.trackId)
	track.seq = track.firstSeq
	lastRTCPTime := time.Now().Second()
	for this.isPlaying {
		if time.Now().Second()-lastRTCPTime > 10 {
			lastRTCPTime = time.Now().Second()

			this.mutexVideo.Lock()
			if this.videoCache == nil || this.videoCache.Len() == 0 {
				this.mutexVideo.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}
			tag := this.videoCache.Front().Value.(*flv.FlvTag).Copy()
			this.mutexVideo.Unlock()
			timestamp := tag.Timestamp - beginTime
			if tag.Timestamp < beginTime {
				timestamp = 0
			}
			tmp64 := int64(track.clockRate / 1000)
			timestamp = uint32((tmp64 * int64(timestamp)) & 0xffffffff)
			this.sendRTCP(track, timestamp)
		}
		//如果是视频 等到有关键帧时开始发送，如果是音频，直接发送
		if ctrl_track_video == track.trackId {
			if false == beginSend {
				//等待关键帧
				beginSend, beginTime = getH264Keyframe(this.videoCache, this.mutexVideo)
				if false == beginSend {
					time.Sleep(30 * time.Millisecond)
					continue
				}
				//把头加回去
				//				if this.videoHeader != nil {
				//					this.mutexVideo.Lock()
				//					this.videoCache.PushFront(this.videoHeader)
				//					this.mutexVideo.Unlock()
				//				}
			}
			//发送数据
			this.mutexVideo.Lock()
			if this.videoCache == nil || this.videoCache.Len() == 0 {
				this.mutexVideo.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}
			tag := this.videoCache.Front().Value.(*flv.FlvTag).Copy()
			this.videoCache.Remove(this.videoCache.Front())
			this.mutexVideo.Unlock()
			err := this.sendFlvH264(track, tag, beginTime)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		} else if ctrl_track_audio == track.trackId {
			//			logger.LOGT("audio not processed now")
			this.mutexAudio.Lock()
			if this.audioCache == nil || this.audioCache.Len() == 0 {
				this.mutexAudio.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}
			tag := this.audioCache.Front().Value.(*flv.FlvTag).Copy()
			this.audioCache.Remove(this.audioCache.Front())
			this.mutexAudio.Unlock()
			if audioBeginTime == 0 {
				audioBeginTime = tag.Timestamp
			}
			if false == beginSend {
				beginSend = true
				err := this.sendFlvAudio(track, this.audioHeader, audioBeginTime)
				if err != nil {
					logger.LOGE(err.Error())
					return
				}
			}
			err := this.sendFlvAudio(track, tag, audioBeginTime)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		} else {
			logger.LOGE(track.trackId + " wrong")
			return
		}

	}
}

func (this *RTSPHandler) threadTCP(ch chan int, track *trackInfo) {
	defer func() {
		close(ch)
	}()
	beginSend := true
	if track.trackId == ctrl_track_video {
		beginSend = false

		//清空之前累计的亢余数据
		this.mutexVideo.Lock()
		this.videoCache = list.New()
		this.mutexVideo.Unlock()
	} else if track.trackId == ctrl_track_audio {
		this.mutexAudio.Lock()
		this.audioCache = list.New()
		this.mutexAudio.Unlock()
	}
	beginTime := uint32(0)
	audioBeginTime := uint32(0)
	logger.LOGT(track.trackId)
	track.seq = track.firstSeq
	for this.isPlaying {
		if ctrl_track_video == track.trackId {
			if false == beginSend {
				//等待关键帧
				beginSend, beginTime = getH264Keyframe(this.videoCache, this.mutexVideo)
				if false == beginSend {
					time.Sleep(30 * time.Millisecond)
					continue
				}
			}
			//发送数据
			this.mutexVideo.Lock()
			if this.videoCache == nil || this.videoCache.Len() == 0 {
				this.mutexVideo.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}
			tag := this.videoCache.Front().Value.(*flv.FlvTag).Copy()
			this.videoCache.Remove(this.videoCache.Front())
			this.mutexVideo.Unlock()
			err := this.sendFlvH264(track, tag, beginTime)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		} else if ctrl_track_audio == track.trackId {
			//			logger.LOGT("audio not processed now")
			this.mutexAudio.Lock()
			if this.audioCache == nil || this.audioCache.Len() == 0 {
				this.mutexAudio.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}
			tag := this.audioCache.Front().Value.(*flv.FlvTag).Copy()
			this.audioCache.Remove(this.audioCache.Front())
			this.mutexAudio.Unlock()
			if audioBeginTime == 0 {
				audioBeginTime = tag.Timestamp
			}
			if false == beginSend {
				beginSend = true
				err := this.sendFlvAudio(track, this.audioHeader, audioBeginTime)
				if err != nil {
					logger.LOGE(err.Error())
					return
				}
			}
			err := this.sendFlvAudio(track, tag, audioBeginTime)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}

		} else {
			logger.LOGE(track.trackId + " wrong")
			return
		}
	}
}

func (this *RTSPHandler) sendFlvH264(track *trackInfo, tag *flv.FlvTag, beginSend uint32) (err error) {
	if "udp" == track.transPort {
		pkts := this.generateH264RTPPackets(tag, beginSend, track)
		if nil == pkts {
			return
		}

		for v := pkts.Front(); v != nil; v = v.Next() {
			//data := v.Value.([]byte)
			//logger.LOGF(data)
			_, err = track.RTPCliConn.Write(v.Value.([]byte))
			track.pktSend++
			track.byteSend += int64(len(v.Value.([]byte)))
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		}
	} else {
		pkts := this.generateH264RTPPackets(tag, beginSend, track)
		if nil == pkts {
			return
		}

		for v := pkts.Front(); v != nil; v = v.Next() {
			pktData := v.Value.([]byte)
			dataSize := len(pktData)
			{
				tcpHeader := make([]byte, 4)
				tcpHeader[0] = '$'
				tcpHeader[1] = byte(track.RTPChannel)
				tcpHeader[2] = byte(dataSize >> 8)
				tcpHeader[3] = byte(dataSize & 0xff)
				track.byteSend += 4
				this.send(tcpHeader)
			}

			err = this.send(pktData)
			track.pktSend++
			track.byteSend += int64(dataSize)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		}
	}
	return
}

func (this *RTSPHandler) sendFlvAudio(track *trackInfo, tag *flv.FlvTag, beginSend uint32) (err error) {
	var pkts *list.List
	switch tag.Data[0] >> 4 {
	case flv.SoundFormat_AAC:
		pkts = this.generateAACRTPPackets(tag, beginSend, track)
	case flv.SoundFormat_MP3:
		pkts = this.generateMP3RTPPackets(tag, beginSend, track)
	default:
		logger.LOGW("audio type not support now")
		return
	}
	//send data
	if pkts == nil {
		return
	}

	for v := pkts.Front(); v != nil; v = v.Next() {

		if "udp" == track.transPort {
			_, err = track.RTPCliConn.Write(v.Value.([]byte))
			track.pktSend++
			track.byteSend += int64(len(v.Value.([]byte)))
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		} else {
			pktData := v.Value.([]byte)
			dataSize := len(pktData)
			{
				tcpHeader := make([]byte, 4)
				tcpHeader[0] = '$'
				tcpHeader[1] = byte(track.RTPChannel)
				tcpHeader[2] = byte(dataSize >> 8)
				tcpHeader[3] = byte(dataSize & 0xff)
				track.byteSend += 4
				this.send(tcpHeader)
			}

			err = this.send(pktData)
			track.pktSend++
			track.byteSend += int64(dataSize)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		}
	}
	return
}

func (this *RTSPHandler) sendRTCP(track *trackInfo, rtpTime uint32) {
	if "udp" == track.transPort {
		data := createSR(track.ssrc, rtpTime, uint32(track.pktSend), uint32(track.byteSend))
		_, err := track.RTCPCliConn.Write(data)
		if err != nil {
			logger.LOGE(err.Error())
		}
	} else {

	}
}

//packetization-mode 1
func (this *RTSPHandler) generateH264RTPPackets(tag *flv.FlvTag, beginTime uint32, track *trackInfo) (rtpPkts *list.List) {
	payLoadSize := RTP_MTU
	if track.transPort == "tcp" {
		payLoadSize -= 4
	}
	//忽略AVC

	if tag.Data[0] == 0x17 && tag.Data[1] == 0x0 {
		return
	}
	//frame类型：1  avc type:1 compositionTime:3 nalsize:4
	//可能有多个nal
	cur := 5
	rtpPkts = list.New()
	for cur < len(tag.Data) {
		nalSize, _ := amf.AMF0DecodeInt32(tag.Data[cur:])
		cur += 4
		nalData := tag.Data[cur : cur+int(nalSize)]
		cur += int(nalSize)
		nalType := nalData[0] & 0xf
		//忽略sps pps
		if nalType == h264.Nal_type_sps || nalType == h264.Nal_type_pps {
			continue
		}

		//计算RTP时间
		timestamp := tag.Timestamp - beginTime
		if tag.Timestamp < beginTime {
			timestamp = 0
		}
		//rtptime/timestamp=rate/1000  rtptime=rate*timestamp/1000
		tmp64 := int64(track.clockRate / 1000)
		timestamp = uint32((tmp64 * int64(timestamp)) & 0xffffffff)
		//关键帧前面加sps pps
		if nalType == h264.Nal_type_idr {
			sps, pps := h264.GetSpsPpsFromAVC(this.videoHeader.Data[5:])

			stapA := &amf.AMF0Encoder{}
			stapA.Init()
			{
				//header
				track.seq++
				if track.seq > 0xffff {
					track.seq = 0
				}
				headerData := createRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc)
				stapA.AppendByteArray(headerData)
				//start bit
				Nri := ((sps[0] & 0x60) >> 5)
				Type := byte(NAL_TYPE_STAP_A)
				stapA.AppendByte(((Nri << 5) | Type))
				//sps size
				stapA.EncodeInt16(int16(len(sps)))
				//sps
				stapA.AppendByteArray(sps)
				//pps size
				stapA.EncodeInt16(int16(len(pps)))
				//pps
				stapA.AppendByteArray(pps)
				pktData, _ := stapA.GetData()
				rtpPkts.PushBack(pktData)
			}
		}
		//帧数据
		{
			payLoadSize -= 13 //12 rtp header,1 f nri type
			//单一包
			if nalSize < uint32(payLoadSize) {
				single := &amf.AMF0Encoder{}
				single.Init()
				track.seq++
				if track.seq > 0xffff {
					track.seq = 0
				}
				headerData := createRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc)
				single.AppendByteArray(headerData)
				single.AppendByteArray(nalData)
				pktData, _ := single.GetData()
				rtpPkts.PushBack(pktData)
			} else {
				//分片包,使用FU_A包
				payLoadSize -= 1 //FU header
				var FU_S, FU_E, FU_R, FU_Type, Nri, Type byte
				Nri = ((nalData[0] & 0x60) >> 5)
				Type = NAL_TYPE_FU_A
				FU_Type = nalType
				count := int(nalSize) / payLoadSize
				if count*payLoadSize < int(nalSize) {
					count++
				}
				//first frame
				curFrame := 0
				curNalData := 1 //nal 的第一个字节的帧类型信息放到fh_header 里面
				{
					FU_S = 1
					FU_E = 0
					FU_R = 0
					fua := amf.AMF0Encoder{}
					fua.Init()
					track.seq++
					if track.seq > 0xffff {
						track.seq = 0
					}
					headerData := createRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc)
					fua.AppendByteArray(headerData)
					fua.AppendByte((Nri << 5) | Type)
					fua.AppendByte((FU_S << 7) | (FU_E << 6) | (FU_R << 5) | FU_Type)
					fua.AppendByteArray(nalData[curNalData : payLoadSize+curNalData])
					curNalData += payLoadSize
					pktData, _ := fua.GetData()
					rtpPkts.PushBack(pktData)
					curFrame++
				}

				//mid frame
				{
					FU_S = 0
					FU_E = 0
					FU_R = 0
					for curFrame+1 < count {
						track.seq++
						if track.seq > 0xffff {
							track.seq = 0
						}
						fua := amf.AMF0Encoder{}
						fua.Init()
						headerData := createRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc)
						fua.AppendByteArray(headerData)
						fua.AppendByte((Nri << 5) | Type)
						fua.AppendByte((FU_S << 7) | (FU_E << 6) | (FU_R << 5) | FU_Type)
						fua.AppendByteArray(nalData[curNalData : payLoadSize+curNalData])
						curNalData += payLoadSize
						pktData, _ := fua.GetData()
						rtpPkts.PushBack(pktData)
						curFrame++
					}
				}
				//last frame
				lastFrameSize := int(nalSize) - curNalData
				if lastFrameSize > 0 {
					FU_S = 0
					FU_E = 1
					FU_R = 0
					track.seq++
					if track.seq > 0xffff {
						track.seq = 0
					}
					fua := amf.AMF0Encoder{}
					fua.Init()
					headerData := createRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc)
					fua.AppendByteArray(headerData)
					fua.AppendByte((Nri << 5) | Type)
					fua.AppendByte((FU_S << 7) | (FU_E << 6) | (FU_R << 5) | FU_Type)
					fua.AppendByteArray(nalData[curNalData:])
					curNalData += payLoadSize
					pktData, _ := fua.GetData()
					rtpPkts.PushBack(pktData)
					curFrame++
				}
			}
		}
	}

	return
}

func (this *RTSPHandler) stopPlayThread() {
	this.isPlaying = false
	this.waitPlaying.Wait()
}

/*
2byte au-header-length

*/
func (this *RTSPHandler) generateAACRTPPackets(tag *flv.FlvTag, beginTime uint32, track *trackInfo) (rtpPkts *list.List) {
	payloadSize := RTP_MTU - 4
	if "tcp" == track.transPort {
		payloadSize -= 4
	}

	dataSize := len(tag.Data) - 2
	if dataSize < 4 {
		return
	}

	au_header := make([]byte, 4)
	au_header[0] = 0x00
	au_header[1] = 0x10
	au_header[2] = byte((dataSize & 0x1fe0) >> 5)
	au_header[3] = byte((dataSize & 0x1f) << 3)
	timestamp := tag.Timestamp - beginTime
	if tag.Timestamp < beginTime {
		timestamp = 0
	}
	//rtptime/timestamp=rate/1000  rtptime=rate*timestamp/1000
	tmp64 := int64(track.clockRate / 1000)
	timestamp = uint32((tmp64 * int64(timestamp)) & 0xffffffff)
	//12 rtp header
	payloadSize -= 12
	cur := 2

	rtpPkts = list.New()
	for dataSize > payloadSize {
		logger.LOGF("big aac")
		tmp := &amf.AMF0Encoder{}
		tmp.Init()
		headerData := createRTPHeaderAAC(Payload_h264, uint32(track.seq), timestamp, track.ssrc)

		track.seq++
		if track.seq > 0xffff {
			track.seq = 0
		}
		tmp.AppendByteArray(headerData)
		tmp.AppendByteArray(au_header)
		tmp.AppendByteArray(tag.Data[cur : cur+payloadSize])
		pktData, _ := tmp.GetData()
		rtpPkts.PushBack(pktData)
		dataSize -= payloadSize
		cur += payloadSize
	}
	//剩下的数据
	if dataSize > 0 {
		single := &amf.AMF0Encoder{}
		single.Init()
		headerData := createRTPHeaderAAC(Payload_h264, uint32(track.seq), timestamp, track.ssrc)
		track.seq++
		if track.seq > 0xffff {
			track.seq = 0
		}
		single.AppendByteArray(headerData)
		single.AppendByteArray(au_header)
		single.AppendByteArray(tag.Data[cur:])
		pktData, _ := single.GetData()
		rtpPkts.PushBack(pktData)
	}
	return
}
