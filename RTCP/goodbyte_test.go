package rtcp

import (
	"bytes"
	"testing"
)

func TestGoodbyeEncodeDecode(t *testing.T) {
	src := &GoodByte{}
	src.SSCSRC = make([]uint32, 3)
	src.SSCSRC[0] = 3
	src.SSCSRC[1] = 6
	src.SSCSRC[2] = 9
	src.ReasonForLeaving = "I just want to leave"

	buf := new(bytes.Buffer)
	err := src.Encode(buf)
	if err != nil {
		t.Fatal(err)
	}
	dst := &GoodByte{}
	err = dst.Decode(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !src.isEq(dst) {
		t.Fatal("not eq")
	}
}
