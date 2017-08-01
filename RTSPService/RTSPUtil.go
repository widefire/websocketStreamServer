package RTSPService

import (
	"bytes"
	"container/list"
	"encoding/base64"
	"fmt"
	"logger"
	"mediaTypes/aac"
	"mediaTypes/amf"
	"mediaTypes/flv"
	"mediaTypes/h264"
	"mediaTypes/mp3"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"wssAPI"
)

const (
	strRegOneSpace = "[' ']+" //多个空格
)

var regOneSpace *regexp.Regexp
var rtpPortSet *wssAPI.Set

func init() {
	regOneSpace, _ = regexp.Compile(strRegOneSpace)
	rtpPortSet = wssAPI.NewSet()
}

//vlc no heart beat
func ReadPacket(conn net.Conn, timeout bool) (data []byte, err error) {
	if timeout {
		logger.LOGT(serviceConfig.TimeoutSec)
		err = conn.SetReadDeadline(time.Now().Add(time.Duration(serviceConfig.TimeoutSec) * time.Second))
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		defer conn.SetReadDeadline(time.Time{})
	}
	firstByte := make([]byte, 1)
	_, err = conn.Read(firstByte)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	if '$' == firstByte[0] {
		threeBytes, err := wssAPI.TcpRead(conn, 3)
		_, err = conn.Read(threeBytes)
		if err != nil {
			logger.LOGE(err.Error())
			return nil, err
		}
		dataLength, err := amf.AMF0DecodeInt16(threeBytes[1:])
		if err != nil {
			logger.LOGE(err.Error())
			return nil, err
		}
		dataLast, err := wssAPI.TcpRead(conn, int(dataLength))
		if err != nil {
			logger.LOGE(err.Error())
			return nil, err
		}
		buf := bytes.NewBuffer([]byte{})
		buf.WriteByte(firstByte[0])
		buf.Write(threeBytes)
		buf.Write(dataLast)
		return buf.Bytes(), nil
	} else {

		buf := bytes.NewBuffer([]byte{})
		buf.WriteByte(firstByte[0])
		tmp := make([]byte, 1)
		for {
			_, err := conn.Read(tmp)
			if err != nil {
				logger.LOGE(err.Error())
				return nil, err
			}
			buf.WriteByte(tmp[0])
			if isFullRTSPPacket(buf.Bytes()) {
				return buf.Bytes(), nil
			}
		}
	}
	return
}

func isFullRTSPPacket(data []byte) bool {
	if nil == data || len(data) < 4 {
		return false
	}
	el := make([]byte, 4)
	el[0] = '\r'
	el[1] = '\n'
	el[2] = '\r'
	el[3] = '\n'
	if bytes.HasSuffix(data, el) {
		strData := string(data)
		strData = strings.ToLower(strData)
		if strings.Contains(strData, "content-length:") {
			splits := strings.Split(strData, "content-length:")
			if len(splits) != 2 {
				logger.LOGW("unknown fmt rtsp data:" + strData)
				return true
			} else {
				contentLength, err := strconv.Atoi(splits[1])
				if err != nil {
					logger.LOGE(err.Error())
					return false
				}
				if contentLength == 0 {
					return true
				} else {
					return strings.Count(strData, "\r\n\r\n") > 1
				}
			}
		} else {
			return true
		}
	}

	return false
}

func getHeaderByName(heads []string, name string, caseinsensitive bool) (val string) {
	if false == caseinsensitive {
		name = strings.ToLower(name)
	}

	for _, v := range heads {
		if false == caseinsensitive {
			v = strings.ToLower(v)
		}
		if strings.HasPrefix(v, name) {
			return v
		}
	}
	return
}

func getCSeq(heads []string) (cseq int) {
	//转换为小写
	cseqLine := getHeaderByName(heads, "CSeq:", false)
	if len(cseqLine) < 1 {
		logger.LOGE("cseq not found")
		return -1
	}
	fmt.Sscanf(cseqLine, "cseq:%d", &cseq)
	return
}

