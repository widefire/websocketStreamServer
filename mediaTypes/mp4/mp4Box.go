package mp4

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"log"
)

//MP4采用盒子套盒子的递归方式存储数据
type boxOffset struct {
	longBox bool //是否是长的盒子，长盒子用8个字节表示长度，否则4个
	pos     int  //写入长度的位置
}

type MP4Box struct {
	writer bytes.Buffer
	offset list.List
}

//压入一个盒子
func (this *MP4Box) Push(name []byte) {
	offset := &boxOffset{}
	offset.longBox = false
	offset.pos = this.writer.Len()
	this.offset.PushBack(offset)
	this.Push4Bytes(0)
	this.PushBytes(name)
}

//压入一个长盒子
func (this *MP4Box) PushLongBox(name []byte) {
	log.Fatal("do not call this func ,bytes.Buffer 32bit only")
	offset := &boxOffset{}
	offset.longBox = true
	offset.pos = this.writer.Len()
	this.offset.PushBack(offset)
	this.Push4Bytes(1)
	this.Push8Bytes(0)
	this.PushBytes(name)
}

//弹出一个盒子
func (this *MP4Box) Pop() {
	if this.offset.Len() == 0 {
		return
	}
	offset := this.offset.Back().Value.(*boxOffset)
	boxSize := this.writer.Len() - offset.pos
	if offset.longBox == true {
		data := this.writer.Bytes()
		data[offset.pos+0] = 0
		data[offset.pos+1] = 0
		data[offset.pos+2] = 0
		data[offset.pos+3] = 1
		//这里不应该使用这个，因为bytes.buffer 只有32位
		log.Fatal("can not pop longbox this way,bytes.buffer 32bit only")
	} else {
		data := this.writer.Bytes()
		data[offset.pos+0] = byte((boxSize >> 24) & 0xff)
		data[offset.pos+1] = byte((boxSize >> 16) & 0xff)
		data[offset.pos+2] = byte((boxSize >> 8) & 0xff)
		data[offset.pos+3] = byte((boxSize >> 0) & 0xff)
	}

	this.offset.Remove(this.offset.Back())
}

//清空整个盒子
func (this *MP4Box) Flush() []byte {
	if this.writer.Len() == 0 {
		return nil
	}
	data := make([]byte, this.writer.Len())
	copy(data, this.writer.Bytes())
	this.writer.Reset()
	return data
}

func (this *MP4Box) Push8Bytes(data uint64) {
	err := binary.Write(&this.writer, binary.BigEndian, data)
	if err != nil {
		log.Println(err.Error())
	}
}

func (this *MP4Box) Push4Bytes(data uint32) {
	err := binary.Write(&this.writer, binary.BigEndian, data)
	if err != nil {
		log.Println(err.Error())
	}
}

func (this *MP4Box) Push2Bytes(data uint16) {
	err := binary.Write(&this.writer, binary.BigEndian, data)
	if err != nil {
		log.Println(err.Error())
	}
}

func (this *MP4Box) PushByte(data uint8) {
	err := this.writer.WriteByte(data)
	if err != nil {
		log.Println(err.Error())
	}
}

func (this *MP4Box) PushBytes(data []byte) {
	_, err := this.writer.Write(data)
	if err != nil {
		log.Println(err.Error())
	}
}
