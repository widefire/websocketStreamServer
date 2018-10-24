package eLiveListCtrl

import (
	"wssAPI"
)

type EveSetUpStreamApp struct {
	Id       string `json:"Id"`
	Add      bool
	App      string `json:"app"`
	Instance string `json:"instance,omitempty"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Addr     string `json:"addr"`
	Weight   int    `json:"weight"`
}

func (this *EveSetUpStreamApp) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveSetUpStreamApp) Type() string {
	return SetUpStreamApp
}

func NewSetUpStreamApp(add bool, app,instance, protocol, addr, name string, port, weight int) (out *EveSetUpStreamApp) {
	out = &EveSetUpStreamApp{}
	out.Add = add
	out.App = app
	out.Instance=instance
	out.Protocol = protocol
	out.Addr = addr
	out.Port = port
	out.Id = name
	out.Weight = weight
	return
}

func (this *EveSetUpStreamApp) Copy() (out *EveSetUpStreamApp) {
	out = &EveSetUpStreamApp{}
	out.Id = this.Id
	out.Add = this.Add
	out.App = this.App
	out.Instance=this.Instance
	out.Protocol = this.Protocol
	out.Addr = this.Addr
	out.Port = this.Port
	out.Weight = this.Weight
	return
}

func (this *EveSetUpStreamApp) Equal(rh *EveSetUpStreamApp) bool {
	return this.Id == rh.Id &&
		this.App == rh.App &&
		this.Protocol == rh.Protocol &&
		this.Addr == rh.Addr &&
		this.Port == rh.Port &&
		this.Weight == rh.Weight
}