func removeSpace(data string) string {
	return regOneSpace.ReplaceAllString(data, " ")
}

//When the value of packetization-mode is equal
//to 0 or packetization-mode is not present, the
//single NAL mode, as defined in section 6.2 of
//RFC 3984, MUST be used. This mode is in use in
//standards using ITU-T Recommendation H.241 [15]
//(see section 12.1). When the value of
//packetization-mode is equal to 1, the noninterleaved mode, as defined in section 6.3 of
//RFC 3984, MUST be used. When the value of
//packetization-mode is equal to 2, the
//interleaved mode, as defined in section 6.4 of
//RFC 3984, MUST be used. The value of
//packetization mode MUST be an integer in the
//range of 0 to 2, inclusive

func generateSDP(videoHeader, audioHeader *flv.FlvTag) (sdp string, ok bool) {
	if videoHeader != nil {
		//读取视频类型，目前仅支持h264
		sdp += genH264sdp(videoHeader.Data)
		if len(sdp) > 0 {
			ok = true
		}
	}
	if audioHeader != nil {
		//读取音频类型,目前什么都不支持
		audioType := int(audioHeader.Data[0] >> 4)
		logger.LOGD(audioHeader.Data[0])
		switch audioType {
		case flv.SoundFormat_AAC:
			sdp += genAACsdp(audioHeader.Data)
			if len(sdp) > 0 {
				ok = true
			}
		case flv.SoundFormat_MP3:
			sdp += genMP3sdp(audioHeader.Data[1:])
			if len(sdp) > 0 {
				ok = true
			}
		default:
			logger.LOGW("audio not support now")
		}

	}

	return
}

func genH264sdp(data []byte) (sdp string) {
	var sps, pps []byte
	cur := 10
	numOfSequenceParameterSets := int(data[cur] & 0x1f)
	cur++
	for i := 0; i < numOfSequenceParameterSets; i++ {
		sequenceParameterSetLength := (int(data[cur]<<8) | int(data[cur+1]))
		cur += 2
		sps = make([]byte, sequenceParameterSetLength)
		copy(sps, data[cur:cur+sequenceParameterSetLength])
		cur += sequenceParameterSetLength
	}
	numOfPictureParameterSets := int(data[cur])
	cur++
	for i := 0; i < numOfPictureParameterSets; i++ {
		pictureParameterSetLength := (int(data[cur]<<8) | int(data[cur+1]))
		cur += 2
		pps = make([]byte, pictureParameterSetLength)
		copy(pps, data[cur:cur+pictureParameterSetLength])
		cur += pictureParameterSetLength
	}
	spsNoEmu := make([]byte, len(sps))
	copy(spsNoEmu, sps)
	spsNoEmu = h264.EmulationPrevention(spsNoEmu)
	var profileLevelId int

	profileLevelId = ((int(spsNoEmu[1]) << 16) | (int(spsNoEmu[2]) << 8) | (int(spsNoEmu[3])))
	spsBase64 := base64.StdEncoding.EncodeToString(sps)
	ppsBase64 := base64.StdEncoding.EncodeToString(pps)

	fmtpFmt := fmt.Sprintf("a=fmtp:%d packetization-mode=1;profile-level-id=%06X;sprop-parameter-sets=%s,%s\r\n",
		96, profileLevelId, spsBase64, ppsBase64)
	mediaType := "video"
	rtpPayloadType := RTSP_payload_h264
	ipAddress := "0.0.0.0"
	rtpmapLine := "a=rtpmap:96 H264/90000\r\n"
	rtcpmuxLine := ""
	rangeLine := "a=range:npt=0-\r\n"
	sdpLines := fmt.Sprintf("m=%s %d RTP/AVP %d\r\nc=IN IP4 %s\r\nb=AS:%d\r\n%s%s%s%sa=control:%s\r\n",
		mediaType, 0, rtpPayloadType, ipAddress, 500, rtpmapLine, rtcpmuxLine, rangeLine, fmtpFmt, ctrl_track_video)
	sdp = sdpLines

	return
}

