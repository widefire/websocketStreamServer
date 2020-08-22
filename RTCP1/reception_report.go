package rtcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

//ReceptionReport ...
type ReceptionReport struct {
	SSRC                          uint32
	FractionLost                  uint8
	CumulativeLost                uint32
	ExtendedHighestSequenceNumber uint32
	Jitter                        uint32
	LSR                           uint32
	DLSR                          uint32
}

//Encode ReceptionReport
func (rr *ReceptionReport) Encode(buffer *bytes.Buffer) (err error) {
	if rr.CumulativeLost > 0xffffff {
		err = fmt.Errorf("invalid Cumulative lost %d", rr.CumulativeLost)
		log.Println(err)
		return
	}

	err = binary.Write(buffer, binary.BigEndian, rr.SSRC)
	if err != nil {
		log.Println(err)
		return
	}
	err = buffer.WriteByte(rr.FractionLost)
	if err != nil {
		log.Println(err)
		return
	}
	err = buffer.WriteByte(byte(rr.CumulativeLost >> 16 & 0xff))
	if err != nil {
		log.Println(err)
		return
	}
	err = buffer.WriteByte(byte(rr.CumulativeLost >> 8 & 0xff))
	if err != nil {
		log.Println(err)
		return
	}
	err = buffer.WriteByte(byte(rr.CumulativeLost >> 0 & 0xff))
	if err != nil {
		log.Println(err)
		return
	}

	err = binary.Write(buffer, binary.BigEndian, rr.ExtendedHighestSequenceNumber)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, rr.Jitter)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, rr.LSR)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, rr.DLSR)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

//Decode ReceptionReport
func (rr *ReceptionReport) Decode(data []byte) (err error) {
	if len(data) < 24 {
		err = fmt.Errorf("invlaid reception report len %d", len(data))
		return
	}

	reader := bytes.NewReader(data)
	err = binary.Read(reader, binary.BigEndian, &rr.SSRC)
	if err != nil {
		log.Println(err)
		return
	}
	rr.FractionLost, err = reader.ReadByte()
	if err != nil {
		log.Println(err)
		return
	}
	losts := make([]byte, 3)
	losts[0], err = reader.ReadByte()
	if err != nil {
		log.Println(err)
		return
	}
	losts[1], err = reader.ReadByte()
	if err != nil {
		log.Println(err)
		return
	}
	losts[2], err = reader.ReadByte()
	if err != nil {
		log.Println(err)
		return
	}
	rr.CumulativeLost = uint32(losts[0])<<16 | uint32(losts[1])<<8 | uint32(losts[2])

	err = binary.Read(reader, binary.BigEndian, &rr.ExtendedHighestSequenceNumber)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &rr.Jitter)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &rr.LSR)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &rr.DLSR)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func (rr *ReceptionReport) isEq(rh *ReceptionReport) bool {
	return rr.SSRC == rh.SSRC &&
		rr.FractionLost == rh.FractionLost &&
		rr.CumulativeLost == rh.CumulativeLost &&
		rr.ExtendedHighestSequenceNumber == rh.ExtendedHighestSequenceNumber &&
		rr.Jitter == rh.Jitter &&
		rr.LSR == rh.LSR &&
		rr.DLSR == rh.DLSR
}

//Len length of ReceptionReport must 24
func (rr *ReceptionReport) Len() int {
	return ReceptionReportLength
}

//ReceptionReportLength ...
const ReceptionReportLength int = 24
