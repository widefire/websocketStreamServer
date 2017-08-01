package RTMPService

import (
	"container/list"
	"errors"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"sync"
	"time"
	"wssAPI"
)

const (
	play_idle = iota
	play_playing
	play_paused
)

type rtmpPlayer struct {
	parent         wssAPI.Obj
	playStatus     int
	mutexStatus    sync.RWMutex
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
	rtmp           *RTMP
}

func (this *rtmpPlayer) Init(msg *wssAPI.Msg) (err error) {
	this.playStatus = play_idle
	this.waitPlaying = new(sync.WaitGroup)
	this.rtmp = msg.Param1.(*RTMP)
	this.resetCache()
	return
}

func (this *rtmpPlayer) Start(msg *wssAPI.Msg) (err error) {
	this.startPlay()
	return
}

func (this *rtmpPlayer) Stop(msg *wssAPI.Msg) (err error) {
	this.stopPlay()
	return
}

func (this *rtmpPlayer) GetType() string {
	return ""
}

func (this *rtmpPlayer) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *rtmpPlayer) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *rtmpPlayer) startPlay() {
	this.mutexStatus.Lock()
	defer this.mutexStatus.Unlock()
	switch this.playStatus {
	case play_idle:
		go this.threadPlay()
		this.playStatus = play_playing
	case play_paused:
		logger.LOGE("pause not processed")
		return
	case play_playing:
		return
	}

}

func (this *rtmpPlayer) stopPlay() {
	this.mutexStatus.Lock()
	defer this.mutexStatus.Unlock()
	switch this.playStatus {
	case play_idle:
		return
	case play_paused:
		logger.LOGE("pause not processed")
		return
	case play_playing:
		//stop play thread
		//reset
		this.playing = false
		this.waitPlaying.Wait()
		this.playStatus = play_idle
		this.resetCache()
	}

}

func (this *rtmpPlayer) pause() {

}

func (this *rtmpPlayer) IsPlaying() bool {
	return this.playStatus == play_playing
}

func (this *rtmpPlayer) appendFlvTag(tag *flv.FlvTag) (err error) {
	this.mutexStatus.RLock()
	defer this.mutexStatus.RUnlock()
	if this.playStatus != play_playing {
		err = errors.New("not playing ,can not recv mediaData")
		logger.LOGE(err.Error())
		return
	}
	tag = tag.Copy()
	//	if this.beginTime == 0 && tag.Timestamp > 0 {
	//		this.beginTime = tag.Timestamp
	//	}
	//	tag.Timestamp -= this.beginTime
	if false == this.keyFrameWrited && tag.TagType == flv.FLV_TAG_Video {
		if this.videoHeader == nil {
			this.videoHeader = tag
		} else {
			if (tag.Data[0] >> 4) == 1 {
				this.keyFrameWrited = true
			} else {
				return
			}
		}

	}
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
	this.cache.PushBack(tag)
	return
}

func (this *rtmpPlayer) setPlayParams(path string, startTime, duration int, reset bool) bool {
	this.mutexStatus.RLock()
	defer this.mutexStatus.RUnlock()
	if this.playStatus != play_idle {
		return false
	}
	return true
}

func (this *rtmpPlayer) resetCache() {
	this.audioHeader = nil
	this.videoHeader = nil
	this.metadata = nil
	this.beginTime = 0
	this.keyFrameWrited = false
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
	this.cache = list.New()
}

func (this *rtmpPlayer) threadPlay() {
	this.playing = true
	this.waitPlaying.Add(1)
	defer func() {
		this.waitPlaying.Done()
		this.sendPlayEnds()
		this.playStatus = play_idle
	}()
	this.sendPlayStarts()

	for this.playing == true {
		this.mutexCache.Lock()
		if this.cache == nil || this.cache.Len() == 0 {
			this.mutexCache.Unlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if this.cache.Len() > serviceConfig.CacheCount {
			this.mutexCache.Unlock()
			//bw not enough
			this.rtmp.CmdStatus("warning", "NetStream.Play.InsufficientBW",
				"instufficient bw", this.rtmp.Link.Path, 0, RTMP_channel_Invoke)
			//shutdown
			return
		}
		tag := this.cache.Front().Value.(*flv.FlvTag)
		this.cache.Remove(this.cache.Front())
		this.mutexCache.Unlock()
		//时间错误

		err := this.rtmp.SendPacket(FlvTagToRTMPPacket(tag), false)
		if err != nil {
			logger.LOGE("send rtmp packet failed in play")
			return
		}
	}
}

func (this *rtmpPlayer) sendPlayEnds() {

	err := this.rtmp.SendCtrl(RTMP_CTRL_streamEof, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	err = this.rtmp.FCUnpublish()
	if err != nil {
		logger.LOGE("FCUnpublish failed:" + err.Error())
		return
	}

	err = this.rtmp.CmdStatus("status", "NetStream.Play.UnpublishNotify",
		fmt.Sprintf("%s is unpublished", this.rtmp.Link.Path),
		this.rtmp.Link.Path, 0, RTMP_channel_Invoke)

	if err != nil {
		logger.LOGE(err.Error())
		return
	}
}

func (this *rtmpPlayer) sendPlayStarts() {
	err := this.rtmp.CmdStatus("status", "NetStream.Play.PublishNotify",
		fmt.Sprintf("%s is now unpublished", this.rtmp.Link.Path),
		this.rtmp.Link.Path,
		0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
	}
	err = this.rtmp.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	if true == this.reset {
		err = this.rtmp.CmdStatus("status", "NetStream.Play.Reset",
			fmt.Sprintf("Playing and resetting %s", this.rtmp.Link.Path),
			this.rtmp.Link.Path, 0, RTMP_channel_Invoke)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
	}

	err = this.rtmp.CmdStatus("status", "NetStream.Play.Start",
		fmt.Sprintf("Started playing %s", this.rtmp.Link.Path), this.rtmp.Link.Path, 0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	logger.LOGT("start playing")
}

func (this *rtmpPlayer) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}
