package webSocketService

import (
	"encoding/json"
	"logger"
	"wssAPI"

	"github.com/gorilla/websocket"
)

const (
	WS_status_status = "status"
	WS_status_error  = "error"
)

//1byte type
const (
	WS_pkt_audio   = 8
	WS_pkt_video   = 9
	WS_pkt_control = 18
)

const (
	WSC_play       = 1
	WSC_play2      = 2
	WSC_resume     = 3
	WSC_pause      = 4
	WSC_seek       = 5
	WSC_close      = 7
	WSC_stop       = 6
	WSC_publish    = 0x10
	WSC_onMetaData = 9
)

var cmdsMap map[int]*wssAPI.Set

func init() {
	cmdsMap = make(map[int]*wssAPI.Set)
	//初始状态close，可以play,close,publish
	{
		tmp := wssAPI.NewSet()
		tmp.Add(WSC_play)
		tmp.Add(WSC_play2)
		tmp.Add(WSC_close)
		tmp.Add(WSC_publish)
		cmdsMap[WSC_close] = tmp
	}
	//play 可以close pause seek
	{
		tmp := wssAPI.NewSet()
		tmp.Add(WSC_pause)
		tmp.Add(WSC_play)
		tmp.Add(WSC_play2)
		tmp.Add(WSC_seek)
		tmp.Add(WSC_close)
		cmdsMap[WSC_play] = tmp
	}
	//play2 =play
	{
		tmp := wssAPI.NewSet()
		tmp.Add(WSC_pause)
		tmp.Add(WSC_play)
		tmp.Add(WSC_play2)
		tmp.Add(WSC_seek)
		tmp.Add(WSC_close)
		cmdsMap[WSC_play2] = tmp
	}
	//pause
	{
		tmp := wssAPI.NewSet()
		tmp.Add(WSC_resume)
		tmp.Add(WSC_play)
		tmp.Add(WSC_play2)
		tmp.Add(WSC_close)
		cmdsMap[WSC_pause] = tmp
	}
	//publish
	{
		tmp := wssAPI.NewSet()
		tmp.Add(WSC_close)
		cmdsMap[WSC_publish] = tmp
	}
}

func supportNewCmd(cmdOld, cmdNew int) bool {
	_, exist := cmdsMap[cmdOld]
	if false == exist {
		return false
	}
	return cmdsMap[cmdOld].Has(cmdNew)
}

func SendWsControl(conn *websocket.Conn, ctrlType int, data []byte) (err error) {
	dataSend := make([]byte, len(data)+4)
	dataSend[0] = WS_pkt_control
	dataSend[1] = byte((ctrlType >> 16) & 0xff)
	dataSend[2] = byte((ctrlType >> 8) & 0xff)
	dataSend[3] = byte((ctrlType >> 0) & 0xff)
	copy(dataSend[4:], data)
	return conn.WriteMessage(websocket.BinaryMessage, dataSend)
}

func SendWsStatus(conn *websocket.Conn, level, code string, req int) (err error) {
	st := &stResult{Level: level, Code: code, Req: req}
	dataJson, err := json.Marshal(st)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	dataSend := make([]byte, len(dataJson)+4)
	dataSend[0] = WS_pkt_control
	dataSend[1] = 0
	dataSend[2] = 0
	dataSend[3] = 0
	copy(dataSend[4:], dataJson)
	err = conn.WriteMessage(websocket.BinaryMessage, dataSend)
	return
}

type stPlay struct {
	Name  string `json:"name"`
	Start int    `json:"start"`
	Len   int    `json:"len"`
	Reset int    `json:"reset"`
	Req   int    `json:"req"`
}

type stPlay2 struct {
	Name  string `json:"name"`
	Start int    `json:"start"`
	Len   int    `json:"len"`
	Reset int    `json:"reset"`
	Req   int    `json:"req"`
}

type stResume struct {
	Req int `json:"req"`
}

type stPause struct {
	Req int `json:"req"`
}

