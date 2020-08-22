package rtcp

import (
	"bytes"
	"testing"
)

func TestSourceDescriptionEncodeDecode(t *testing.T) {
	src := &SDES{}
	src.Chunks = make([]*SDESChunk, 0)

	chunk0 := &SDESChunk{}
	chunk0.SSCSRC = 12340
	chunk0.Items = make([]*SDESItem, 0)
	src.Chunks = append(src.Chunks, chunk0)

	chunk1 := &SDESChunk{}
	chunk1.SSCSRC = 12341
	chunk1.Items = make([]*SDESItem, 2)
	chunk1.Items[0] = &SDESItem{}
	chunk1.Items[0].Type = SDESName
	chunk1.Items[0].Text = "I'm SDESName"
	chunk1.Items[1] = &SDESItem{}
	chunk1.Items[1].Type = SDESPhone
	chunk1.Items[1].Text = "1511511151"
	src.Chunks = append(src.Chunks, chunk1)

	buf := new(bytes.Buffer)
	err := src.Encode(buf)
	if err != nil {
		t.Error(err)
	}
	dst := &SDES{}
	err = dst.Decode(buf.Bytes())
	if err != nil {
		t.Error(err)
	}
	if !src.isEq(dst) {
		t.Error("not eq")
	}
	srcHeader := src.Header()
	dstHeader := dst.Header()
	if !srcHeader.isEq(dstHeader) {
		t.Error("header not eq")
	}
}
