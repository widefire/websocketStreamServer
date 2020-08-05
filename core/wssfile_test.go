package core

import (
	"testing"
)

func TestFileIO(t *testing.T) {
	wssfile := NewWSSFile()
	if wssfile == nil {
		t.Error()
	}
	err := wssfile.Open("d:/test.yuv")
	if err != nil {
		t.Error(err)
	}
	buf := make([]uint8, 100)
	count, err := wssfile.Read(buf)
	if count != 100 || err != nil {
		t.Error(err)
	}

	var val int16
	err = WSSIOReadL(wssfile, &val)
	if err != nil {
		t.Error(err)
	}

	err = WSSIOWriteL(wssfile, val)

	if err != nil {
		t.Error(err)
	}

	var f32val float32 = 222.123
	wssfile.Seek(0, 0)
	err = WSSIOWriteL(wssfile, f32val)
	if err != nil {
		t.Error(err)
	}

	f32val = 0
	wssfile.Seek(0, 0)
	err = WSSIOReadL(wssfile, &f32val)
	if err != nil || f32val != 222.123 {
		t.Error(err)
	}

	count, err = wssfile.Write(buf)
	if count != 100 || err != nil {
		t.Error(err)
	}

	err = wssfile.Seek(100, 1)
	if err != nil {
		t.Error(err)
	}

	err = wssfile.Close()
	if err != nil {
		t.Error(err)
	}

}
