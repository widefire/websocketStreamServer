package eRTMPEvent

import (
	"wssAPI"
)

const (
	PullRTMPStream = "PullRTMPStream"
)

type EvePullRTMPStream struct {
	SourceName string //用来创建和删除源，源名称和app+streamName 并不一样
	Protocol   string //RTMP,RTMPS,RTMPS and so on
	App        string
	Address    string
	Port       int
	StreamName string
	Src        chan wssAPI.Obj
}

func (this *EvePullRTMPStream) Receiver() string {
	return wssAPI.OBJ_RTMPServer
}

func (this *EvePullRTMPStream) Type() string {
	return PullRTMPStream
}

func (this *EvePullRTMPStream) Init(protocol, app, addr, streamName, sourceName string, port int) {
	this.Protocol = protocol
	this.App = app
	this.Address = addr
	this.Port = port
	this.StreamName = streamName
	this.SourceName = sourceName
	this.Src = make(chan wssAPI.Obj)
}

func (this *EvePullRTMPStream) Copy() (out *EvePullRTMPStream) {
	out = &EvePullRTMPStream{}
	out.Protocol = this.Protocol
	out.App = this.App
	out.Address = this.Address
	out.Port = this.Port
	out.StreamName = this.StreamName
	out.SourceName = this.SourceName
	out.Src = this.Src
	return
}
