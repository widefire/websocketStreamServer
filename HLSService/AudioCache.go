package HLSService

import (
	"logger"
	"mediaTypes/aac"
	"mediaTypes/flv"
)

//may be aac or mp3 or other
//first support aac
type audioCache struct {
	audioType int
	aacCache  *aac.AACCreater
}

func (this *audioCache) Init(tag *flv.FlvTag) {
	this.audioType = int((tag.Data[0] >> 4) & 0xf)
	if this.audioType == flv.SoundFormat_AAC {
		this.aacCache = &aac.AACCreater{}
		this.aacCache.Init(tag.Data[2:])
	} else {
		logger.LOGW("sound fmt not processed")
	}
}

func (this *audioCache) AddTag(tag *flv.FlvTag) {
	if this.audioType == flv.SoundFormat_AAC {
		this.aacCache.Add(tag.Data[2:])
	} else {
		//		logger.LOGW("sound fmt not processed")
	}
}

func (this *audioCache) Flush() (data []byte) {
	if this.audioType == flv.SoundFormat_AAC {
		data = this.aacCache.Flush()
		return data
	} else {
		return nil
	}
	return nil
}
