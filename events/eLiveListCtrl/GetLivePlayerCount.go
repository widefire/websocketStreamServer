package eLiveListCtrl

import (
	"wssAPI"
)

type EveGetLivePlayerCount struct {
	LiveName string
	Count    int
}

func (this *EveGetLivePlayerCount) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveGetLivePlayerCount) Type() string {
	return GetLivePlayerCount
}
