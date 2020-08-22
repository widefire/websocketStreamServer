package rtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

//SDESType uint8
type SDESType uint8

//SDES types enum
const (
	SDESEnd     SDESType = 0
	SDESCName   SDESType = 1
	SDESName    SDESType = 2
	SDESEmail   SDESType = 3
	SDESPhone   SDESType = 4
	SDESLoc     SDESType = 5
	SDESTool    SDESType = 6
	SDESNote    SDESType = 7
	SDESPrivate SDESType = 8
)

//SDESItem ...
type SDESItem struct {
	Type SDESType
	Text string
}

//Len SDESItem length
func (item *SDESItem) Len() int {
	return 2 + len([]byte(item.Text))
}

//Encode SDESItem
func (item *SDESItem) Encode(buffer *bytes.Buffer) (err error) {
	if item.Type == SDESEnd {
		err = errors.New("end of sdes item not encode")
		log.Println(err)
		return
	}

	err = buffer.WriteByte(byte(item.Type))
	if err != nil {
		log.Println(err)
		return
	}
	bytes := []byte(item.Text)
	if len(bytes) > 0xff {
		err = fmt.Errorf("text max size is 255,but now %d", len(bytes))
		log.Println(err)
		return
	}
	err = buffer.WriteByte(byte(len(bytes)))
	if err != nil {
		log.Println(err)
		return
	}
	n := 0
	n, err = buffer.Write(bytes)
	if err != nil {
		log.Println(err)
		return
	}
	if n != len(bytes) {
		err = fmt.Errorf("Write bytes failed ,%d < %d ", n, len(bytes))
		log.Println(err)
		return
	}

	return
}

//Decode SDESItem
func (item *SDESItem) Decode(data []byte) (err error) {
	if len(data) < 2 {
		err = fmt.Errorf("a SDESItem at least two byte,now %d", len(data))
		log.Println(err)
		return
	}
	item.Type = SDESType(data[0])
	bytesLen := int(data[1])
	if bytesLen > len(data)-2 {
		err = fmt.Errorf("SDESItem length error,total %d but length %d", len(data), bytesLen)
		log.Println(err)
		return
	}
	bytes := data[2 : 2+bytesLen]
	item.Text = string(bytes)
	return
}

func (item *SDESItem) isEq(rh *SDESItem) bool {
	return item.Type == rh.Type &&
		item.Text == rh.Text
}

//SDESChunk ...
type SDESChunk struct {
	SSCSRC uint32 //SSRC/CSRC
	Items  []*SDESItem
}

//Len SDESChunk length
func (chunk *SDESChunk) Len() int {
	length := 5
	for _, item := range chunk.Items {
		length += item.Len()
	}
	length += padLengthByCount(length)
	return length
}

func (chunk *SDESChunk) isEq(rh *SDESChunk) bool {
	if chunk.SSCSRC != rh.SSCSRC {
		log.Println("ssrc/csrc not eq")
		return false
	}

	if len(chunk.Items) != len(rh.Items) {
		log.Println("item count not eq")
		return false
	}

	for i := 0; i < len(chunk.Items); i++ {
		if !chunk.Items[i].isEq(rh.Items[i]) {
			log.Printf("item %d not eq", i)
			return false
		}
	}

	return true
}

//Encode SDESChunk
func (chunk *SDESChunk) Encode(buffer *bytes.Buffer) (err error) {
	err = binary.Write(buffer, binary.BigEndian, chunk.SSCSRC)
	if err != nil {
		log.Println(err)
		return
	}

	for _, item := range chunk.Items {
		err = item.Encode(buffer)
		if err != nil {
			log.Println(err)
			return
		}
	}

	err = buffer.WriteByte(byte(SDESEnd))
	if err != nil {
		log.Println(err)
		return
	}

	length := 5
	for _, item := range chunk.Items {
		length += item.Len()
	}

	padLength := padLengthByCount(length)
	if padLength != 0 {
		pad := make([]byte, padLength)
		pad[padLength-1] = byte(padLength)
		n := 0
		n, err = buffer.Write(pad)
		if err != nil {
			log.Println(err)
			return
		}
		if n != padLength {
			err = errors.New("write buff failed,n is too small")
			log.Println(err)
			return
		}
	}

	return
}

