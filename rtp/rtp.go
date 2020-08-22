package rtp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

//Header ...
type Header struct {
	V              uint8    //version 2bit
	P              uint8    //padding 1bit
	X              uint8    //extension 1bit
	CC             uint8    //CSRC count 4bit
	M              uint8    //marker 1bit
	PT             uint8    //payload type 7bit
	SequenceNumber uint16   //sequence number uint16
	Timestamp      uint32   //timestamp uint32
	SSRC           uint32   //uint32
	CSRC           []uint32 //[]uint32
	*HeaderExtension
}

//NewHeader ...
func NewHeader() (header *Header) {
	return &Header{
		V:               2,
		P:               0,
		X:               0,
		CC:              0,
		M:               0,
		PT:              0,
		SequenceNumber:  0,
		Timestamp:       0,
		SSRC:            0,
		CSRC:            make([]uint32, 0),
		HeaderExtension: nil,
	}
}

//HeaderExtension ...
type HeaderExtension struct {
	DefineByProfile uint16 //uint16
	Length          uint16 //uint16
	HeaderExtension []byte
}

//Encode Header
func (header *Header) Encode(buffer *bytes.Buffer) (err error) {
	err = buffer.WriteByte((header.V << 6) | (header.P << 5) | (header.X << 4) | (header.CC))
	if err != nil {
		log.Println(err)
		return
	}
	err = buffer.WriteByte((header.M << 7) | header.PT)
	if err != nil {
		log.Println(err)
		return
	}

	err = binary.Write(buffer, binary.BigEndian, header.SequenceNumber)
	if err != nil {
		log.Println(err)
		return
	}

	err = binary.Write(buffer, binary.BigEndian, header.Timestamp)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(buffer, binary.BigEndian, header.SSRC)
	if err != nil {
		log.Println(err)
		return
	}

	if header.CC > 0xf || int(header.CC) != len(header.CSRC) {
		err = fmt.Errorf("CSRC count %d ,CC %d", len(header.CSRC), int(header.CC))
		log.Println(err)
		return
	}

	if header.CSRC != nil {
		for _, v := range header.CSRC {
			err = binary.Write(buffer, binary.BigEndian, v)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	if header.HeaderExtension != nil {
		err = binary.Write(buffer, binary.BigEndian, header.DefineByProfile)
		if err != nil {
			log.Println(err)
			return
		}
		err = binary.Write(buffer, binary.BigEndian, header.Length)
		if err != nil {
			log.Println(err)
			return
		}

		if len(header.HeaderExtension.HeaderExtension) != int(header.Length) {
			err = fmt.Errorf("exten length %d,but buffer length %d", int(header.Length), header.HeaderExtension.HeaderExtension)
			log.Println(err)
			return
		}
		n, err := buffer.Write(header.HeaderExtension.HeaderExtension)
		if err != nil {
			log.Println(err)
			return err
		}
		if n != len(header.HeaderExtension.HeaderExtension) {
			err = errors.New("buff write failed")
			return err
		}
	}

	return nil
}

//Decode Header
func (header *Header) Decode(data []byte) (offset int, err error) {
	if len(data) < 12 {
		err = fmt.Errorf("a rtp header at least 12 byte")
		log.Println(err)
		return
	}
	header.V = data[0] >> 6
	header.P = (data[0] >> 5) & 1
	header.X = (data[0] >> 4) & 1
	header.CC = data[0] & 0xf
	header.M = data[1] >> 7
	header.PT = data[1] & 0x7f

	reader := bytes.NewReader(data[2:])
	err = binary.Read(reader, binary.BigEndian, &header.SequenceNumber)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &header.Timestamp)
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Read(reader, binary.BigEndian, &header.SSRC)
	if err != nil {
		log.Println(err)
		return
	}

	offset = 12

	if header.CC != 0 {
		for i := 0; i < int(header.CC); i++ {
			if offset+4 > len(data) {
				err = errors.New("bad header data,out of range CSRC")
				log.Println(err)
				return
			}
			var csrc uint32
			err = binary.Read(reader, binary.BigEndian, &csrc)
			if err != nil {
				log.Println(err)
				return
			}
			header.CSRC = append(header.CSRC, csrc)
			offset += 4
		}
	}
	if header.X != 0 {
		if offset+4 > len(data) {
			err = errors.New("bad header data,out of range extension")
			log.Println(err)
			return
		}
		header.HeaderExtension = &HeaderExtension{}
		err = binary.Read(reader, binary.BigEndian, &header.DefineByProfile)
		if err != nil {
			log.Println(err)
			return
		}
		offset += 2
		err = binary.Read(reader, binary.BigEndian, &header.Length)
		if err != nil {
			log.Println(err)
			return
		}
		offset += 2
		if header.Length > 0 {
			if offset+int(header.Length) > len(data) {
				err = errors.New("bad header data,out of range extension data")
				log.Println(err)
				return
			}
			header.HeaderExtension.HeaderExtension = data[offset : offset+int(header.Length)]
			offset += int(header.Length)
		}
	}
	return
}
