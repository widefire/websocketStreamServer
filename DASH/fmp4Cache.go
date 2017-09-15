package DASH

import "sync"

type FMP4Cache struct {
	audioHeader []byte
	videoHeader []byte
	muxAudio sync.RWMutex
	audioCache map[int64][]byte
	muxVideo sync.RWMutex
	videoCache map[int64][]byte
}

func NewFMP4Cache()(cache *FMP4Cache){
	cache=&FMP4Cache{}
	cache.audioCache=make(map[int64][]byte)
	cache.videoCache=make(map[int64][]byte)
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
}
func (this *FMP4Cache)AudioHeaderGenerated(audioHeader []byte){
	this.muxAudio.Lock()
	defer this.muxAudio.Unlock()
}
func (this *FMP4Cache)AudioSegmentGenerated(audioSegment []byte,timestamp int64,duration int){
	this.muxAudio.Lock()
	defer this.muxAudio.Unlock()
}