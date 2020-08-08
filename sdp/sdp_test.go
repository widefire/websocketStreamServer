package sdp

import (
	"errors"
	"log"
	"os"
	"testing"
)

func TestParseSDP(t *testing.T) {
	log.SetOutput(os.Stdout)
	err := LoadAndTestSDPFromFile("sdp1.data")
	if err != nil {
		t.Error(err)
	}
	err = LoadAndTestSDPFromFile("sdp2.data")
	if err != nil {
		t.Error(err)
	}
}

func LoadAndTestSDPFromFile(filename string) error {
	fp, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	fileInfo, err := fp.Stat()
	if err != nil {
		return err
	}
	sdpBuf := make([]byte, fileInfo.Size())
	readCount, err := fp.Read(sdpBuf)
	if err != nil {
		return err
	}
	if int64(readCount) != fileInfo.Size() {
		return errors.New("read sdp file failed")
	}
	sdp, err := ParseSDP(string(sdpBuf))
	if err != nil {
		log.Println(err)
		return err
	}
	if sdp != nil {
		log.Println(*sdp.ProtocolVersion)
	}
	return nil
}
