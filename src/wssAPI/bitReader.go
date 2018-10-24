package wssAPI

import (
	"bytes"
	"encoding/binary"
)

//read data by bit
type BitReader struct {
	buf    []byte
	curBit int
}

func (this *BitReader) Init(data []byte) {
	this.curBit = 0
	this.buf = make([]byte, len(data))
	copy(this.buf, data)
}

func (this *BitReader) ReadBit() int {
	if this.curBit > (len(this.buf) << 3) {
		return -1
	}
	idx := (this.curBit >> 3)
	offset := this.curBit%8 + 1
	this.curBit++
	return int(this.buf[idx]>>uint(8-offset)) & 0x01
}

func (this *BitReader) ReadBits(num int) int {
	r := 0
	for i := 0; i < num; i++ {
		r |= (this.ReadBit() << uint(num-i-1))
	}
	return r
}

func (this *BitReader) Read32Bits() uint32 {
	idx := (this.curBit >> 3)
	var r uint32
	binary.Read(bytes.NewReader(this.buf[idx:]), binary.BigEndian, &r)
	this.curBit += 32

	return r
}

func (this *BitReader) ReadExponentialGolombCode() int {
	r := 0
	i := 0
	for this.ReadBit() == 0 && (i < 32) {
		i++
	}
	r = this.ReadBits(i)
	r += (1 << uint(i)) - 1
	return r
}

func (this *BitReader) ReadSE() int {
	r := this.ReadExponentialGolombCode()
	if (r & 0x01) != 0 {
		r = (r + 1) / 2
	} else {
		r = -(r / 2)
	}
	return r
}

func (this *BitReader) CopyBits(num int) int {
	cur := this.curBit
	r := 0
	for i := 0; i < num; i++ {
		r |= (this.copyBit(cur+i) << uint(num-i-1))
	}
	return r
}

func (this *BitReader) copyBit(cur int) int {
	if cur > (len(this.buf) << 3) {
		return -1
	}
	idx := (cur >> 3)
	offset := cur%8 + 1
	return int(this.buf[idx]>>uint(8-offset)) & 0x01
}

func (this *BitReader) BitsLeft() int {
	return (len(this.buf) << 3) - this.curBit
}
