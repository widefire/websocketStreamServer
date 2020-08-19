package rtsp

import (
	"net/http"
	"testing"
)

func TestRTSPClient(t *testing.T) {
	/*
		urls:
		rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov
		rtsp://192.168.2.103/111.aac
		rtsp://192.168.2.103/lyf.264
	*/
	client := NewClient()
	err := client.Dial("rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov")
	//err := client.Dial("rtsp://192.168.2.103/111.aac")
	if err != nil {
		t.Fatal(err)
	}

	optionHeader := http.Header{}
	optionHeader.Add("Require", "implicit-play")
	optionHeader.Add("Proxy-Require", "gzipped-messages")
	err = client.SendOptions(optionHeader)
	if err != nil {
		t.Fatal(err)
	}
	client.AddCSeq()
	err = client.SendDescribe(nil)
	if err != nil {
		t.Fatal(err)
	}
	client.AddCSeq()
	err = client.SendSetup(nil, true)
	if err != nil {
		t.Fatal(err)
	}
}
