package rtcp

import (
	"bytes"
	"testing"
)

func TestSenderReportEncodeDecode(t *testing.T) {
	sr := &SenderReport{
		SSRC:               0x123,
		NTPTime:            0x234,
		RTPTime:            0x345,
		SendersPacketCount: 0x456,
		SendersOctetCount:  0x567,
	}
	sr.Reports = make([]*ReceptionReport, 2)
	sr.Reports[0] = &ReceptionReport{
		SSRC:                          1,
		FractionLost:                  2,
		CumulativeLost:                3,
		ExtendedHighestSequenceNumber: 4,
		Jitter:                        5,
		LSR:                           6,
		DLSR:                          7,
	}
	sr.Reports[1] = &ReceptionReport{
		SSRC:                          1111,
		FractionLost:                  2,
		CumulativeLost:                3333,
		ExtendedHighestSequenceNumber: 4444,
		Jitter:                        5555,
		LSR:                           6666,
		DLSR:                          7777,
	}
	padCount := 6
	sr.ProfileSpecificExtensions = make([]byte, padCount)
	for i := 0; i < len(sr.ProfileSpecificExtensions); i++ {
		sr.ProfileSpecificExtensions[i] = byte(i)
	}
	buf := new(bytes.Buffer)
	err := sr.Encode(buf)
	if err != nil {
		t.Error(err)
	}

	decSrc := &SenderReport{}
	if err = decSrc.Decode(buf.Bytes()); err != nil {
		t.Error(err)
	}

	decHeader := decSrc.Header()
	if decHeader.Padding == (padCount%4 == 0) {
		t.Errorf("pad flag error")
	}
	if len(decSrc.ProfileSpecificExtensions) != padCount {
		t.Errorf("invalid extension")
	}

	if !sr.isEq(decSrc) {
		t.Fail()
	}
	t.Log("succeed")
}
