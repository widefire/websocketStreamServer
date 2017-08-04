package fragmentMP4

import (
	"bytes"
	"encoding/binary"
)

//not support large box now
type MP4Box struct {
	boxType      []byte
	version      uint8
	flag         int
	fullBox      bool
	writer       bytes.Buffer
	boxSize      uint32
	boxLargeSize uint64
}

func (box *MP4Box) Init(boxType []byte) {
	box.boxType = boxType
}

func (box *MP4Box) SetVersionFlags(version uint8, flag int) {
	box.version = version
	box.flag = flag
	box.fullBox = true
}

func (box *MP4Box) Flush() []byte {
	defer func() {
		box.writer.Reset()
		box.fullBox = false
	}()
	box.boxSize = 8
	if box.fullBox {
		box.boxSize += 4
	}
	if box.writer.Len() >= int(0xffffffff-box.boxSize) {
		box.boxLargeSize = uint64(int(box.boxSize) + 8 + box.writer.Len())
		box.boxSize = 1
	} else {
		box.boxSize += uint32(box.writer.Len())
	}
	writer := bytes.Buffer{}
	binary.Write(writer, binary.BigEndian, box.boxSize)
	writer.Write(box.boxType)

	if 1 == box.boxSize {
		binary.Write(writer, binary.BigEndian, box.boxLargeSize)
	}
	if box.fullBox {
		writer.WriteByte(byte(box.version))
		var tmp8 byte
		tmp8 = byte((box.flag >> 16) & 0xff)
		writer.WriteByte(tmp8)
		tmp8 = byte((box.flag >> 8) & 0xff)
		writer.WriteByte(tmp8)
		tmp8 = byte((box.flag >> 0) & 0xff)
		writer.WriteByte(tmp8)
	}

	writer.Write(box.writer.Bytes())
	return writer.Bytes()
}

func (box *MP4Box) Push8Bytes(data uint64) {
	binary.Write(box.writer, binary.BigEndian, data)
}

func (box *MP4Box) Push4Bytes(data uint32) {
	binary.Write(box.writer, binary.BigEndian, data)
}

func (box *MP4Box) Push2Bytes(data uint16) {
	binary.Write(box.writer, binary.BigEndian, data)
}

func (box *MP4Box) PushByte(data byte) {
	box.writer.WriteByte(data)
}

func (box *MP4Box) PushBytes(data []byte) {
	box.writer.Write(data)
}

func (box *MP4Box)PushBox(inBox *MP4Box){
	data:=inBox.Flush()
	box.writer.Write(data)
}