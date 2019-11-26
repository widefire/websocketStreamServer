package core

import(
	"testing"
)

func TestFileIO(t *testing.T){
	wssfile := NewWSSFile()
	if wssfile == nil {
		t.Error()
	}
	err := wssfile.Open("d:/test.yuv")
	if	err != nil{
		t.Error(err)
	}
	buf := make([]uint8,100)
	count, err := wssfile.Read(buf,100)
	if count != 100 || err != nil{
		t.Error(err)
	}

	count, err = wssfile.Write(buf)
	if count != 100 || err != nil{
		t.Error(err)
	}

	err = wssfile.Seek(100)
	if err != nil{
		t.Error(err)
	}

	err = wssfile.Close()
	if err != nil{
		t.Error(err)
	}

}