package RTMP

import (
	"errors"

	"github.com/widefire/websocketStreamServer/core"
)

type RTMPFormat struct {
	io         core.WSSIO
	isInput    bool
	remoteTime uint32
	path       string
}

func NewRTMPFormat(io core.WSSIO, path string, isInput bool) (core.WSSFormat, error) {
	f := &RTMPFormat{}
	f.io = io
	f.isInput = isInput
	f.path = path
	return f, nil
}

func (r *RTMPFormat) Open() error {
	if r.io == nil {
		return errors.New("need io")
	}
	if r.isInput {
		return r.HandShakeAsServer()
	} else {
		return r.HandShakeAsClient()
	}
}

func (r *RTMPFormat) ReadMetadata() error {
	return nil
}

func (r *RTMPFormat) WriteMetadata() error {
	return nil
}

func (r *RTMPFormat) ReadPacket(packt *core.WSSPacket) error {
	return nil
}

func (r *RTMPFormat) WritePacket(packet core.WSSPacket) error {
	return nil
}

func (r *RTMPFormat) Seek(time int64) error {
	return nil
}

func (r *RTMPFormat) Close() error {
	return nil
}
