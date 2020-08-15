package rtcp

import (
	"bytes"
	"testing"
)

func TestReceptionEncodeDecode(t *testing.T) {
	src := &ReceptionReport{
		SSRC:                          12345,
		FractionLost:                  32,
		CumulativeLost:                33223,
		ExtendedHighestSequenceNumber: 44444,
		Jitter:                        77777,
		LSR:                           11111,
		DLSR:                          22222,
	}
	buf := new(bytes.Buffer)
	err := src.Encode(buf)
	if err != nil {
		t.Error(err)
	}
	dst := &ReceptionReport{}
	err = dst.Decode(buf.Bytes())
	if err != nil {
		t.Error(err)
	}

	if !src.isEq(dst) {
		t.Error("not eq")
	}

}