func genAACsdp(data []byte) (sdp string) {
	asc := aac.MP4AudioGetConfig(data[2:])
	if nil == asc {
		return
	}
	//mpeg4的音频RTP 时间 单位就是采样率
	sdp += "m=audio 0 RTP/AVP 96" + RTSP_EL
	sdp += "a=rtpmap:96 MPEG4-GENERIC/" + strconv.Itoa(asc.Sample_rate) + "/" + strconv.Itoa(asc.Channels) + RTSP_EL
	sdp += "a=fmtp:96 streamtype=5;profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3;" +
		"config=" + aac.CreateAudioSpecificConfigForSDP(asc) + RTSP_EL
	sdp += "a=control:" + ctrl_track_audio + RTSP_EL
	logger.LOGD(sdp)
	return
}

func genMP3sdp(data []byte) (sdp string) {
	header, err := mp3.ParseMP3Header(data)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	logger.LOGD(header.SampleRate)
	if false == MP3_ADU {
		sdp += "m=audio 0 RTP/AVP " + strconv.Itoa(Payload_MPA) + RTSP_EL
		sdp += "b=AS:" + strconv.Itoa(header.Bitrate) + RTSP_EL
		sdp += "a=control:" + ctrl_track_audio + RTSP_EL
		logger.LOGD(sdp)
	} else {
		sdp += "m=audio 0 RTP/AVP " + strconv.Itoa(Payload_h264) + RTSP_EL
		sdp += "b=AS:" + strconv.Itoa(header.Bitrate) + RTSP_EL
		sdp += "a=rtpmap:96 MPA-ROBUST/90000" + RTSP_EL
		sdp += "a=control:" + ctrl_track_audio + RTSP_EL
		logger.LOGD(sdp)
	}
	return
}
func genRTPRTCP() (rtpPort, rtcpPort int, rtpConn, rtcpConn *net.UDPConn, ok bool) {
	//尝试10000次
	counts := 0
	for counts < 10000 {
		counts++
		addr := ":0"

		rtpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			logger.LOGE("try rtp port failed:" + err.Error())
			return
		}
		rtpConn, err = net.ListenUDP("udp", rtpAddr)
		if err != nil {
			logger.LOGE("listen rtp failed:" + err.Error())
			return
		}

		subs := strings.Split(rtpConn.LocalAddr().String(), ":")
		rtpPort, err = strconv.Atoi(subs[len(subs)-1])
		if err != nil {
			rtpConn.Close()
			logger.LOGE("get port failed:" + err.Error())
			return
		}

		if rtpPort%2 != 0 {
			logger.LOGT("rtp even retry!")
			rtpConn.Close()
			continue
		}

		rtcpPort = rtpPort + 1

		addr = ":" + strconv.Itoa(rtcpPort)
		rtcpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			logger.LOGE("resolve udp addr faild")
			rtpConn.Close()
			continue
		}
		rtcpConn, err = net.ListenUDP("udp", rtcpAddr)
		if err != nil {
			logger.LOGE("listen rtcp failed:" + err.Error())
			rtpConn.Close()
			continue
		}
		logger.LOGT(fmt.Sprintf("server rtp:%d  rtcp:%d", rtpPort, rtcpPort))
		ok = true
		break
	}
	return
}

func getH264Keyframe(tags *list.List, mutex sync.RWMutex) (getKeyFrame bool, beginTime uint32) {
	mutex.Lock()
	defer mutex.Unlock()
	if tags == nil || tags.Len() == 0 {
		return false, 0
	}
	//udp 从任意包开始
	return true, tags.Front().Value.(*flv.FlvTag).Timestamp
	for tags.Len() > 0 {
		tag := tags.Front().Value.(*flv.FlvTag)
		if (tag.Data[0] & 0xf) != 0x7 {
			logger.LOGE("invalid video type:" + string(tag.Data[0]))
		}
		if tag.Data[0] == 0x17 {
			return true, tag.Timestamp
		}
		tags.Remove(tags.Front())
	}
	return
}
