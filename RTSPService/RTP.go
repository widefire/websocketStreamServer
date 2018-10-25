package RTSPService

import (
	//	"logger"
	"mediaTypes/amf"
)

const (
	NAL_TYPE_SINGLE_NAL_MIN = 1
	NAL_TYPE_SINGLE_NAL_MAX = 23
	NAL_TYPE_STAP_A         = 24
	NAL_TYPE_STAP_B         = 25
	NAL_TYPE_MTAP16         = 26
	NAL_TYPE_MTAP24         = 27
	NAL_TYPE_FU_A           = 28
	NAL_TYPE_FU_B           = 29
	Payload_h264            = 96
	Payload_MPA             = 14 //mp3 freq 90000
	RTP_H264_freq           = 90000

	RTP_MTU = 1500
)

func createRTPHeader(payloadType, seq, timestamp, ssrc uint32) []byte {
	encoder := &amf.AMF0Encoder{}
	encoder.Init()
	version := byte(2)
	padding := byte(0)
	extension := byte(0)
	marker := byte(0)
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

func createRTPHeaderAAC(payloadType, seq, timestamp, ssrc uint32) []byte {
	encoder := &amf.AMF0Encoder{}
	encoder.Init()
	version := byte(2)
	padding := byte(0)
	extension := byte(0)
	marker := byte(1)
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
