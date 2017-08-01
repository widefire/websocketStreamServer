package wssAPI

const (
	OBJ_ServerBus       = "ServerBus"
	OBJ_RTMPServer      = "RTMPServer"
	OBJ_WebSocketServer = "WebsocketServer"
	OBJ_BackendServer   = "BackendServer"
	OBJ_StreamerServer  = "StreamerServer"
	OBJ_RTSPServer      = "RTSPServer"
	OBJ_HLSServer       = "HLSServer"
	OBJ_DASHServer      = `DASHServer`
)

const (
	MSG_FLV_TAG            = "FLVTag"
	MSG_GetSource_NOTIFY   = "MSG.GetSource.Notify.Async"
	MSG_GetSource_Failed   = "MSG.GetSource.Failed"
	MSG_SourceClosed_Force = "MSG.SourceClosed.Force"
	MSG_PUBLISH_START      = "NetStream.Publish.Start"
	MSG_PUBLISH_STOP       = "NetStream.Publish.Stop"
	MSG_PLAY_START         = "NetStream.Play.Start"
	MSG_PLAY_STOP          = "NetStream.Play.Stop"
)
