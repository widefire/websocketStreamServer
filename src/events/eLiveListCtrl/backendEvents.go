package eLiveListCtrl

import (
	"container/list"
	"wssAPI"
)

const (
	EnableBlackList    = "EnableBlackList"
	SetBlackList       = "SetBlackList"
	EnableWhiteList    = "EnableWhiteList"
	SetWhiteList       = "SetWhiteList"
	GetLiveList        = "GetLiveList"
	GetLivePlayerCount = "GetLivePlayerCount"
	SetUpStreamApp     = "SetUpStreamApp"
)

//black list
type EveEnableBlackList struct {
	Enable bool
}

func (this *EveEnableBlackList) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveEnableBlackList) Type() string {
	return EnableBlackList
}

type EveSetBlackList struct {
	Add   bool
	Names *list.List
}

func (this *EveSetBlackList) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveSetBlackList) Type() string {
	return SetBlackList
}

//white list
type EveEnableWhiteList struct {
	Enable bool
}

func (this *EveEnableWhiteList) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveEnableWhiteList) Type() string {
	return EnableWhiteList
}

type EveSetWhiteList struct {
	Add   bool
	Names *list.List
}

func (this *EveSetWhiteList) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveSetWhiteList) Type() string {
	return SetWhiteList
}
