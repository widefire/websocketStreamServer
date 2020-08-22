package rtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
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
	Reports                   []*ReceptionReport
	ProfileSpecificExtensions []byte
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

	err = encodeExtensions(buffer, sr.ProfileSpecificExtensions, header.Padding)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

//Decode SenderReport
func (sr *SenderReport) Decode(data []byte) (err error) {
	if len(data) < HeaderLength+24 {
		err = fmt.Errorf("invalid data length %d", len(data))
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

	if header.PacketType != TypeSendReport {
		err = fmt.Errorf("bad header type for decode %d", header.PacketType)
		log.Println(err)
		return
	}

	reader := bytes.NewReader(data[HeaderLength:])
	err = binary.Read(reader, binary.BigEndian, &sr.SSRC)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &sr.NTPTime)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &sr.RTPTime)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &sr.SendersPacketCount)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &sr.SendersOctetCount)
	if err != nil {
		log.Println(err)
		return
	}

	sr.Reports, sr.ProfileSpecificExtensions, err = decodeReceptionReportAndExtension(data[HeaderLength+24:totalLength], header)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func (sr *SenderReport) isEq(rh *SenderReport) bool {
	lheader := sr.Header()
	rheader := rh.Header()
	if !lheader.isEq(rheader) {
		log.Fatalln("header not eq")
		return false
	}
	if !byteaIsEq(sr.ProfileSpecificExtensions, rh.ProfileSpecificExtensions) {
		log.Fatalln("ProfileSpecificExtensions not eq")
		return false
	}

	if len(sr.Reports) != len(rh.Reports) {
		log.Fatalln("reports count not eq")
		return false
	}
	for i := 0; i < len(sr.Reports); i++ {
		if !sr.Reports[i].isEq(rh.Reports[i]) {
			log.Fatalf("Reports %d not eq", i)
			return false
		}
	}

	return sr.SSRC == rh.SSRC &&
		sr.NTPTime == rh.NTPTime &&
		sr.RTPTime == rh.RTPTime &&
		sr.SendersPacketCount == rh.SendersPacketCount &&
		sr.SendersOctetCount == rh.SendersOctetCount

}

//Header generate this SenderReport's header
func (sr *SenderReport) Header() (header *Header) {
	return &Header{
		Padding:    PadLength(sr.ProfileSpecificExtensions) > 0,
		Count:      uint8(len(sr.Reports)),
		PacketType: TypeSendReport,
		Length:     uint16(sr.Len()/4 - 1),
	}
}

//Len the SenderReport total length
func (sr *SenderReport) Len() int {
	return HeaderLength + 24 + len(sr.Reports)*ReceptionReportLength + len(sr.ProfileSpecificExtensions) +
		PadLength(sr.ProfileSpecificExtensions)
}

func encodeExtensions(buffer *bytes.Buffer, extension []byte, padding bool) (err error) {

	if len(extension) > 0 {
		var n int
		n, err = buffer.Write(extension)
		if err != nil {
			log.Println(err)
			return
		}
		if n != len(extension) {
			err = fmt.Errorf("buffer write failed")
			log.Println(err)
			return
		}
	}

	if padding {
		padCount := PadLength(extension)
		if padCount == 0 {
			err = fmt.Errorf("pad flag is true but no need pad")
			log.Println(err)
			return
		}
		pad := make([]byte, padCount)
		pad[padCount-1] = byte(padCount)
		n, err := buffer.Write(pad)
		if err != nil {
			log.Println(err)
			return err
		}
		if n != padCount {
			err = fmt.Errorf("buffer write failed")
			log.Println(err)
			return err
		}
	}

	return
}

func decodeReceptionReportAndExtension(data []byte, header *Header) (reports []*ReceptionReport, profileSpecificExtensions []byte, err error) {
	reports = make([]*ReceptionReport, int(header.Count))
	cur := 0
	for i := 0; i < int(header.Count); i++ {
		reports[i] = &ReceptionReport{}
		err = reports[i].Decode(data[cur:])
		if err != nil {
			log.Println(err)
			return
		}
		cur += ReceptionReportLength
	}

	if cur < len(data) {
		if header.Padding {
			padCount := int(data[len(data)-1])
			if padCount > len(data)-cur {
				err = fmt.Errorf("pad count %d > total exten length %d", padCount, len(data)-cur)
				log.Println(err)
				return
			}
			profileSpecificExtensions = data[cur : len(data)-padCount]
		} else {
			profileSpecificExtensions = data[cur:]
		}
	}
	return
}
