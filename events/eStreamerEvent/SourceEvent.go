package eStreamerEvent

import (
	"net"
	"wssAPI"
)

const (
	AddSource = "AddSource"
	DelSource = "DelSource"
	GetSource = "GetSource"
)

type EveAddSource struct {
	StreamName string
	RemoteIp	net.Addr
	Producer   wssAPI.Obj
	Id         int64      //outPut
	SrcObj     wssAPI.Obj //out
}

func (this *EveAddSource) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveAddSource) Type() string {
	return AddSource
}

type EveDelSource struct {
	StreamName string //in
	Id         int64  //in
}

func (this *EveDelSource) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveDelSource) Type() string {
	return DelSource
}

type EveGetSource struct {
	StreamName  string
	SrcObj      wssAPI.Obj
	HasProducer bool
}

func (this *EveGetSource) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveGetSource) Type() string {
	return GetSource
}
