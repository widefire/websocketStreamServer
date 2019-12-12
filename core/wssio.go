package core

import (
	"encoding/binary"
	"math"
)

// WSSIO is WSS io interface
type WSSIO interface {
	Read(data []uint8) (num int, err error)
	Write(data []uint8) (num int, err error)
	// Seek to offset
	// offset the offset you seek
	// whence 0 for relative to the origin of the stream, 1 for relative to the current
	// offset, 2 for relative to the end
	Seek(offset int64, whence int) (err error)
	Open(path string) (err error)
	Close() (err error)
}

func WSSIOReadL(io WSSIO, v interface{}) error {
	switch v := v.(type) {
	case *float32:
		var intval uint32
		err := binary.Read(io, binary.LittleEndian, &intval)
		if err != nil {
			return err
		}
		*v = math.Float32frombits(uint32(intval))
		return nil
	case *float64:
		var intval uint64
		err := binary.Read(io, binary.LittleEndian, &intval)
		if err != nil {
			return err
		}
		*v = math.Float64frombits(intval)
		return nil
	}
	err := binary.Read(io, binary.LittleEndian, v)
	return err
}

func WSSIOReadB(io WSSIO, v interface{}) error {
	switch v := v.(type) {
	case *float32:
		var intval uint32
		err := binary.Read(io, binary.BigEndian, &intval)
		if err != nil {
			return err
		}
		*v = math.Float32frombits(uint32(intval))
		return nil
	case *float64:
		var intval uint64
		err := binary.Read(io, binary.BigEndian, &intval)
		if err != nil {
			return err
		}
		*v = math.Float64frombits(intval)
		return nil
	}

	err := binary.Read(io, binary.BigEndian, v)
	return err
}

func WSSIOWriteL(io WSSIO, v interface{}) error {
	switch data := v.(type) {
	case float32:
		val := math.Float32bits(float32(data))
		return binary.Write(io, binary.LittleEndian, val)
	case float64:
		val := math.Float64bits(float64(data))
		return binary.Write(io, binary.LittleEndian, val)
	}

	return binary.Write(io, binary.LittleEndian, v)
}

func WSSIOWriteB(io WSSIO, v interface{}) error {
	switch data := v.(type) {
	case float32:
		val := math.Float32bits(float32(data))
		return binary.Write(io, binary.BigEndian, val)
	case float64:
		val := math.Float64bits(float64(data))
		return binary.Write(io, binary.BigEndian, val)
	}

	return binary.Write(io, binary.BigEndian, v)
}
