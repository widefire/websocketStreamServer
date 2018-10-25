package RTSPService

import (
	"logger"
	"net"
	"strconv"
)

const (
	RTSP_VER             = "RTSP/1.0"
	RTSP_EL              = "\r\n"
	RTSP_RTP_AVP         = "RTP/AVP"
	RTSP_RTP_AVP_TCP     = "RTP/AVP/TCP"
	RTSP_RTP_AVP_UDP     = "RTP/AVP/UDP"
	RTSP_RAW_UDP         = "RTP/RAW/UDP"
	RTSP_CONTROL_ID_0    = "track1"
	RTSP_CONTROL_ID_1    = "track2"
	HDR_CONTENTLENGTH    = "Content-Length"
	HDR_ACCEPT           = "Accept"
	HDR_ALLOW            = "Allow"
	HDR_BLOCKSIZE        = "Blocksize"
	HDR_CONTENTTYPE      = "Content-Type"
	HDR_DATE             = "Date"
	HDR_REQUIRE          = "Require"
	HDR_TRANSPORTREQUIRE = "Transport-Require"
	HDR_SEQUENCENO       = "SequenceNo"
	HDR_CSEQ             = "CSeq"
	HDR_STREAM           = "Stream"
	HDR_SESSION          = "Session"
	HDR_TRANSPORT        = "Transport"
	HDR_RANGE            = "Range"
	HDR_USER_AGENT       = "User-Agent"

	RTSP_METHOD_MAXLEN         = 15
	RTSP_METHOD_DESCRIBE       = "DESCRIBE"
	RTSP_METHOD_ANNOUNCE       = "ANNOUNCE"
	RTSP_METHOD_GET_PARAMETERS = "GET_PARAMETERS"
	RTSP_METHOD_OPTIONS        = "OPTIONS"
	RTSP_METHOD_PAUSE          = "PAUSE"
	RTSP_METHOD_PLAY           = "PLAY"
	RTSP_METHOD_RECORD         = "RECORD"
	RTSP_METHOD_REDIRECT       = "REDIRECT"
	RTSP_METHOD_SETUP          = "SETUP"
	RTSP_METHOD_SET_PARAMETER  = "SET_PARAMETER"
	RTSP_METHOD_TEARDOWN       = "TEARDOWN"

	DEFAULT_MTU_2 = 0xfff

	RTSP_payload_h264  = 96
	RTP_video_freq     = 90000
	RTP_audio_freq     = 8000
	SOCKET_PACKET_SIZE = 1456

	RTSPServerName   = "StreamServer_RTSP_alpha"
	ctrl_track_audio = "track_audio"
	ctrl_track_video = "track_video"
)

func getRTSPStatusByCode(code int) (status string) {
	switch code {
	case 100:
		status = "Continue"
	case 200:
		status = "OK"
	case 201:
		status = "Created"
	case 202:
		status = "Accepted"
	case 203:
		status = "Non-Authoritative Information"
	case 204:
		status = "No Content"
	case 205:
		status = "Reset Content"
	case 206:
		status = "Partial Content"
	case 300:
		status = "Multiple Choices"
	case 301:
		status = "Moved Permanently"
	case 302:
		status = "Moved Temporarily"
	case 400:
		status = "Bad Request"
	case 401:
		status = "Unauthorized"
	case 402:
		status = "Payment Required"
	case 403:
		status = "Forbidden"
	case 404:
		status = "Not Found"
	case 405:
		status = "Method Not Allowed"
	case 406:
		status = "Not Acceptable"
	case 407:
		status = "Proxy Authentication Required"
	case 408:
		status = "Request Time-out"
	case 409:
		status = "Conflict"
	case 410:
		status = "Gone"
	case 411:
		status = "Length Required"
	case 412:
		status = "Precondition Failed"
	case 413:
		status = "Request Entity Too Large"
	case 414:
		status = "Request-URI Too Large"
	case 415:
		status = "Unsupported Media Type"
	case 420:
		status = "Bad Extension"
	case 450:
		status = "Invalid Parameter"
	case 451:
		status = "Parameter Not Understood"
	case 452:
		status = "Conference Not Found"
	case 453:
		status = "Not Enough Bandwidth"
	case 454:
		status = "Session Not Found"
	case 455:
		status = "Method Not Valid In This State"
	case 456:
		status = "Header Field Not Valid for Resource"
	case 457:
		status = "Invalid Range"
	case 458:
		status = "Parameter Is Read-Only"
	case 461:
		status = "Unsupported transport"
	case 500:
		status = "Internal Server Error"
	case 501:
		status = "Not Implemented"
	case 502:
		status = "Bad Gateway"
	case 503:
		status = "Service Unavailable"
	case 504:
		status = "Gateway Time-out"
	case 505:
		status = "RTSP Version Not Supported"
	case 551:
		status = "Option not supported"
	case 911:
		status = "Extended Error:"
	}
	return
}

func sendRTSPErrorReply(code int, cseq int, conn net.Conn) error {
	strOut := RTSP_VER + " " + strconv.Itoa(code) + " " + getRTSPStatusByCode(code) + RTSP_EL
	strOut += "CSeq: " + strconv.Itoa(cseq) + RTSP_EL
	strOut += RTSP_EL
	_, err := conn.Write([]byte(strOut))
	if err != nil {
		logger.LOGE(err.Error())
	}
	return err
}
