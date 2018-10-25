package RTSPService

import (
	"errors"
	"fmt"
	"logger"
	"math/rand"
	"mediaTypes/aac"
	"strconv"
	"strings"
	"time"
)

func (this *RTSPHandler) sendErrorReply(lines []string, code int) (err error) {
	cseq := getCSeq(lines)
	strOut := RTSP_VER + " " + strconv.Itoa(code) + " " + getRTSPStatusByCode(code) + RTSP_EL
	strOut += "CSeq: " + strconv.Itoa(cseq) + RTSP_EL
	strOut += RTSP_EL
	err = this.send([]byte(strOut))
	return
}

func (this *RTSPHandler) serveOptions(lines []string) (err error) {
	cseq := getCSeq(lines)
	cliSession := this.getSession(lines)
	hasSession := false
	if len(cliSession) > 0 && cliSession != this.session {
		hasSession = true
	}
	str := RTSP_VER + " " + strconv.Itoa(200) + " " + getRTSPStatusByCode(200) + RTSP_EL
	str += HDR_CSEQ + ": " + strconv.Itoa(cseq) + RTSP_EL
	if hasSession {
		str += "Session: " + this.session + RTSP_EL
	}
	str += "Public: OPTIONS,DESCRIBE,SETUP,PLAY,PAUSE,TEARDOWN " + RTSP_EL + RTSP_EL
	err = this.send([]byte(str))

	if this.tcpTimeout {
		userAgent := getHeaderByName(lines, "User-Agent:", true)
		if len(userAgent) > 0 {
			this.tcpTimeout = !strings.Contains(userAgent, "LibVLC")
		}
	}
	return
}

func (this *RTSPHandler) serveDescribe(lines []string) (err error) {
	cseq := getCSeq(lines)
	str := removeSpace(lines[0])
	strSpaces := strings.Split(str, " ")
	if len(strSpaces) < 2 {
		return this.sendErrorReply(lines, 455)
	}

	_, this.streamName, err = this.parseUrl(strSpaces[1])
	if err != nil {
		logger.LOGE(err.Error())
		return this.sendErrorReply(lines, 455)
	}

	//添加槽
	if false == this.addSink() {
		err = errors.New("add sink failed:" + this.streamName)
		return
	}
	//看是否需要sdp
	acceptLine := getHeaderByName(lines, "Accept", false)
	needSdp := false
	if len(acceptLine) > 0 {
		if strings.Contains(acceptLine, "application/sdp") {
			needSdp = true
		}
	}
	if needSdp {
		//等待直到有音视频头或超时2s
		counts := 0
		needSdp = false
		for {
			counts++
			if counts > 200 {
				break
			}
			if this.audioHeader != nil && this.videoHeader != nil {
				needSdp = true
				break
			}
			this.mutexVideo.RLock()
			if nil != this.videoCache && this.videoCache.Len() > 0 {
				this.mutexVideo.RUnlock()
				needSdp = true
				break
			}
			this.mutexVideo.RUnlock()
			time.Sleep(10 * time.Millisecond)
		}
	}
	sdp := ""
	if needSdp {
		//生成sdp
		ok := false
		sdp, ok = generateSDP(this.videoHeader.Copy(), this.audioHeader.Copy())
		if false == ok {
			logger.LOGE("generate sdp failed")
			needSdp = false
		}
	}

	strOut := RTSP_VER + " " + strconv.Itoa(200) + " " + getRTSPStatusByCode(200) + RTSP_EL
	strOut += HDR_CSEQ + ": " + strconv.Itoa(cseq) + RTSP_EL

	if needSdp {
		strOut += "Content-Type: application/sdp" + RTSP_EL
		strOut += "Content-Length: " + strconv.Itoa(len(sdp)+len(RTSP_EL)) + RTSP_EL + RTSP_EL
		strOut += sdp + RTSP_EL
	} else {
		//strOut += RTSP_EL
		logger.LOGE("can not generate sdp")
		err = this.sendErrorReply(lines, 415)
		return
	}

	err = this.send([]byte(strOut))
	return
}

func (this *RTSPHandler) parseUrl(url string) (port int, streamName string, err error) {
	if false == strings.HasPrefix(url, "rtsp://") {
		err = errors.New("bad rtsp url:" + url)
		logger.LOGE(err.Error())
		return
	}
	sub := strings.TrimPrefix(url, "rtsp://")
	subs := strings.Split(sub, "/")
	if len(subs) < 2 {
		err = errors.New("bad rtsp url:" + url)
		logger.LOGE(err.Error())
		return
	}
	streamName = strings.TrimPrefix(sub, subs[0])
	streamName = strings.TrimPrefix(streamName, "/")
	if false == strings.Contains(subs[0], ":") {
		port = 554
	} else {
		subs = strings.Split(subs[0], ":")
		if len(subs) != 2 {
			err = errors.New("bad rtsp url:" + url)
			logger.LOGE(err.Error())
			return
		}
		port, err = strconv.Atoi(subs[1])
	}
	return
}

