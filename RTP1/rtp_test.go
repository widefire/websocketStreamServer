package rtp

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

func TestHeaderEncodeDecode(t *testing.T) {
	encHeader := NewHeader()

	encHeader.V = 2
	encHeader.P = 1
	encHeader.X = 1
	encHeader.CC = 5
	encHeader.PT = 99
	encHeader.SequenceNumber = 5678
	encHeader.Timestamp = 0x1f2f3f
	encHeader.SSRC = 0x2f3f5f
	for i := 0; i < int(encHeader.CC); i++ {
		encHeader.CSRC = append(encHeader.CSRC, uint32(0x123456+i))
	}
	encHeader.HeaderExtension = &HeaderExtension{}
	encHeader.DefineByProfile = 22
	encHeader.Length = 32
	encHeader.HeaderExtension.HeaderExtension = make([]byte, 32)
	for i := 0; i < 32; i++ {
		encHeader.HeaderExtension.HeaderExtension[i] = byte(i)
	}

	encBuf := new(bytes.Buffer)
	err := encHeader.Encode(encBuf)
	if err != nil {
		log.Fatalln(err)
	}
	encData := encBuf.Bytes()
	log.Println(len(encData))

	decHeader := NewHeader()
	offset, err := decHeader.Decode(encData)
	if err != nil {
		log.Panic(err)
	}
	log.Println(fmt.Sprintf("offset %d", offset))

	if encHeader.V != decHeader.V ||
		encHeader.P != decHeader.P ||
		encHeader.X != decHeader.X ||
		encHeader.CC != decHeader.CC ||
		encHeader.M != decHeader.M ||
		encHeader.PT != decHeader.PT ||
		encHeader.SequenceNumber != decHeader.SequenceNumber ||
		encHeader.Timestamp != decHeader.Timestamp ||
		encHeader.SSRC != decHeader.SSRC ||
		encHeader.DefineByProfile != decHeader.DefineByProfile ||
		encHeader.Length != decHeader.Length {
		t.Errorf("header value not eq")
	}
	for i := 0; i < len(encHeader.CSRC); i++ {
		if encHeader.CSRC[i] != decHeader.CSRC[i] {
			t.Errorf("csrc not eq")
		}
	}
	log.Println(decHeader.SSRC)
	if 0 != bytes.Compare(encHeader.HeaderExtension.HeaderExtension, decHeader.HeaderExtension.HeaderExtension) {
		t.Errorf("extension data error")
	}
}
