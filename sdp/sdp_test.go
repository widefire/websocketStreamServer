package sdp

import (
	"log"
	"os"
	"testing"
)

func TestParseSDP(t *testing.T) {
	fp, err := os.Open("org.dat")
	if err != nil {
		t.Error(err)
		return
	}
	defer fp.Close()
	fileInfo, err := fp.Stat()
	if err != nil {
		t.Error(err)
		return
	}
	sdpBuf := make([]byte, fileInfo.Size())
	readCount, err := fp.Read(sdpBuf)
	if err != nil {
		t.Error(err)
		return
	}
	if int64(readCount) != fileInfo.Size() {
		t.Error("read sdp file failed")
		return
	}
	sdp, err := ParseSdp(string(sdpBuf))
	if err != nil {
		t.Error(err)
		return
	}
	if sdp != nil {
		log.Println(sdp.ProtocolVersion)
	}
}
