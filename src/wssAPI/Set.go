package wssAPI

import (
	"sync"
)

type Set struct {
	m map[interface{}]bool
	sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		m: make(map[interface{}]bool),
	}
}

func (this *Set) Add(item interface{}) {
	this.Lock()
	defer this.Unlock()
	this.m[item] = true
}

func (this *Set) Del(item interface{}) {
	this.Lock()
	defer this.Unlock()
	delete(this.m, item)
}

func (this *Set) Has(item interface{}) bool {
	this.RLock()
	defer this.RUnlock()
	return this.m[item]
}
