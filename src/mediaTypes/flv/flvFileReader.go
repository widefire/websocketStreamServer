package flv

import (
	"logger"
	"os"
)

type FlvFileReader struct {
	fp *os.File
}

func (this *FlvFileReader) Init(name string) error {
	var err error
	this.fp, err = os.Open(name)
	if err != nil {
		logger.LOGE(err.Error())
		return err
	}
	tmp := make([]byte, 13)
	_, err = this.fp.Read(tmp)
	return err
}

func (this *FlvFileReader) GetNextTag() (tag *FlvTag, err error) {
	buf := make([]byte, 11)
	_, err = this.fp.Read(buf)
	if err != nil {
		return
	}
	tag = &FlvTag{}
	tag.TagType = buf[0]
	dataSize := (int(int(buf[1])<<16) | (int(buf[2]) << 8) | (int(buf[3])))
	tag.Timestamp = uint32(int(int(buf[7])<<24) | (int(buf[4]) << 16) | (int(buf[5]) << 8) | (int(buf[6])))

	tag.Data = make([]byte, dataSize)
	_, err = this.fp.Read(tag.Data)
	if err != nil {
		return
	}
	buf = make([]byte, 4)
	this.fp.Read(buf)
	return
}

func (this *FlvFileReader) Close() {
	if this.fp != nil {
		this.fp.Close()
	}
}
