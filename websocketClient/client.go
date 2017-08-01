package main

import (
	"encoding/json"
	"fmt"
	"logger"
	"net/http"
	"webSocketService"
	"wssAPI"

	"github.com/gorilla/websocket"
)

func main() {
	logger.SetFlags(logger.LOG_SHORT_FILE)
	cli := &websocket.Dialer{}
	req := http.Header{}

	conn, _, err := cli.Dial("ws://127.0.0.1:8080/ws/live", req)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	Play(conn)
	defer conn.Close()

}
func SetTest() {
	si := wssAPI.NewSet()
	si.Add("111")
	si.Add(2)
	fmt.Println(si.Has(3))
	fmt.Println(si.Has(2))
	fmt.Println(si.Has("111"))
}

type stPlay struct {
	Name  string `json:"name"`
	Start int    `json:"start"`
	Len   int    `json:"len"`
	Reset int    `json:"reset"`
	Req   int    `json:"req"`
}

func Play(conn *websocket.Conn) {

	//play
	stPlay := &stPlay{}

	stPlay.Name = "hks"

	dataSend, _ := json.Marshal(stPlay)
	err := webSocketService.SendWsControl(conn, webSocketService.WSC_play, dataSend)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	err = readResult(conn)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	err = readResult(conn)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	for {
		readResult(conn)
	}
	//end
	return
}

func readResult(conn *websocket.Conn) (err error) {
	msgType, data, err := conn.ReadMessage()
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	if msgType == websocket.BinaryMessage {
		logger.LOGT(data)
		pktType := data[0]
		switch pktType {
		case webSocketService.WS_pkt_audio:
			logger.LOGT("audio")
		case webSocketService.WS_pkt_video:
			logger.LOGT("video")
		case webSocketService.WS_pkt_control:
			logger.LOGT(data)
		}
	}
	return
}
