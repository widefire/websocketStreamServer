package streamer

import (
	"container/list"
	"errors"
	"events/eLiveListCtrl"
	"events/eRTMPEvent"
	"fmt"
	"logger"
	"math/rand"
	"reflect"
	"strings"
	"time"
	"wssAPI"
)

func enableBlackList(enable bool) (err error) {

	service.mutexBlackList.Lock()
	defer service.mutexBlackList.Unlock()
	service.blackOn = enable
	return
}

func addBlackList(blackList *list.List) (err error) {

	service.mutexBlackList.Lock()
	defer service.mutexBlackList.Unlock()
	errs := ""
	for e := blackList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if false == ok {
			logger.LOGE("add blackList itor not string")
			errs += " add blackList itor not string \n"
			continue
		}
		service.blacks[name] = name
		if service.blackOn {
			service.delSource(name, 0xffffffff)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func delBlackList(blackList *list.List) (err error) {
	service.mutexBlackList.Lock()
	defer service.mutexBlackList.Unlock()
	errs := ""
	for e := blackList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("del blackList itor not string")
			errs += " del blackList itor not string \n"
			continue
		}
		delete(service.blacks, name)
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func enableWhiteList(enable bool) (err error) {

	service.mutexWhiteList.Lock()
	defer service.mutexWhiteList.Unlock()
	service.whiteOn = enable
	return
}

func addWhiteList(whiteList *list.List) (err error) {

	service.mutexWhiteList.Lock()
	defer service.mutexWhiteList.Unlock()
	errs := ""
	for e := whiteList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("add whiteList itor not string")
			errs += " add blackList itor not string \n"
			continue
		}
		service.whites[name] = name
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func delWhiteList(whiteList *list.List) (err error) {

	service.mutexWhiteList.Lock()
	defer service.mutexWhiteList.Unlock()
	errs := ""
	for e := whiteList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("del whiteList itor not string")
			errs += " del blackList itor not string \n"
			continue
		}
		delete(service.whites, name)
		if service.whiteOn {
			service.delSource(name, 0xffffffff)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func getLiveCount() (count int, err error) {
	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	count = len(service.sources)
	return
}

func getLiveList() (liveList *list.List, err error) {
	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	liveList = list.New()
	for k, v := range service.sources {
		info := &eLiveListCtrl.LiveInfo{}
		info.StreamName = k
		v.mutexSink.RLock()
		info.PlayerCount = len(v.sinks)
		info.Ip = v.addr.String()
		v.mutexSink.RUnlock()
		liveList.PushBack(info)
	}
	return
}

func getPlayerCount(name string) (count int, err error) {

	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	src, exist := service.sources[name]
	if exist == false {
		count = 0
	} else {
		count = len(src.sinks)
	}

	return
}

func (this *StreamerService) checkStreamAddAble(appStreamname string) bool {
	tmp := strings.Split(appStreamname, "/")
	var name string
	if len(tmp) > 1 {
		name = tmp[1]
	} else {
		name = appStreamname
	}
	this.mutexBlackList.RLock()
	defer this.mutexBlackList.RUnlock()
	if this.blackOn {
		for k, _ := range this.blacks {
			if name == k {
				return false
			}
		}
	}
	this.mutexWhiteList.RLock()
	defer this.mutexWhiteList.RUnlock()
	if this.whiteOn {
		for k, _ := range this.whites {
			if name == k {
				return true
			}
		}
		return false
	}
	return true
}

func (this *StreamerService) addUpstream(app *eLiveListCtrl.EveSetUpStreamApp) (err error) {
	this.mutexUpStream.Lock()
	defer this.mutexUpStream.Unlock()
	exist := false
	if app.Weight < 1 {
		app.Weight = 1
	}
	logger.LOGD(app.Id)
	for e := this.upApps.Front(); e != nil; e = e.Next() {
		v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
		if v.Equal(app) {
			exist = true
			break
		}
	}

	if exist {
		return errors.New("add up app:" + app.Id + " existed")
	} else {
		this.upApps.PushBack(app.Copy())
	}

	return
}

func (this *StreamerService) delUpstream(app *eLiveListCtrl.EveSetUpStreamApp) (err error) {
	this.mutexUpStream.Lock()
	defer this.mutexUpStream.Unlock()
	for e := this.upApps.Front(); e != nil; e = e.Next() {
		v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
		if v.Equal(app) {
			this.upApps.Remove(e)
			return
		}
	}
	return errors.New("del up app: " + app.Id + " not existed")
}

func (this *StreamerService) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}

func (this *StreamerService) badIni() {
	logger.LOGW("some bad init here!!!")
	//taskAddUp := eLiveListCtrl.NewSetUpStreamApp(true, "live", "rtmp", "live.hkstv.hk.lxdns.com", 1935)
	//	taskAddUp := eLiveListCtrl.NewSetUpStreamApp(true, "live", "rtmp", "127.0.0.1", 1935)
	//	this.HandleTask(taskAddUp)
}

func (this *StreamerService) InitUpstream(up eLiveListCtrl.EveSetUpStreamApp) {

	up.Add = true
	this.HandleTask(&up)
}

func (this *StreamerService) getUpAddrAuto() (addr *eLiveListCtrl.EveSetUpStreamApp) {
	this.mutexUpStream.RLock()
	defer this.mutexUpStream.RUnlock()
	size := this.upApps.Len()
	if size > 0 {
		totalWeight := 0
		for e := this.upApps.Front(); e != nil; e = e.Next() {
			v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
			totalWeight += v.Weight
		}
		if totalWeight == 0 {
			logger.LOGF(totalWeight)
			return
		}
		idx := rand.Intn(totalWeight) + 1
		cur := 0
		for e := this.upApps.Front(); e != nil; e = e.Next() {
			v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
			cur += v.Weight
			if cur >= idx {
				return v
			}
		}
	}
	return
}

func (this *StreamerService) getUpAddrCopy() (addrs *list.List) {
	this.mutexUpStream.RLock()
	defer this.mutexUpStream.RUnlock()
	addrs = list.New()
	for e := this.upApps.Front(); e != nil; e = e.Next() {
		addrs.PushBack(e.Value.(*eLiveListCtrl.EveSetUpStreamApp))
	}
	return
}

func (this *StreamerService) pullStreamExec(app, streamName string, addr *eLiveListCtrl.EveSetUpStreamApp) (src wssAPI.Obj, ok bool) {
	chRet := make(chan wssAPI.Obj) //这个ch由任务执行者来关闭
	protocol := strings.ToLower(addr.Protocol)
	switch protocol {
	case "rtmp":
		task := &eRTMPEvent.EvePullRTMPStream{}
		task.App = addr.App
		if strings.Contains(app,"/"){
			tmp:=strings.Split(app,"/")
			task.Instance=strings.TrimPrefix(app,tmp[0])
			task.Instance=strings.TrimPrefix(task.Instance,"/")
			task.App+="/"+task.Instance
		}else{
			task.Instance=addr.Instance
		}
		task.Address = addr.Addr
		task.Port = addr.Port
		task.Protocol = addr.Protocol
		task.StreamName = streamName
		task.Src = chRet
		task.SourceName = app + "/" + streamName
		err := wssAPI.HandleTask(task)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
	default:
		close(chRet)
		logger.LOGE(fmt.Sprintf("%s not support now...", addr.Protocol))
		return
	}
	//wait for success or timeout
	select {
	case src, ok = <-chRet:
		if ok {
			logger.LOGD("pull up stream true")
		} else {
			logger.LOGD("pull up stream false")
		}
		return
	case <-time.After(time.Duration(serviceConfig.UpstreamTimeoutSec) * time.Second):
		logger.LOGD("pull up stream timeout")
		return
	}
	return
}

func (this *StreamerService) pullStream(app, streamName, sinkId string, sinker wssAPI.Obj) {
	//按权重随机一个
	addr := this.getUpAddrAuto()
	if nil == addr {
		logger.LOGE("upstream not found")
		return
	}
	src, ok := this.pullStreamExec(app, streamName, addr)
	defer func() {
		if true == ok && wssAPI.InterfaceValid(src) {
			source, ok := src.(*streamSource)
			if true == ok {
				logger.LOGD("add sink")
				msg := &wssAPI.Msg{}
				msg.Type = wssAPI.MSG_GetSource_NOTIFY
				sinker.ProcessMessage(msg)
				source.AddSink(sinkId, sinker)
			} else {
				logger.LOGE("add sink failed", source, ok)
				msg := &wssAPI.Msg{Type: wssAPI.MSG_GetSource_Failed}
				sinker.ProcessMessage(msg)
			}
		} else {
			logger.LOGE("bad add", ok, src)
			logger.LOGD(reflect.TypeOf(src))
			msg := &wssAPI.Msg{Type: wssAPI.MSG_GetSource_Failed}
			sinker.ProcessMessage(msg)
		}

	}()
	if true == ok && wssAPI.InterfaceValid(src) {
		return
	}
	//按顺序进行
	addrs := this.getUpAddrCopy()
	for e := addrs.Front(); e != nil; e = e.Next() {
		var addr *eLiveListCtrl.EveSetUpStreamApp
		addr, ok = e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
		if false == ok || nil == addr {
			logger.LOGE("invalid addr")
			continue
		}
		src, ok = this.pullStreamExec(app, streamName, addr)
		if true == ok && wssAPI.InterfaceValid(src) {
			return
		}
	}
}
