package rtcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

//RtpVersion  must 2
const RtpVersion uint8 = 2

//PacketType enum
type PacketType uint8

//RTCP packet types
const (
	TypeSendReport         PacketType = 200 // SR
	TypeReceiverReport     PacketType = 201 // RR
	TypeSourceDescription  PacketType = 202 // SDES
	TypeGoodbye            PacketType = 203 // BYE
	TypeApplicationDefined PacketType = 204 // APP
)

//HeaderLength rtcp header length
const HeaderLength = 4

//Header rtcp common header
type Header struct {
	Padding              bool
	ReceptionReportCount uint8
	PacketType           PacketType
	Length               uint16
}

//Encode rtcp header
func (header *Header) Encode(buffer *bytes.Buffer) (err error) {
	if header.ReceptionReportCount > 0x1f {
		err = fmt.Errorf("invalid reception report count %d", header.ReceptionReportCount)
		log.Println(err)
		return
	}

	b0 := RtpVersion << 6
	if header.Padding {
		b0 |= 0x20
	}
	b0 |= header.ReceptionReportCount
	err = buffer.WriteByte(b0)
	if err != nil {
		log.Println(err)
		return
	}
	err = buffer.WriteByte(byte(header.PacketType))
	if err != nil {
		log.Println(err)
		return
	}

	err = binary.Write(buffer, binary.BigEndian, header.Length)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

//Decode rtcp header
func (header *Header) Decode(data []byte) (err error) {
	if len(data) < 4 {
		err = fmt.Errorf("invliad header buffer len %d", len(data))
		log.Println(err)
		return
	}

	version := data[0] >> 6
	if version != RtpVersion {
		err = fmt.Errorf("invliad rtp version %d", version)
		log.Println(err)
		return
	}

	header.Padding = (data[0] >> 5 & 0x1) != 0
	header.ReceptionReportCount = data[0] & 0x1f
	header.PacketType = PacketType(data[1])
	reader := bytes.NewReader(data[2:])
	err = binary.Read(reader, binary.BigEndian, &header.Length)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func (header *Header) isEq(rh *Header) bool {
	return header.Padding == rh.Padding &&
		header.ReceptionReportCount == rh.ReceptionReportCount &&
		header.PacketType == rh.PacketType &&
		header.Length == rh.Length
}

func byteaIsEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

//PadLength get pad count
func PadLength(extern []byte) int {
	externLength := len(extern)
	if externLength > 0 {
		mod := externLength % 4
		if mod != 0 {
			return 4 - mod
		}
	}
	return 0
}