func (this *RTSPHandler) getSession(lines []string) (cliSession string) {
	tmp := getHeaderByName(lines, "Session:", true)
	if len(tmp) > 0 {
		tmp = strings.Replace(tmp, " ", "", -1)
		fmt.Sscanf(tmp, "Session:%s", &cliSession)
		logger.LOGD(cliSession)
		if strings.Contains(cliSession, ";") {
			cliSession = strings.Split(cliSession, ";")[0]
		}
	}
	return
}

func (this *RTSPHandler) serveSetup(lines []string) (err error) {
	cseq := getCSeq(lines)
	if this.isPlaying {
		return this.sendErrorReply(lines, 405)
	}
	cliSession := this.getSession(lines)
	if len(cliSession) > 0 && strings.Compare(cliSession, this.session) != 0 {
		//session错误
		logger.LOGE("session wrong")
		return this.sendErrorReply(lines, 454)
	}
	//取出track
	trackName := ""
	{
		strLine := removeSpace(lines[0])
		subs := strings.Split(strLine, " ")
		if len(subs) < 2 {
			logger.LOGE("setup failed")
			return this.sendErrorReply(lines, 400)
		}
		logger.LOGD(subs)
		subs = strings.Split(subs[1], "/")
		trackName = subs[len(subs)-1]
	}
	if strings.Compare(trackName, ctrl_track_audio) != 0 && strings.Compare(trackName, ctrl_track_video) != 0 {
		logger.LOGE("track :" + trackName + " not found")
		return this.sendErrorReply(lines, 404)
	}

	this.mutexTracks.RLock()
	track, exist := this.tracks[trackName]
	if exist {
		logger.LOGE("track :" + trackName + " has setuped,sotp play ")
		this.stopPlayThread()
	} else {
		track = &trackInfo{}
	}
	this.mutexTracks.RUnlock()
	logger.LOGD("lok track")
	this.mutexTracks.Lock()
	logger.LOGD("unlock track")
	defer this.mutexTracks.Unlock()
	//取出协议类型和端口号
	{

		//video 90000
		if strings.Compare(trackName, ctrl_track_audio) == 0 {
			if this.audioHeader != nil {
				asc := aac.GenerateAudioSpecificConfig(this.audioHeader.Data[2:])
				track.clockRate = uint32(asc.SamplingFrequency)
			}
		}
		//audio 90000
		if strings.Compare(trackName, ctrl_track_video) == 0 {
			track.clockRate = RTP_H264_freq
		}
		strTransport := getHeaderByName(lines, HDR_TRANSPORT, true)
		if len(strTransport) == 0 {
			logger.LOGE("setup failed,no transport")
			return this.sendErrorReply(lines, 400)
		}
		strTransport = removeSpace(strTransport)
		strTransport = strings.TrimPrefix(strTransport, HDR_TRANSPORT)
		strTransport = strings.TrimPrefix(strTransport, ":")
		strTransport = strings.TrimPrefix(strTransport, " ")
		subs := strings.Split(strTransport, ";")

		logger.LOGT(subs)
		if len(subs) != 3 {
			logger.LOGT(len(subs))
			logger.LOGE("setup failed,parse transport failed")
			return this.sendErrorReply(lines, 400)
		}
		track.transPort = subs[0]
		if strings.Compare(subs[1], "unicast") == 0 {
			track.unicast = true
		} else {
			track.unicast = false
		}
		if subs[0] == RTSP_RTP_AVP || subs[0] == RTSP_RTP_AVP_UDP {
			track.transPort = "udp"
			if false == strings.HasPrefix(subs[2], "client_port=") {
				logger.LOGE("udp not found client port")
				return this.sendErrorReply(lines, 461)
			}
			cliports := strings.TrimPrefix(subs[2], "client_port=")
			ports := strings.Split(cliports, "-")
			if len(ports) != 2 {
				logger.LOGE("udp not found client port")
				return this.sendErrorReply(lines, 461)
			}
			track.RTPCliPort, err = strconv.Atoi(ports[0])

			if err != nil {
				logger.LOGE(err.Error())
				return this.sendErrorReply(lines, 461)
			}
			track.RTCPCliPort, err = strconv.Atoi(ports[1])
			if err != nil {
				logger.LOGE(err.Error())
				return this.sendErrorReply(lines, 461)
			}
			ok := false

			track.RTPSvrPort, track.RTCPSvrPort, track.RTPSvrConn, track.RTCPSvrConn, ok = genRTPRTCP()
			if false == ok {
				logger.LOGE(err.Error())
				return this.sendErrorReply(lines, 461)
			}

		} else if subs[0] == RTSP_RTP_AVP_TCP {
			track.transPort = "tcp"
			if false == strings.Contains(strTransport, "interleaved=") {
				logger.LOGE("not found tcp interleaved")
				return this.sendErrorReply(lines, 461)
			}
			strsubs := strings.Split(strTransport, "interleaved=")
			if len(strsubs) != 2 {
				logger.LOGE("not found tcp interleaved")
				return this.sendErrorReply(lines, 461)
			}

			strsubs = strings.Split(strsubs[1], "-")
			if len(strsubs) != 2 {
				logger.LOGE("not found tcp interleaved")
				return this.sendErrorReply(lines, 461)
			}
			track.RTPChannel, err = strconv.Atoi(strsubs[0])
			if err != nil {
				logger.LOGE(err.Error())
				return this.sendErrorReply(lines, 461)
			}
			track.RTCPChannel, err = strconv.Atoi(strsubs[1])
			if err != nil {
				logger.LOGE(err.Error())
				return this.sendErrorReply(lines, 461)
			}
			logger.LOGT(track.RTCPChannel)
		} else {
			logger.LOGE(subs[0] + " not support now")
			return this.sendErrorReply(lines, 551)
		}
	}
	track.trackId = trackName
	this.tracks[trackName] = track
	track.firstSeq = rand.Intn(0xffff)
	track.mark = true
	track.ssrc = uint32(rand.Intn(0xffff))
	//返回结果

	strOut := RTSP_VER + " " + strconv.Itoa(200) + " " + getRTSPStatusByCode(200) + RTSP_EL
	strOut += HDR_CSEQ + ": " + strconv.Itoa(cseq) + RTSP_EL
	strOut += "Server: " + RTSPServerName + RTSP_EL
	strOut += "Session: " + this.session + ";timeout=" + strconv.Itoa(serviceConfig.TimeoutSec) + RTSP_EL
	if track.transPort == "udp" {
		strOut += "Transport: RTP/AVP;unicast;"
		strOut += "client_port=" + strconv.Itoa(track.RTPCliPort) + "-" + strconv.Itoa(track.RTCPCliPort) + ";"
		strOut += "server_port=" + strconv.Itoa(track.RTPSvrPort) + "-" + strconv.Itoa(track.RTCPSvrPort)
		strOut += RTSP_EL
	} else {
		//tcp
		strOut += "Transport: RTP/AVP/TCP;unicast;"
		strOut += "interleaved=" + strconv.Itoa(track.RTPChannel) + "-" + strconv.Itoa(track.RTCPChannel)
		strOut += RTSP_EL
	}
	strOut += RTSP_EL

	return this.send([]byte(strOut))
}