type stSeek struct {
	Offset int `json:"offset"`
	Req    int `json:"req"`
}

type stClose struct {
	Req int `json:"req"`
}

type stStop struct {
	Req int `json:"req"`
}

type stPublish struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Req  int    `json:"req"`
}

type stResult struct {
	Level string `json:"level"`
	Code  string `json:"code"`
	Req   int    `json:"req"`
}

const (
	NETCONNECTION_CALL_FAILED         = "NetConnection.Call.Failed"
	NETCONNECTION_CONNECT_APPSHUTDOWN = "NetConnection.Connect.AppShutdown"
	NETCONNECTION_CONNECT_CLOSED      = "NetConnection.Connect.Closed"
	NETCONNECTION_CONNECT_FAILED      = "NetConnection.Connect.Failed"
	NETCONNECTION_CONNECT_IDLETIMEOUT = "NetConnection.Connect.IdleTimeout"
	NETCONNECTION_CONNECT_INVALIDAPP  = "NetConnection.Connect.InvalidApp"
	NETCONNECTION_CONNECT_REJECTED    = "NetConnection.Connect.Rejected"
	NETCONNECTION_CONNECT_SUCCESS     = "NetConnection.Connect.Success"

	NETSTREAM_BUFFER_EMPTY              = "NetStream.Buffer.Empty"
	NETSTREAM_BUFFER_FLUSH              = "NetStream.Buffer.Flush"
	NETSTREAM_BUFFER_FULL               = "NetStream.Buffer.Full"
	NETSTREAM_FAILED                    = "NetStream.Failed"
	NETSTREAM_PAUSE_NOTIFY              = "NetStream.Pause.Notify"
	NETSTREAM_PLAY_FAILED               = "NetStream.Play.Failed"
	NETSTREAM_PLAY_FILESTRUCTUREINVALID = "NetStream.Play.FileStructureInvalid"
	NETSTREAM_PLAY_PUBLISHNOTIFY        = "NetStream.Play.PublishNotify"
	NETSTREAM_PLAY_RESET                = "NetStream.Play.Reset"
	NETSTREAM_PLAY_START                = "NetStream.Play.Start"
	NETSTREAM_PLAY_STOP                 = "NetStream.Play.Stop"
	NETSTREAM_PLAY_STREAMNOTFOUND       = "NetStream.Play.StreamNotFound"
	NETSTREAM_PLAY_UNPUBLISHNOTIFY      = "NetStream.Play.UnpublishNotify"
	NETSTREAM_PUBLISH_BADNAME           = "NetStream.Publish.BadName"
	NETSTREAM_PUBLISH_IDLE              = "NetStream.Publish.Idle"
	NETSTREAM_PUBLISH_START             = "NetStream.Publish.Start"
	NETSTREAM_RECORD_ALREADYEXISTS      = "NetStream.Record.AlreadyExists"
	NETSTREAM_RECORD_FAILED             = "NetStream.Record.Failed"
	NETSTREAM_RECORD_NOACCESS           = "NetStream.Record.NoAccess"
	NETSTREAM_RECORD_START              = "NetStream.Record.Start"
	NETSTREAM_RECORD_STOP               = "NetStream.Record.Stop"
	NETSTREAM_SEEK_FAILED               = "NetStream.Seek.Failed"
	NETSTREAM_SEEK_INVALIDTIME          = "NetStream.Seek.InvalidTime"
	NETSTREAM_SEEK_NOTIFY               = "NetStream.Seek.Notify"
	NETSTREAM_STEP_NOTIFY               = "NetStream.Step.Notify"
	NETSTREAM_UNPAUSE_NOTIFY            = "NetStream.Unpause.Notify"
	NETSTREAM_UNPUBLISH_SUCCESS         = "NetStream.Unpublish.Success"
	NETSTREAM_VIDEO_DIMENSIONCHANGE     = "NetStream.Video.DimensionChange"
)
