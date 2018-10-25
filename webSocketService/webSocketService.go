package webSocketService

import (
	"encoding/json"
	"errors"
	"fmt"
	"logger"
	"net/http"
	"strconv"
	"strings"
	"wssAPI"

	"github.com/gorilla/websocket"
	"HTTPMUX"
)

type WebSocketService struct {
	parent wssAPI.Obj
}

type WebSocketConfig struct {
	Port int `json:"Port"`
	Route string `json:"Route"`
}

var service *WebSocketService
var serviceConfig WebSocketConfig

func (this *WebSocketService) Init(msg *wssAPI.Msg) (err error) {

	if msg == nil || msg.Param1 == nil {
		logger.LOGE("init Websocket service failed")
		return errors.New("invalid param")
	}

	fileName := msg.Param1.(string)
	err = this.loadConfigFile(fileName)
	if err != nil {
		logger.LOGE(err.Error())
		return errors.New("load websocket config failed")
	}
	service = this
	strPort := ":" + strconv.Itoa(serviceConfig.Port)
	HTTPMUX.AddRoute(strPort,serviceConfig.Route,this.ServeHTTP)
	//go func() {
	//	if true {
	//		strPort := ":" + strconv.Itoa(serviceConfig.Port)
	//		//http.Handle("/", this)
	//		mux := http.NewServeMux()
	//		mux.Handle("/", this)
	//		err = http.ListenAndServe(strPort, mux)
	//		if err != nil {
	//			logger.LOGE("start websocket failed:" + err.Error())
	//		}
	//	}
	//}()
	return
}

func (this *WebSocketService) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *WebSocketService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *WebSocketService) GetType() string {
	return wssAPI.OBJ_WebSocketServer
}

func (this *WebSocketService) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *WebSocketService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *WebSocketService) loadConfigFile(fileName string) (err error) {
	data, err := wssAPI.ReadFileAll(fileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &serviceConfig)
	if err != nil {
		return
	}

	return
}

func (this *WebSocketService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	path=strings.TrimPrefix(path,serviceConfig.Route)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	//logger.LOGT(path)
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	logger.LOGT(fmt.Sprintf("new websocket connect %s", conn.RemoteAddr().String()))
	this.handleConn(conn, req, path)
	defer func() {
		conn.Close()
		logger.LOGD("close websocket conn")
	}()
}

func (this *WebSocketService) handleConn(conn *websocket.Conn, req *http.Request, path string) {
	handler := &websocketHandler{}
	msg := &wssAPI.Msg{}
	msg.Param1 = conn
	msg.Param2 = path

	handler.Init(msg)
	defer func() {
		handler.processWSMessage(nil)
	}()
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		switch messageType {
		case websocket.TextMessage:
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				logger.LOGI("send text msg failed:" + err.Error())
				return
			}
		case websocket.BinaryMessage:
			err = handler.processWSMessage(data)
			if err != nil {
				logger.LOGE(err.Error())
				logger.LOGE("ws binary error")
				return
			}
		case websocket.CloseMessage:
			err = errors.New("websocket closed:" + conn.RemoteAddr().String())
			return
		case websocket.PingMessage:
			//conn.WriteMessage()
			conn.WriteMessage(websocket.PongMessage, []byte(" "))
		case websocket.PongMessage:
		default:
		}
	}
}

func (this *WebSocketService) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}
