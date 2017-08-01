package ts

import (
	"logger"
	"mediaTypes/aac"
	"mediaTypes/amf"
	"mediaTypes/flv"
)

func (this *TsCreater) audioPayload(tag *flv.FlvTag) (payload []byte, size int) {
	if this.audioTypeId == 0xf {
		//adth:=aac.GenerateADTHeader(this.asc,len(tag.Data)-2)
		adth := aac.CreateAACADTHeader(this.asc, len(tag.Data)-2)
		size = len(adth) + len(tag.Data) - 2
		payload = make([]byte, size)
		copy(payload, adth)
		copy(payload[len(adth):], tag.Data[2:])
		return
	} else if this.audioTypeId == 0x03 || this.audioTypeId == 0x04 {
		size = len(tag.Data) - 1
		payload = make([]byte, size)
		copy(payload, tag.Data[1:])
		return
	} else {
		logger.LOGF(this.audioTypeId)
		logger.LOGE("invalid audio type")
	}
	return
}

func (this *TsCreater) calPcrPtsDts(tag *flv.FlvTag) (pcr, pcrExt, pts, dts uint64) {
	timeMS := uint64(tag.Timestamp )
	pcr = (timeMS * 90) & 0x1ffffffff
	pcrExt = (timeMS * PCR_HZ / 1000) & 0x1ff
	if len(tag.Data) < 5 {
		logger.LOGF("wtf")
	}
	compositionTime, _ := amf.AMF0DecodeInt24(tag.Data[2:])
	u64 := uint64(compositionTime)
	dts = timeMS*90 + 90
	pts = dts + u64*90
	return
}

func (this *TsCreater)calAudioTime(tag *flv.FlvTag)  {
	//tmp:=int64(90*this.audioSampleHz*int(tag.Timestamp-this.beginTime))
	tmp:=int64(90*int(tag.Timestamp))
	//logger.LOGT(tmp,tag.Timestamp-this.beginTime)
	//audioPtsDelta := int64(90000 * int64(this.audioFrameSize) / int64(this.audioSampleHz))
	//this.audioPts += audioPtsDelta
	//logger.LOGD(this.audioPts)
	this.audioPts=tmp
}