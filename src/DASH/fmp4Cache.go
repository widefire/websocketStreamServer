package DASH

import (
	"sync"
	"container/list"
	"errors"
)

type FMP4Cache struct {
	audioHeader []byte
	videoHeader []byte
	cacheSize int
	muxAudio sync.RWMutex
	audioCache map[int64][]byte
	audioKeys *list.List
	muxVideo sync.RWMutex
	videoCache map[int64][]byte
	videoKeys *list.List
}

func NewFMP4Cache(cacheSize int)(cache *FMP4Cache){
	cache=&FMP4Cache{}
	cache.cacheSize=cacheSize
	cache.audioCache=make(map[int64][]byte)
	cache.videoCache=make(map[int64][]byte)
	cache.audioKeys=list.New()
	cache.videoKeys=list.New()
	return
}


func (this *FMP4Cache)VideoHeaderGenerated(videoHeader []byte){
	this.muxVideo.Lock()
	defer this.muxVideo.Unlock()
	this.videoHeader=make([]byte,len(videoHeader))
	copy(this.videoHeader,videoHeader)
}
func (this *FMP4Cache)VideoSegmentGenerated(videoSegment []byte,timestamp int64,duration int){
	this.muxVideo.Lock()
	defer this.muxVideo.Unlock()
	this.videoCache[timestamp]=videoSegment
	this.videoKeys.PushBack(timestamp)
	if this.videoKeys.Len()>this.cacheSize{
		key:=this.videoKeys.Front().Value.(int64)
		delete(this.videoCache,key)
		this.videoKeys.Remove(this.videoKeys.Front())
	}
}
func (this *FMP4Cache)AudioHeaderGenerated(audioHeader []byte){
	this.muxAudio.Lock()
	defer this.muxAudio.Unlock()
	this.audioHeader=make([]byte,len(audioHeader))
	copy(this.audioHeader,audioHeader)
}
func (this *FMP4Cache)AudioSegmentGenerated(audioSegment []byte,timestamp int64,duration int){
	this.muxAudio.Lock()
	defer this.muxAudio.Unlock()
	this.audioCache[timestamp]=audioSegment
	this.audioKeys.PushBack(timestamp)
	if this.audioKeys.Len()>this.cacheSize{
		key:=this.audioKeys.Front().Value.(int64)
		delete(this.audioCache,key)
		this.audioKeys.Remove(this.audioKeys.Front())
	}
}

func (this *FMP4Cache)GetAudioHeader()(data []byte,err error){
	this.muxAudio.RLock()
	defer this.muxAudio.RUnlock()
	data=this.audioHeader
	if nil==data||len(data)==0{
		err=errors.New("no audio header now")
	}
	return
}

func (this *FMP4Cache)GetAudioSegment(timestamp int64)(seg []byte,err error){
	this.muxAudio.RLock()
	defer this.muxAudio.RUnlock()
	seg=this.audioCache[timestamp]
	if nil==seg{
		err=errors.New("audio segment not found")
	}
	return
}

func (this *FMP4Cache)GetVideoHeader()(data []byte,err error){
	this.muxVideo.RLock()
	defer this.muxVideo.RUnlock()
	data=this.videoHeader
	if nil==data||len(data)==0{
		err=errors.New("no video header")
	}
	return
}

func (this *FMP4Cache)GetVideoSegment(timestamp int64)(seg []byte,err error){
	this.muxVideo.RLock()
	defer this.muxVideo.RUnlock()
	seg=this.videoCache[timestamp]
	if nil==seg{
		err=errors.New("video segment not found")
	}
	return
}