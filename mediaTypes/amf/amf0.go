package amf

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"logger"
)

const (
	AMF0_number         = 0x00
	AMF0_boolean        = 0x01
	AMF0_string         = 0x02
	AMF0_object         = 0x03 //hash table=amfstring:key ,amftype:value,0x 00 00 09 end
	AMF0_movieclip      = 0x04
	AMF0_null           = 0x05
	AMF0_undefined      = 0x06
	AMF0_reference      = 0x07
	AMF0_ecma_array     = 0x08 //object with size of hashTable
	AMF0_object_end     = 0x09
	AMF0_strict_array   = 0x0a //arrycount+propArray
	AMF0_date           = 0x0b
	AMF0_long_string    = 0x0c
	AMF0_unsupported    = 0x0d
	AMF0_recordset      = 0x0e
	AMF0_xml_document   = 0x0f
	AMF0_typed_object   = 0x10
	AMF0_avmplus_object = 0x11
)

type AMF0Object struct {
	Props list.List
}
type AMF0Property struct {
	PropType int32
	Name     string
	Value    AMF0Data
}
type AMF0Data struct {
	StrValue  string
	NumValue  float64
	S16Value  int16
	BoolValue bool
	ObjValue  AMF0Object
}

type AMF0Encoder struct {
	writer *bytes.Buffer
}

func (this *AMF0Encoder) Init() {
	this.writer = new(bytes.Buffer)
}

