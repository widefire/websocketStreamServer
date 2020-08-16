package rtcp

import (
	"bytes"
	"testing"
)

func TestAppEncodeDecode(t *testing.T) {
	src := &ApplicationDefined{}
	src.SSCSRC = 13456
	src.SubType = 20
	src.Name = make([]byte, 4)
	src.ApplicationDependentData = make([]byte, 223)
	for i := 0; i < 4; i++ {
		src.Name[i] = byte(i)
	}
	for i := 0; i < len(src.ApplicationDependentData); i++ {
		src.ApplicationDependentData[i] = byte(i)
	}

	buf := new(bytes.Buffer)
	err := src.Encode(buf)
	if err != nil {
		t.Fatal(err)
	}

	dst := &ApplicationDefined{}
	err = dst.Decode(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !src.isEq(dst) {
		t.Error("not eq")
	}
}