func (this *RTSPHandler) servePlay(lines []string) (err error) {
	errCode := 0
	defer func() {
		if err != nil {
			logger.LOGE(err.Error())
		}
		if errCode != 0 {
			this.sendErrorReply(lines, errCode)
		}
	}()
	//状态
	if len(this.tracks) == 0 {
		err = errors.New("play a playing stream")
		errCode = 405
		return
	}
	if this.isPlaying || len(this.tracks) == 0 {
		err = errors.New("play a playing stream")
		errCode = 405
		return
	}
	//cseq
	cseq := getCSeq(lines)
	//session
	cliSession := this.getSession(lines)
	if cliSession != this.session {
		logger.LOGD(cliSession)
		err = errors.New("session wrong")
		errCode = 454
		return
	}
	//begin time:use default zero
	//start play

	strOut := RTSP_VER + " " + strconv.Itoa(200) + " " + getRTSPStatusByCode(200) + RTSP_EL
	strOut += HDR_CSEQ + ": " + strconv.Itoa(cseq) + RTSP_EL
	strOut += "Session: " + this.session + RTSP_EL
	strOut += "Range: npt=0.0-" + RTSP_EL
	strOut += "RTP-Info: "
	addCmma := false
	line0 := removeSpace(lines[0])
	subs := strings.Split(line0, " ")
	if len(subs) != 3 {
		err = errors.New("play cmd failed")
		errCode = 450
		return
	}
	url := subs[1]
	this.mutexTracks.RLock()
	for _, v := range this.tracks {
		if addCmma {
			strOut += ","
		}
		addCmma = true
		strOut += url + "/" + v.trackId + ";"
		strOut += "seq=" + strconv.Itoa(v.firstSeq) + ";"
		strOut += "rtptime=" + strconv.Itoa(int(v.RTPStartTime))
	}
	strOut += RTSP_EL
	strOut += RTSP_EL
	this.mutexTracks.RUnlock()

	err = this.send([]byte(strOut))
	go this.threadPlay()
	return
}

func (this *RTSPHandler) servePause(lines []string) (err error) {
	cseq := getCSeq(lines)
	errCode := 0
	defer func() {
		if err != nil {
			logger.LOGE(err.Error())
		}
		if errCode != 0 {
			this.sendErrorReply(lines, errCode)
		}
	}()
	//session
	cliSession := this.getSession(lines)
	if cliSession != this.session {
		logger.LOGD(cliSession)
		err = errors.New("session wrong")
		errCode = 454
		return
	}
	if this.isPlaying == false {
		err = errors.New("pause not playing")
		errCode = 455
		return
	} else {
		//停止播放
		this.stopPlayThread()
	}
	strOut := RTSP_VER + " " + strconv.Itoa(200) + " " + getRTSPStatusByCode(200) + RTSP_EL
	strOut += HDR_CSEQ + ": " + strconv.Itoa(cseq) + RTSP_EL
	strOut += "Session: " + this.session + RTSP_EL
	strOut += RTSP_EL
	err = this.send([]byte(strOut))
	return
}