//Decode SDESChunk
func (chunk *SDESChunk) Decode(data []byte) (err error) {
	if len(data) < 5 {
		err = fmt.Errorf("a chunk at least 5 byte,but now %d", len(data))
		log.Println(err)
		return
	}

	reader := bytes.NewReader(data)
	err = binary.Read(reader, binary.BigEndian, &chunk.SSCSRC)
	if err != nil {
		log.Println(err)
		return
	}

	chunk.Items = make([]*SDESItem, 0)
	cur := 4
	for {
		if cur >= len(data) {
			err = fmt.Errorf("vad packet ,out of range")
			log.Println(err)
			return
		}
		itemType := SDESType(data[cur])
		if itemType == SDESEnd {
			return
		}
		item := &SDESItem{}
		err = item.Decode(data[cur:])
		if err != nil {
			log.Println(err)
			return
		}
		chunk.Items = append(chunk.Items, item)
		cur += item.Len()
	}

}

//SDES source description
type SDES struct {
	Chunks []*SDESChunk
}

//Len SDES length
func (sdes *SDES) Len() int {
	length := HeaderLength
	for _, chunk := range sdes.Chunks {
		length += chunk.Len()
	}
	return length
}

//Header SDES header
func (sdes *SDES) Header() (header *Header) {
	header = &Header{
		Padding:    false,
		Count:      uint8(len(sdes.Chunks)),
		PacketType: TypeSourceDescription,
		Length:     uint16(sdes.Len()/4 - 1),
	}
	return
}

func (sdes *SDES) isEq(rh *SDES) bool {
	lheader := sdes.Header()
	rheader := rh.Header()
	if !lheader.isEq(rheader) {
		log.Fatalln("header not eq")
		return false
	}

	if len(sdes.Chunks) != len(rh.Chunks) {
		log.Println("chunk count not eq")
		return false
	}

	for i := 0; i < len(sdes.Chunks); i++ {
		if !sdes.Chunks[i].isEq(rh.Chunks[i]) {
			log.Printf("chunk %d not eq \r\n", i)
			return false
		}
	}

	return true
}

//Encode SDES
func (sdes *SDES) Encode(buffer *bytes.Buffer) (err error) {
	if len(sdes.Chunks) > 0x1f {
		err = fmt.Errorf("chunk max count 31,but now %d", len(sdes.Chunks))
		log.Println(err)
		return
	}
	header := sdes.Header()
	err = header.Encode(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	for _, chunk := range sdes.Chunks {
		err = chunk.Encode(buffer)
		if err != nil {
			log.Println(err)
			return
		}
	}
	return
}

//Decode SDES
func (sdes *SDES) Decode(data []byte) (err error) {
	header := &Header{}
	err = header.Decode(data)
	if err != nil {
		log.Println(err)
		return
	}
	if header.PacketType != TypeSourceDescription {
		err = errors.New("not SDES type")
		log.Println(err)
		return
	}

	sdes.Chunks = make([]*SDESChunk, 0)
	cur := HeaderLength
	for {
		if cur == len(data) {
			break
		}
		if cur > len(data) {
			err = errors.New("invalid data,out of range")
			log.Println(err)
			return
		}
		chunk := &SDESChunk{}
		err = chunk.Decode(data[cur:])
		if err != nil {
			log.Println(err)
		}
		cur += chunk.Len()
		sdes.Chunks = append(sdes.Chunks, chunk)
	}

	if len(sdes.Chunks) != len(sdes.Chunks) {
		err = fmt.Errorf("header chunk count=%d,but decode chunk count=%d", header.Count, len(sdes.Chunks))
		log.Println(err)
		return
	}

	return
}
