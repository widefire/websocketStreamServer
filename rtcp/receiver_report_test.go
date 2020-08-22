package rtcp

import (
	"bytes"
	"testing"
)

func TestReceiverReportEncodeDecode(t *testing.T) {
	rr := &ReceiverReport{
		SSRC: 0x123,
	}
	rr.Reports = make([]*ReceptionReport, 2)
	rr.Reports[0] = &ReceptionReport{
		SSRC:                          1,
		FractionLost:                  2,
		CumulativeLost:                3,
		ExtendedHighestSequenceNumber: 4,
		Jitter:                        5,
		LSR:                           6,
		DLSR:                          7,
	}
	rr.Reports[1] = &ReceptionReport{
		SSRC:                          1111,
		FractionLost:                  2,
		CumulativeLost:                3333,
		ExtendedHighestSequenceNumber: 4444,
		Jitter:                        5555,
		LSR:                           6666,
		DLSR:                          7777,
	}
	extenCount := 7
	rr.ProfileSpecificExtensions = make([]byte, extenCount)
	for i := 0; i < len(rr.ProfileSpecificExtensions); i++ {
		rr.ProfileSpecificExtensions[i] = byte(i)
	}
	buf := new(bytes.Buffer)
	err := rr.Encode(buf)
	if err != nil {
		t.Error(err)
	}

	decSrc := &ReceiverReport{}
	if err = decSrc.Decode(buf.Bytes()); err != nil {
		t.Error(err)
	}

	decHeader := decSrc.Header()
	if decHeader.Padding == (extenCount%4 == 0) {
		t.Errorf("pad flag value error")
	}
	if len(decSrc.ProfileSpecificExtensions) != extenCount {
		t.Errorf("invalid extension")
	}

	if !rr.isEq(decSrc) {
		t.Fail()
	}
	t.Log("succeed")
}
