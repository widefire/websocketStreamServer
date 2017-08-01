package RTMPService

import (
	"fmt"
	"logger"
	"sync"
	"wssAPI"
)

type rtmpPublisher struct {
	parent      wssAPI.Obj
	bPublishing bool
	mutexStatus sync.RWMutex
	rtmp        *RTMP
}

func (this *rtmpPublisher) Init(msg *wssAPI.Msg) (err error) {
	this.rtmp = msg.Param1.(*RTMP)
	this.bPublishing = false
	return
}

func (this *rtmpPublisher) Start(msg *wssAPI.Msg) (err error) {
	this.startPublish()
	return
}

func (this *rtmpPublisher) Stop(msg *wssAPI.Msg) (err error) {
	this.stopPublish()
	return
}

func (this *rtmpPublisher) GetType() string {
	return ""
}

func (this *rtmpPublisher) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *rtmpPublisher) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *rtmpPublisher) isPublishing() bool {
	return this.bPublishing
}

func (this *rtmpPublisher) startPublish() bool {
	this.mutexStatus.Lock()
	defer this.mutexStatus.Unlock()
	if this.bPublishing == true {
		return false
	}
	err := this.rtmp.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return false
	}
	err = this.rtmp.CmdStatus("status", "NetStream.Publish.Start",
		fmt.Sprintf("publish %s", this.rtmp.Link.Path), "", 0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
		return false
	}
	this.bPublishing = true
	return true
}

func (this *rtmpPublisher) stopPublish() bool {
	this.mutexStatus.Lock()
	defer this.mutexStatus.Unlock()
	if this.bPublishing == false {
		return false
	}
	err := this.rtmp.SendCtrl(RTMP_CTRL_streamEof, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return false
	}
	err = this.rtmp.CmdStatus("status", "NetStream.Unpublish.Succes",
		fmt.Sprintf("unpublish %s", this.rtmp.Link.Path), "", 0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
		return false
	}
	this.bPublishing = false
	return true
}
func (this *rtmpPublisher) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}
