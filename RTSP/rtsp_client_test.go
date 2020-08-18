package rtsp

import (
	"net/http"
	"testing"
)

func TestRTSPClient(t *testing.T) {
	/*
		urls:
		rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov
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
}
