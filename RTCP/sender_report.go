package rtcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

//SenderReport SR
type SenderReport struct {
	SSRC                      uint32
	NTPTime                   uint64
	RTPTime                   uint32
	SendersPacketCount        uint32
	SendersOctetCount         uint32
	Reports                   []ReceptionReport
	ProfileSpecificExtensions []byte
	Padding                   []byte
}

//Encode SenderReport
func (sr *SenderReport) Encode(buffer *bytes.Buffer) (err error) {
	if len(sr.Reports) > 0x1f {
		err = fmt.Errorf("invalid reports count %d", len(sr.Reports))
		log.Println(err)
		return
	}

	header := sr.Header()
	err = header.Encode(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, sr.SSRC)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, sr.NTPTime)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, sr.RTPTime)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, sr.SendersPacketCount)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, sr.SendersOctetCount)
	if err != nil {
		log.Println(err)
		return
	}
	for _, reports := range sr.Reports {
		err = reports.Encode(buffer)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if len(sr.ProfileSpecificExtensions) > 0 {
		var n int
		n, err = buffer.Write(sr.ProfileSpecificExtensions)
		if err != nil {
			log.Println(err)
			return
		}
		if n != len(sr.ProfileSpecificExtensions) {
			err = fmt.Errorf("buffer write failed")
			log.Println(err)
			return
		}
	}
	if len(sr.Padding) > 0 {
		var n int
		n, err = buffer.Write(sr.Padding)
		if err != nil {
			log.Println(err)
			return
		}
		if n != len(sr.Padding) {
			err = fmt.Errorf("buffer write failed")
			log.Println(err)
			return
		}
	}

	return
}

//Decode SenderReport
func (sr *SenderReport) Decode(data []byte) (err error) {
	if len(data) < HeaderLength+24 {
		err = fmt.Errorf("invalid data length %d", len(data))
		return
	}

	return
}

//Header generate this SenderReport's header
func (sr *SenderReport) Header() (header *Header) {
	return &Header{
		Padding:              sr.HasPadding(),
		ReceptionReportCount: uint8(len(sr.Reports)),
		PacketType:           TypeSendReport,
		Length:               uint16(sr.Len()/4 - 1),
	}
}

//Len the SenderReport total length
func (sr *SenderReport) Len() int {
	extensionLength := 0
	if len(sr.ProfileSpecificExtensions) > 0 {

	} else {

	}
	return HeaderLength + 24 + len(sr.Padding) + len(sr.Reports)*ReceptionReportLength + len(sr.ProfileSpecificExtensions)
}

//HasPadding a sr has padding
func (sr *SenderReport) HasPadding() bool {
	return len(sr.ProfileSpecificExtensions) > 0
}
