package rtsp

import (
	"log"
	"net/http"
	"testing"
)

func TestRTSPClient(t *testing.T) {
	/*
		urls:
		rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov
		rtsp://192.168.2.103/test
	*/
	client := NewClient()
	err := client.Dial("rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov")
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
	resp, err := client.ReadResponse()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(resp)
}