func (this *AMF0Encoder) EncodeString(str string) (err error) {
	length := uint32(len(str))
	if length == 0 {
		return errors.New("invald string")
	}
	if length >= 0xffff {
		err = this.writer.WriteByte(AMF0_long_string)
		if err != nil {
			return err
		}
		err = binary.Write(this.writer, binary.BigEndian, &length)
		if err != nil {
			return err
		}
	} else {
		err = this.writer.WriteByte(AMF0_string)
		if err != nil {
			return err
		}
		var tmp16 uint16
		tmp16 = uint16(length)
		err = binary.Write(this.writer, binary.BigEndian, &tmp16)
		if err != nil {
			return err
		}
	}

	_, err = this.writer.Write([]byte(str))
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeNumber(num float64) (err error) {
	err = this.writer.WriteByte(AMF0_number)
	if err != nil {
		return err
	}
	err = binary.Write(this.writer, binary.BigEndian, &num)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder)EncodeNU(){

}

func (this *AMF0Encoder) EncodeBool(boo bool) (err error) {
	err = this.writer.WriteByte(AMF0_boolean)
	if err != nil {
		return err
	}
	var tmp8 byte
	if boo {
		tmp8 = 1
	} else {
		tmp8 = 0
	}
	err = this.writer.WriteByte(tmp8)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeInt16(num int16) (err error) {
	err = binary.Write(this.writer, binary.BigEndian, &num)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeInt24(num int32) (err error) {
	var tmp8 int8
	tmp8 = int8((num >> 16) & 0xff)
	this.writer.WriteByte(byte(tmp8))
	tmp8 = int8((num >> 8) & 0xff)
	this.writer.WriteByte(byte(tmp8))
	tmp8 = int8((num >> 0) & 0xff)
	this.writer.WriteByte(byte(tmp8))

	return nil
}

func (this *AMF0Encoder) EncodeInt32(num int32) (err error) {
	err = binary.Write(this.writer, binary.BigEndian, &num)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeUint32(num uint32) (err error) {
	err = binary.Write(this.writer, binary.BigEndian, &num)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeInt32LittleEndian(num int32) (err error) {
	err = binary.Write(this.writer, binary.LittleEndian, &num)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeNamedString(name string, str string) (err error) {
	this.EncodeInt16(int16(len(name)))
	this.writer.Write([]byte(name))

	err = this.EncodeString(str)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeNamedBool(name string, boolean bool) (err error) {
	this.EncodeInt16(int16(len(name)))
	this.writer.Write([]byte(name))

	err = this.EncodeBool(boolean)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) EncodeNamedNumber(name string, num float64) (err error) {
	this.EncodeInt16(int16(len(name)))
	this.writer.Write([]byte(name))

	err = this.EncodeNumber(num)
	if err != nil {
		return err
	}
	return err
}

func (this *AMF0Encoder) AppendByteArray(data []byte) (err error) {
	_, err = this.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (this *AMF0Encoder) AppendByte(data byte) (err error) {
	err = this.writer.WriteByte(data)
	return err
}

func (this *AMF0Encoder) GetData() (data []byte, err error) {
	if this.writer == nil {
		return nil, errors.New("not inited amf0 encoder")
	}
	return this.writer.Bytes(), nil
}

func (this *AMF0Encoder) GetDataSize() int {
	return len(this.writer.Bytes())
}

func (this *AMF0Encoder) EncodeAMFObj(obj *AMF0Object) {
	for v := obj.Props.Front(); v != nil; v = v.Next() {
		ret := this.encodeProp(v.Value.(*AMF0Property))
		if len(ret) > 0 {
			this.writer.Write(ret)
		}
	}
	return
}

func (this *AMF0Encoder) encodeObj(obj *AMF0Object) (data []byte) {
	//each prop +obj_end
	enc := &AMF0Encoder{}
	enc.Init()
	for v := obj.Props.Front(); v != nil; v = v.Next() {
		enc.AppendByteArray(this.encodeProp(v.Value.(*AMF0Property)))
	}
	enc.EncodeInt24(AMF0_object_end)
	data, _ = enc.GetData()
	return data
}

func (this *AMF0Encoder) encodeProp(prop *AMF0Property) (data []byte) {
	enc := &AMF0Encoder{}
	enc.Init()

	//has name
	if len(prop.Name) > 0 {
		enc.EncodeString(prop.Name)
	}
	//encode type
	switch prop.PropType {
	case AMF0_number:
		enc.EncodeNumber(prop.Value.NumValue)
	case AMF0_boolean:
		enc.EncodeBool(prop.Value.BoolValue)
	case AMF0_string:
		enc.EncodeString(prop.Value.StrValue)
	case AMF0_object:
		enc.AppendByte(AMF0_object)
		enc.AppendByteArray(this.encodeObj(&prop.Value.ObjValue))
	case AMF0_null:
		enc.AppendByte(AMF0_null)
	case AMF0_ecma_array:
		enc.AppendByte(AMF0_ecma_array)
		//size
		//object
		tmp := enc.encodeObj(&prop.Value.ObjValue)
		tmpSize := int32(len(tmp))
		enc.EncodeInt32(tmpSize)
		enc.AppendByteArray(tmp)
	case AMF0_strict_array:
		enc.AppendByte(AMF0_strict_array)
		enc.EncodeInt32(int32(prop.Value.ObjValue.Props.Len()))
		for v := prop.Value.ObjValue.Props.Front(); v != nil; v = v.Next() {
			enc.AppendByteArray(this.encodeProp(v.Value.(*AMF0Property)))
		}

	case AMF0_date:
		enc.AppendByte(AMF0_date)
		binary.Write(enc.writer, binary.BigEndian, &prop.Value.NumValue)
		enc.EncodeInt16(prop.Value.S16Value)
	case AMF0_long_string:
		enc.EncodeString(prop.Value.StrValue)
	default:
		logger.LOGW(fmt.Sprintf("not support amf type:%d", prop.PropType))
	}
	data, _ = enc.GetData()
	return
}

//解码
func AMF0DecodeInt16(data []byte) (ret uint16, err error) {
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &ret)
	return ret, err
}

func AMF0DecodeInt24(data []byte) (ret uint32, err error) {
	ret = ((uint32(data[0])) << 16) | ((uint32(data[1])) << 8) | ((uint32(data[2])) << 0)
	return ret, nil
}

func AMF0DecodeInt32(data []byte) (ret uint32, err error) {
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &ret)
	return ret, err
}

func AMF0DecodeInt32LE(data []byte) (ret uint32, err error) {
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &err)
	return ret, err
}

func AMF0DecodeNumber(data []byte) (ret float64, err error) {
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &ret)
	return ret, err
}

func AMF0DecodeBoolean(data []byte) (ret bool, err error) {
	if data[0] == 0 {
		ret = false
	} else {
		ret = true
	}
	return ret, nil
}

func AMF0DecodeString(data []byte) (ret string, err error) {
	strLength, err := AMF0DecodeInt16(data)
	if err != nil {
		return ret, err
	}
	tmpArray := make([]byte, strLength)
	_, err = bytes.NewReader(data[2:]).Read(tmpArray)
	ret = string(tmpArray)
	return ret, err
}

func AMFDecodeLongString(data []byte) (ret string, err error) {
	strLength, err := AMF0DecodeInt32(data)
	if err != nil {
		return ret, err
	}
	tmpArray := make([]byte, strLength)
	_, err = bytes.NewReader(data[4:]).Read(tmpArray)
	ret = string(tmpArray)
	return ret, err
}

func AMF0DecodeObj(data []byte) (ret *AMF0Object, err error) {
	ret, _, err = amf0DecodeObj(data, false)
	return ret, err
}

func amf0DecodeObj(data []byte, decodeName bool) (ret *AMF0Object, sizeUsed int32, err error) {
	ret = &AMF0Object{}
	var start, end int32
	start = 0
	end = int32(len(data))
	for start < end {
		if end-start >= 3 {
			endType, _ := AMF0DecodeInt24(data[start:])
			if endType == AMF0_object_end {
				start += 3
				err = nil
				break
			}
		}
		if err != nil {
			break
		}
		propGeted, bytesUsed, err := amf0DecodeProp(data[start:], decodeName)
		if err != nil {
			break
		}
		if propGeted == nil || bytesUsed < 1 {
			break
		} else {
			start += bytesUsed
			ret.Props.PushBack(propGeted)
		}

	}
	if err != nil {
		return nil, -1, err
	}

	sizeUsed = start

	return ret, sizeUsed, err
}

func amf0DecodeProp(data []byte, decodeName bool) (ret *AMF0Property, sizeUsed int32, err error) {
	ret = &AMF0Property{}
	sizeUsed = 0
	err = nil

	if decodeName {
		if len(data) < 4 {
			return ret, sizeUsed, errors.New("no enough data for decode name")
		}
		nameSize, err := AMF0DecodeInt16(data[sizeUsed:])
		if err != nil {
			return ret, sizeUsed, nil
		}
		name, err := AMF0DecodeString(data[sizeUsed:])
		if err != nil {
			return ret, sizeUsed, nil
		}
		ret.Name = name
		sizeUsed += int32(2 + nameSize)
	}

	ret.PropType = int32(data[sizeUsed])
	sizeUsed += 1

	switch ret.PropType {
	case AMF0_number:
		err = binary.Read(bytes.NewReader(data[sizeUsed:]), binary.BigEndian, &ret.Value.NumValue)
		if err != nil {
			return ret, sizeUsed, err
		}
		sizeUsed += 8
	case AMF0_boolean:
		if data[sizeUsed] == 0 {
			ret.Value.BoolValue = false
		} else {
			ret.Value.BoolValue = true
		}
		sizeUsed += 1
	case AMF0_string:
		stringLength, err := AMF0DecodeInt16(data[sizeUsed:])
		if err != nil {
			return ret, sizeUsed, err
		}
		if stringLength == 0 {
			ret.Value.StrValue = ""
		} else {
			ret.Value.StrValue, err = AMF0DecodeString(data[sizeUsed:])
			if err != nil {
				return ret, sizeUsed, err
			}
		}
		sizeUsed += int32(2 + stringLength)
	case AMF0_object:
		tmpObj, size, err := amf0DecodeObj(data[sizeUsed:], true)
		if err != nil {
			return ret, sizeUsed, err
		}
		sizeUsed += size
		ret.Value.ObjValue = *tmpObj
	case AMF0_null:
	case AMF0_ecma_array:
		sizeUsed += 4
		tmpObj, size, err := amf0DecodeObj(data[sizeUsed:], true)
		if err != nil {
			return ret, sizeUsed, err
		}
		sizeUsed += size
		ret.Value.ObjValue = *tmpObj
	case AMF0_strict_array:
		size, err := amf0ReadStrictArray(data[sizeUsed:], ret)
		if err != nil {
			return ret, sizeUsed, err
		}
		sizeUsed += size
	case AMF0_date:
		ret.Value.NumValue, err = AMF0DecodeNumber(data[sizeUsed:])
		if err != nil {
			return ret, sizeUsed, err
		}
		sizeUsed += 8
		tmpu16, err := (AMF0DecodeInt16(data[sizeUsed:]))
		ret.Value.S16Value = int16(tmpu16)
		if err != nil {
			return ret, sizeUsed, err
		}
		sizeUsed += 2
	case AMF0_long_string:
		stringLength, err := AMF0DecodeInt32(data[sizeUsed:])
		if err != nil {
			return ret, sizeUsed, err
		}
		if stringLength == 0 {
			ret.Value.StrValue = ""
		} else {
			ret.Value.StrValue, err = AMFDecodeLongString(data[sizeUsed:])
			if err != nil {
				return ret, sizeUsed, err
			}
		}
		sizeUsed += int32(4 + stringLength)
	default:
		err = errors.New(fmt.Sprintf("not support amf type:%d", ret.PropType))
	}

	return ret, sizeUsed, err
}

func amf0ReadStrictArray(data []byte, prop *AMF0Property) (sizeUsed int32, err error) {
	if prop == nil {
		return 0, errors.New("invalid prop in amfReadStrictArray")
	}
	var arrayCount uint32
	sizeUsed = 0
	arrayCount, err = AMF0DecodeInt32(data)
	if err != nil {
		return 0, err
	}
	sizeUsed += 4
	if arrayCount == 0 {
		return sizeUsed, err
	}
	for arrayCount > 0 {
		arrayCount--
		tmpProp, size, err := amf0DecodeProp(data[sizeUsed:], false)
		if err != nil {
			return sizeUsed, err
		}
		prop.Value.ObjValue.Props.PushBack(tmpProp)
		sizeUsed += size
	}

	return sizeUsed, err
}

func (this *AMF0Object) AMF0GetPropByIndex(index int) (prop *AMF0Property) {
	if index >= this.Props.Len() {
		return nil
	}
	i := 0
	for e := this.Props.Front(); e != nil; e = e.Next() {
		if i == index {
			prop = e.Value.(*AMF0Property)
			return prop
		}
		i++
	}
	return
}

func (this *AMF0Object) AMF0GetPropByName(name string) (prop *AMF0Property) {
	for e := this.Props.Front(); e != nil; e = e.Next() {
		v := e.Value.(*AMF0Property)
		if v.Name == name {
			prop = v
			return
		}
	}
	return nil
}

func (this *AMF0Object) Dump() {
	for e := this.Props.Front(); e != nil; e = e.Next() {
		prop := e.Value.(*AMF0Property)
		this.dumpProp(prop)
	}
}

func (this *AMF0Object) dumpProp(prop *AMF0Property) {
	if len(prop.Name) > 0 {
		logger.LOGT(prop.Name)
	}
	switch prop.PropType {
	case AMF0_ecma_array:
		logger.LOGT("ecma array")
		this.dumpProp(prop)
	case AMF0_object:
		logger.LOGT("object")

		prop.Value.ObjValue.Dump()
	case AMF0_strict_array:
		logger.LOGT("static array")
		this.dumpProp(prop)
	case AMF0_string:
		logger.LOGT("string:" + prop.Value.StrValue)
	case AMF0_number:
		logger.LOGT(prop.Value.NumValue)
	case AMF0_boolean:
		logger.LOGT(prop.Value.BoolValue)
	}

}
