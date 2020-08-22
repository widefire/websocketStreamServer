package rtcp

import (
	"bytes"
	"errors"
	"fmt"
	"log"
)

//Packet RTCP packet interface
type Packet interface {
	Header() (header *Header)
	Len() int
	Encode(buffer *bytes.Buffer) error
	Decode(data []byte) error
}

//CompoundPacket ...
type CompoundPacket struct {
	Packet []Packet
}

//Encode Compound packet
func (compound *CompoundPacket) Encode(buffer *bytes.Buffer) (err error) {
	for i, p := range compound.Packet {
		if p == nil {
			err = fmt.Errorf("packet %d is nil", i)
			log.Println(err)
			return
		}
		err = p.Encode(buffer)
		if err != nil {
			log.Println(err)
			return
		}
	}
	return
}

//Decode Compound packet
func (compound *CompoundPacket) Decode(data []byte) (err error) {
	compound.Packet = make([]Packet, 0)
	dataCount := len(data)
	if dataCount == 0 {
		err = errors.New("empty data to decode")
		log.Println(err)
		return
	}

	cur := 0
	for cur < dataCount {
		header := &Header{}
		err = header.Decode(data[cur:])
		if err != nil {
			log.Println(err)
			return
		}
		var packet Packet
		switch header.PacketType {
		case TypeSendReport:
			packet = &SenderReport{}
		case TypeReceiverReport:
			packet = &ReceiverReport{}
		case TypeSourceDescription:
			packet = &SDES{}
		case TypeGoodbye:
			packet = &GoodBye{}
		case TypeApplicationDefined:
			packet = &ApplicationDefined{}
		default:
			err = fmt.Errorf("unknown packet type %d", header.PacketType)
			log.Println(err)
			return
		}
		err = packet.Decode(data[cur:])
		if err != nil {
			log.Println(err)
			return
		}
		cur += packet.Len()
		compound.Packet = append(compound.Packet, packet)
	}

	return
}
