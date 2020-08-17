package rtsp

import "testing"

func TestRTSPClient(t *testing.T) {
	client := &Client{
		url: "rtsp://192.168.2.103/test.h264",
	}
	err := client.Dial()
	if err != nil {
		t.Fatal(err)
	}
}
