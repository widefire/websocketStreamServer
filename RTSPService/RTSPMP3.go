package RTSPService

import (
	"container/list"
	"logger"
	"mediaTypes/amf"
	"mediaTypes/flv"
)

const (
	MP3_ADU          = true
	MP3_Interleaving = false //交错包，将MP3的固定头修改下即可，目前没有实现
)

func (this *RTSPHandler) generateMP3RTPPackets(tag *flv.FlvTag, beginTime uint32, trackInfo *trackInfo) (rtpPkts *list.List) {
	if MP3_ADU {
		if MP3_Interleaving {
			logger.LOGD("interleaving")
		} else {
			return this.generateMP3ADU(tag, beginTime, trackInfo)
		}
	} else {
		//logger.LOGD("raw mp3")
		return this.generateMP3RTPRaw(tag, beginTime, trackInfo)
	}
	return
}

//rfc2250
func (this *RTSPHandler) generateMP3RTPRaw(tag *flv.FlvTag, beginTime uint32, track *trackInfo) (rtpPkts *list.List) {
	//logger.LOGD(len(tag.Data))
	rtpPkts = list.New()
	payloadSize := RTP_MTU - 4
	if "tcp" == track.transPort {
		payloadSize -= 4
	}
	//rtp header
	payloadSize -= 12
	dataCur := 0
	dataSize := len(tag.Data) - 1
	data := tag.Data[1:] //flv mp3 tag,
	clockRate := 90000   //raw mp3

	timestamp := tag.Timestamp - beginTime
	if tag.Timestamp < beginTime {
		timestamp = 0
	}
	//rtptime/timestamp=rate/1000  rtptime=rate*timestamp/1000
	tmp64 := int64(clockRate / 1000)
	timestamp = uint32((tmp64 * int64(timestamp)) & 0xffffffff)
	for dataSize > payloadSize {
		tmp := amf.AMF0Encoder{}
		tmp.Init()
		var m byte
		if track.mark {
			track.mark = false
			m = 1
		} else {
			m = 0
		}
		headerData := this.createMPEGRTPHeader(Payload_MPA, uint32(track.seq), timestamp, track.ssrc, m)
		tmp.AppendByteArray(headerData)
		track.seq++
		if track.seq >= 0xffff {
			track.seq = 0
		}
		tmp.EncodeInt32(int32(dataCur))
		tmp.AppendByteArray(data[dataCur : dataCur+payloadSize])
		pktData, _ := tmp.GetData()
		rtpPkts.PushBack(pktData)
		dataSize -= payloadSize
		dataCur += payloadSize
	}

	if dataSize > 0 {
		tmp := amf.AMF0Encoder{}
		tmp.Init()
		var m byte
		if track.mark {
			track.mark = false
			m = 1
		} else {
			m = 0
		}
		headerData := this.createMPEGRTPHeader(Payload_MPA, uint32(track.seq), timestamp, track.ssrc, m)
		tmp.AppendByteArray(headerData)
		track.seq++
		if track.seq >= 0xffff {
			track.seq = 0
		}
		tmp.EncodeInt32(int32(dataCur))
		tmp.AppendByteArray(data[dataCur:])
		pktData, _ := tmp.GetData()
		rtpPkts.PushBack(pktData)
	}

	return
}

//rfc 5219
func (this *RTSPHandler) generateMP3ADU(tag *flv.FlvTag, beginTime uint32, track *trackInfo) (rtpPkts *list.List) {
	rtpPkts = list.New()
	payloadSize := RTP_MTU - 4
	if "tcp" == track.transPort {
		payloadSize -= 4
	}
	//rtp header
	payloadSize -= 12
	dataSize := len(tag.Data) - 1

	//ADU Descriptors
	if dataSize > 0x3f {
		payloadSize -= 2
	} else {
		payloadSize -= 1
	}

	clockRate := 90000 //adu

	timestamp := tag.Timestamp - beginTime
	if tag.Timestamp < beginTime {
		timestamp = 0
	}
	//rtptime/timestamp=rate/1000  rtptime=rate*timestamp/1000
	tmp64 := int64(clockRate / 1000)
	timestamp = uint32((tmp64 * int64(timestamp)) & 0xffffffff)

	var m byte
	m = 0
	if dataSize <= payloadSize {
		//只有一帧
		tmp := amf.AMF0Encoder{}
		tmp.Init()

		//rtp header
		headerData := this.createMPEGRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc, m)
		tmp.AppendByteArray(headerData)
		track.seq++
		if track.seq > 0xffff {
			track.seq = 0
		}
		//desc
		desc := createAduDesc(dataSize, false)
		tmp.AppendByteArray(desc)
		//data
		tmp.AppendByteArray(tag.Data[1:])
		pktData, _ := tmp.GetData()
		rtpPkts.PushBack(pktData)
	} else {
		cont := false
		data := tag.Data[1:]
		dataCur := 0
		for dataSize > payloadSize {
			tmp := amf.AMF0Encoder{}
			tmp.Init()
			headerData := this.createMPEGRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc, m)
			tmp.AppendByteArray(headerData)
			track.seq++
			if track.seq > 0xffff {
				track.seq = 0
			}

			//desc
			desc := createAduDesc(dataSize, cont)
			if false == cont {
				cont = true
			}
			tmp.AppendByteArray(desc)
			tmp.AppendByteArray(data[dataCur : dataCur+payloadSize])
			pktData, _ := tmp.GetData()
			rtpPkts.PushBack(pktData)
			dataCur += payloadSize
			dataSize -= payloadSize
		}
		//last frame
		if dataSize > 0 {
			tmp := amf.AMF0Encoder{}
			tmp.Init()
			//rtp header
			headerData := this.createMPEGRTPHeader(Payload_h264, uint32(track.seq), timestamp, track.ssrc, m)
			tmp.AppendByteArray(headerData)
			track.seq++
			if track.seq > 0xffff {
				track.seq = 0
			}
			//desc
			desc := createAduDesc(dataSize, cont)
			tmp.AppendByteArray(desc)
			//data
			tmp.AppendByteArray(data[dataCur:])
			pktData, _ := tmp.GetData()
			rtpPkts.PushBack(pktData)
		}
	}

	return

}

func createAduDesc(size int, cont bool) (desc []byte) {
	if size > 0x3f {
		desc = make([]byte, 2)
		if cont {
			desc[0] = (1 << 7) | (1 << 6) | byte(size>>8)
			desc[1] = byte(size & 0xff)
		} else {
			desc[0] = (0 << 7) | (1 << 6) | byte(size>>8)
			desc[1] = byte(size & 0xff)
		}
	} else {
		desc = make([]byte, 1)
		if cont {
			desc[0] = (1 << 7) | (0 << 6) | byte(size)
		} else {
			desc[0] = (0 << 7) | (0 << 6) | byte(size)
		}
	}
	return
}

func (this *RTSPHandler) createMPEGRTPHeader(payloadType, seq, timestamp, ssrc uint32, m byte) []byte {
	encoder := &amf.AMF0Encoder{}
	encoder.Init()
	version := byte(2)
	padding := byte(0)
	extension := byte(0)
	marker := m
	cc := byte(0)

	tmp := (version << 6) | (padding << 5) | (extension << 4) | (cc)
	encoder.AppendByte(tmp)
	tmp = (marker << 7) | byte(payloadType)
	encoder.AppendByte(tmp)
	encoder.EncodeInt16(int16(seq))
	encoder.EncodeInt32(int32(timestamp))
	//	logger.LOGD(timestamp)
	encoder.EncodeInt32(int32(ssrc))

	data, _ := encoder.GetData()

	return data
}
