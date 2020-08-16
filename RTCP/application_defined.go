package rtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

//ApplicationDefined ...
type ApplicationDefined struct {
	SSCSRC                   uint32
	Name                     []byte
	ApplicationDependentData []byte
	SubType                  byte
}

//Header APP header
func (app *ApplicationDefined) Header() (header *Header) {
	header = &Header{
		Padding:    len(app.ApplicationDependentData)%4 != 0,
		Count:      app.SubType,
		PacketType: TypeApplicationDefined,
		Length:     uint16(app.Len()/4 - 1),
	}
	return
}

//isEq
func (app *ApplicationDefined) isEq(rh *ApplicationDefined) bool {
	lheader := app.Header()
	rheader := rh.Header()
	if !lheader.isEq(rheader) {
		log.Println("header not eq")
		return false
	}
	if app.SSCSRC != rh.SSCSRC {
		log.Println("ssrc/csrc not eq")
		return false
	}
	if !byteaIsEq(app.Name, rh.Name) {
		log.Println("name not eq")
		return false
	}
	if !byteaIsEq(app.ApplicationDependentData, rh.ApplicationDependentData) {
		log.Println("app data not eq")
		return false
	}
	if app.SubType != rh.SubType {
		log.Println("sub type not eq")
		return false
	}

	return true
}

//Len ...
func (app *ApplicationDefined) Len() int {
	length := HeaderLength + 8
	if len(app.ApplicationDependentData) > 0 {
		padCount := PadLength(app.ApplicationDependentData)
		length += len(app.ApplicationDependentData)
		length += padCount
	}
	return length
}

//Encode encode App
func (app *ApplicationDefined) Encode(buffer *bytes.Buffer) (err error) {
	if app.SubType > 0x1f {
		err = fmt.Errorf("max subtype is 0x1f,now %d", app.SubType)
		log.Println(err)
		return
	}
	if len(app.Name) != 4 {
		err = fmt.Errorf("Name must 4bytes,now %d", len(app.Name))
		log.Println(err)
		return
	}

	header := app.Header()
	err = header.Encode(buffer)
	if err != nil {
		log.Println(err)
		return
	}

	err = binary.Write(buffer, binary.BigEndian, app.SSCSRC)
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range app.Name {
		err = buffer.WriteByte(v)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if len(app.ApplicationDependentData) > 0 {
		n := 0
		n, err = buffer.Write(app.ApplicationDependentData)
		if err != nil {
			log.Println(err)
			return
		}
		if n != len(app.ApplicationDependentData) {
			err = fmt.Errorf("write %d but return %d", len(app.ApplicationDependentData), n)
			log.Println(err)
			return
		}
		padcount := padLengthByCount(len(app.ApplicationDependentData))
		if padcount > 0 {
			pad := make([]byte, padcount)
			pad[padcount-1] = byte(padcount)
			n, err = buffer.Write(pad)
			if err != nil {
				log.Println(err)
				return
			}
			if n != padcount {
				err = fmt.Errorf("write %d but return %d", padcount, n)
				log.Println(err)
				return
			}
		}
	}
	return
}

//Decode decode App
func (app *ApplicationDefined) Decode(data []byte) (err error) {
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
	app.SubType = header.Count
	if header.PacketType != TypeApplicationDefined {
		err = fmt.Errorf("bad header type for decode %d", header.PacketType)
		log.Println(err)
		return
	}
	reader := bytes.NewReader(data[HeaderLength:])
	err = binary.Read(reader, binary.BigEndian, &app.SSCSRC)
	if err != nil {
		log.Println(err)
		return
	}

	app.Name = make([]byte, 4)
	for i := 0; i < 4; i++ {
		app.Name[i], err = reader.ReadByte()
		if err != nil {
			log.Println(err)
			return
		}
	}

	if totalLength > HeaderLength+8 {
		if header.Padding {
			lenAppDefWithPad := totalLength - HeaderLength - 8
			padCount := int(data[totalLength-1])
			if padCount > lenAppDefWithPad {
				err = errors.New("invalid pad count,out of range")
				log.Println(err)
				return
			}
			app.ApplicationDependentData = data[HeaderLength+8 : totalLength-padCount]
		} else {
			app.ApplicationDependentData = data[HeaderLength+8 : totalLength]

		}
	}

	return
}
