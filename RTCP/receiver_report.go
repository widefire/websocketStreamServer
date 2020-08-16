package rtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

//ReceiverReport RR
type ReceiverReport struct {
	SSRC                      uint32
	Reports                   []*ReceptionReport
	ProfileSpecificExtensions []byte
}

//Encode ReceiverReport
func (rr *ReceiverReport) Encode(buffer *bytes.Buffer) (err error) {
	if len(rr.Reports) > 0x1f {
		err = fmt.Errorf("invalid reports count %d", len(rr.Reports))
		log.Println(err)
		return
	}
	header := rr.Header()
	err = header.Encode(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, rr.SSRC)
	if err != nil {
		log.Println(err)
		return
	}

	for _, reports := range rr.Reports {
		err = reports.Encode(buffer)
		if err != nil {
			log.Println(err)
			return
		}
	}
	err = encodeExtensions(buffer, rr.ProfileSpecificExtensions, header.Padding)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

//Decode ReceiverReport
func (rr *ReceiverReport) Decode(data []byte) (err error) {
	if len(data) < HeaderLength+4 {
		err = fmt.Errorf("too less packet size %d ,at least %d", len(data), HeaderLength+4)
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

	if header.PacketType != TypeReceiverReport {
		err = fmt.Errorf("bad header type for decode %d", header.PacketType)
		log.Println(err)
		return
	}

	reader := bytes.NewReader(data[HeaderLength:])
	err = binary.Read(reader, binary.BigEndian, &rr.SSRC)
	if err != nil {
		log.Println(err)
		return
	}

	rr.Reports, rr.ProfileSpecificExtensions, err = decodeReceptionReportAndExtension(data[HeaderLength+4:totalLength], header)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func (rr *ReceiverReport) isEq(rh *ReceiverReport) bool {
	if !byteaIsEq(rr.ProfileSpecificExtensions, rh.ProfileSpecificExtensions) {
		log.Fatal("profilespecificextension not eq")
		return false
	}
	if len(rr.Reports) != len(rh.Reports) {
		log.Fatal("report count not eq")
		return false
	}
	for i := 0; i < len(rr.Reports); i++ {
		if !rr.Reports[i].isEq(rh.Reports[i]) {
			log.Fatalf("reports %d not eq", i)
			return false
		}
	}

	return rr.SSRC == rh.SSRC
}

//Header generate this ReciverReport's header
func (rr *ReceiverReport) Header() (header *Header) {
	return &Header{
		Padding:    PadLength(rr.ProfileSpecificExtensions) > 0,
		Count:      uint8(len(rr.Reports)),
		PacketType: TypeReceiverReport,
		Length:     uint16(rr.Len()/4 - 1),
	}
}

//Len the ReceiverReport total length
func (rr *ReceiverReport) Len() int {
	return HeaderLength + 4 + len(rr.Reports)*ReceptionReportLength + len(rr.ProfileSpecificExtensions) +
		PadLength(rr.ProfileSpecificExtensions)
}
