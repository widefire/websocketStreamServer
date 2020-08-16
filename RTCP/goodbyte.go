package rtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

//GoodByte ...
type GoodByte struct {
	SSCSRC           []uint32
	ReasonForLeaving string
}

//Header GoodBye header
func (gb *GoodByte) Header() (header *Header) {

	header = &Header{
		Padding:    false,
		Count:      uint8(len(gb.SSCSRC)),
		PacketType: TypeGoodbye,
		Length:     uint16(gb.Len()/4 - 1),
	}
	return
}

//Len GoodBye length
func (gb *GoodByte) Len() int {
	length := HeaderLength
	length += 4 * len(gb.SSCSRC)
	bytes := []byte(gb.ReasonForLeaving)
	if len(bytes) > 0 {
		length++
		length += len(bytes)
		length += padLengthByCount(length)
	}
	return length
}

//Encode GoodBye
func (gb *GoodByte) Encode(buffer *bytes.Buffer) (err error) {
	if len(gb.SSCSRC) > 0x1f {
		err = fmt.Errorf("invalid reports count %d", len(gb.SSCSRC))
		log.Println(err)
		return
	}

	header := gb.Header()
	err = header.Encode(buffer)
	if err != nil {
		log.Println(err)
		return
	}

	for _, i := range gb.SSCSRC {
		err = binary.Write(buffer, binary.BigEndian, i)
		if err != nil {
			log.Println(err)
			return
		}
	}

	bytes := []byte(gb.ReasonForLeaving)
	if len(bytes) > 0 {
		if len(bytes) > 0xff {
			err = fmt.Errorf("reason max size is 0xff,but now %d", len(bytes))
			log.Println(err)
			return
		}
		validLength := 1 + len(bytes)
		padCount := padLengthByCount(validLength)
		reasonLength := validLength + padCount
		reason := make([]byte, reasonLength)
		reason[0] = byte(len(bytes))
		copy(reason[1:], bytes)
		n := 0
		n, err = buffer.Write(reason)
		if err != nil {
			log.Println(err)
			return
		}
		if n != len(reason) {
			err = fmt.Errorf("write target %d,but return %d", len(reason), n)
			log.Println(err)
			return
		}
	}

	return
}

//Decode GoodBye
func (gb *GoodByte) Decode(data []byte) (err error) {

	if 0 != len(data)%4 {
		err = errors.New("not 32bit align")
		log.Println(err)
		return
	}

	header := &Header{}
	err = header.Decode(data)
	if err != nil {
		log.Println(err)
		return
	}
	totalLength := (int(header.Length) + 1) * 4
	if totalLength > len(data) {
		err = errors.New("header length > packet length")
		log.Println(err)
		return
	}

	if header.PacketType != TypeGoodbye {
		err = fmt.Errorf("bad header type for decode %d", header.PacketType)
		log.Println(err)
		return
	}

	reader := bytes.NewReader(data[HeaderLength:])
	gb.SSCSRC = make([]uint32, header.Count)
	cur := HeaderLength
	for i := 0; i < int(header.Count); i++ {
		err = binary.Read(reader, binary.BigEndian, &gb.SSCSRC[i])
		if err != nil {
			log.Println(err)
			return
		}
		cur += 4
	}

	if cur < totalLength {
		reasonLength := int(data[cur])
		reasonEnd := cur + reasonLength + 1
		if reasonEnd > totalLength {
			err = errors.New("invalid reason,out of range")
			log.Println(err)
			return
		}
		gb.ReasonForLeaving = string(data[cur+1 : reasonEnd])
	}

	return
}

func (gb *GoodByte) isEq(rh *GoodByte) bool {
	lheader := gb.Header()
	rheader := rh.Header()
	if !lheader.isEq(rheader) {
		log.Println("header not eq")
		return false
	}
	if len(gb.SSCSRC) != len(rh.SSCSRC) {
		log.Println("ssrc/csrc count not eq")
		return false
	}
	for i := 0; i < len(gb.SSCSRC); i++ {
		if gb.SSCSRC[i] != rh.SSCSRC[i] {
			log.Printf("ssrc/csrc %d not eq,%d!=%d\r\n", i, gb.SSCSRC[i], rh.SSCSRC[i])
			return false
		}
	}

	if gb.ReasonForLeaving != rh.ReasonForLeaving {
		log.Print("reason for leaving not eq")
		return false
	}

	return true
}
