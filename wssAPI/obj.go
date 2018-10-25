package wssAPI

import (
	"container/list"
	"errors"
)

//task
type Task interface {
	Receiver() string
	Type() string
}

//msg
type Msg struct {
	Type    string
	Version string
	Param1  interface{}
	Param2  interface{}
	Params  *list.List
}

//obj
type Obj interface {
	Init(msg *Msg) error
	Start(msg *Msg) error
	Stop(msg *Msg) error
	GetType() string
	HandleTask(task Task) error
	ProcessMessage(msg *Msg) error
	//	SetParent(parent Obj)
}

var svrbus Obj

func SetBus(bus Obj) {
	svrbus = bus
}

func HandleTask(task Task) error {
	if svrbus != nil {
		return svrbus.HandleTask(task)
	}
	return errors.New("bus not ready")
}

//func (this *tmp) Init(msg *wssAPI.Msg) (err error) {
//	return
//}

//func (this *tmp) Start(msg *wssAPI.Msg) (err error) {
//	return
//}

//func (this *tmp) Stop(msg *wssAPI.Msg) (err error) {
//	return
//}

//func (this *tmp) GetType() string {
//	return
//}

//func (this *tmp) HandleTask(task wssAPI.Task) (err error) {
//	return
//}

//func (this *tmp) ProcessMessage(msg *wssAPI.Msg) (err error) {
//	return
//}
