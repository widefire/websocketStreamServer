package RTSPService

import (
	"math/rand"
	"sync"
	"wssAPI"
)

type ssrcManager struct {
	set   *wssAPI.Set
	mutex sync.RWMutex
}

func newSSRCManager() (manager *ssrcManager) {
	manager = &ssrcManager{}
	manager.set = wssAPI.NewSet()
	return
}

func (this *ssrcManager) NewSSRC() (id uint32) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for {
		id = rand.Uint32()
		if false == this.set.Has(id) {
			this.set.Add(id)
			return
		}
	}
	return
}
