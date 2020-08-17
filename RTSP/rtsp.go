package rtsp

import (
	"errors"
	"log"
	"strings"
)

//method
const (
	//                                           direction       	object  requirement
	MethodDESCRIBE     = "DESCRIBE"      //C->S 				P,S		recommended
	MethodANNOUNCE     = "ANNOUNCE"      //C->S, S->C		P,S		optional
	MethodGetPARAMETER = "GET_PARAMETER" //C->S, S->C		P,S		optional
	MethodOPTIONS      = "OPTIONS"       //C->S, S->C		P,S		required (S->C: optional)
	MethodPAUSE        = "PAUSE"         //C->S				P,S		recommended
	MethodPLAY         = "PLAY"          //C->S				P,S		required
	MethodRECORD       = "RECORD"        //C->S				P,S		optional
	MethodREDIRECT     = "REDIRECT"      //S->C				P,S		optional
	MethodSETUP        = "SETUP"         //C->S				S		required
	MethodSetPARAMETER = "PARAMETER"     //C->S, S->C		P,S		optional
	MethodSetTEARDOWN  = "TEARDOWN"      //C->S				P,S		required

)

//status code
const (
	StatusContinue                       int = 100
	StatusOK                             int = 200
	StatusCreated                        int = 201
	StatusLowOnStorageSpace              int = 250
	StatusMultipleChoices                int = 300
	StatusMovedPermanently               int = 301
	StatusMovedTemporarily               int = 302
	StatusSeeOther                       int = 303
	StatusNotModified                    int = 304
	StatusUseProxy                       int = 305
	StatusBadRequest                     int = 400
	StatusUnauthorized                   int = 401
	StatusPaymentRequired                int = 402
	StatusForbidden                      int = 403
	StatusNotFound                       int = 404
	StatusMethodNotAllowed               int = 405
	StatusNotAcceptable                  int = 406
	StatusProxyAuthenticationRequired    int = 407
	StatusRequestTimeout                 int = 408
	StatusGone                           int = 410
	StatusLengthRequired                 int = 411
	StatusPreconditionFaile              int = 412
	StatusRequestEntityTooLarge          int = 413
	StatusRequestURITooLarge             int = 414
	StatusUnsupportedMediaType           int = 415
	StatusParameterNotUnderstood         int = 451
	StatusConferenceNotFound             int = 452
	StatusNotEnoughBandwidth             int = 453
	StatusSessionNotFound                int = 454
	StatusMethodNotValidInThisState      int = 455
	StatusHeaderFieldNotValidForResource int = 456
	StatusInvalidRange                   int = 457
	StatusParameterIsReadOnly            int = 458
	StatusAggregateOperationNotAllowed   int = 459
	StatusOnlyAggregateOperationAllowed  int = 460
	StatusUnsupportedTransport           int = 461
	StatusDestinationUnreachable         int = 462
	StatusInternalServerError            int = 500
	StatusNotImplemented                 int = 501
	StatusBadGateway                     int = 502
	StatusServiceUnavailable             int = 503
	StatusGatewayTimeout                 int = 504
	StatusRTSPVersionNotSupported        int = 505
	StatusOptionNotSupported             int = 551
)

var statusCodeDescMap map[int]string

func init() {
	statusCodeDescMap = make(map[int]string)
	statusCodeDescMap[StatusContinue] = "Continue"
	statusCodeDescMap[StatusOK] = "OK"
	statusCodeDescMap[StatusCreated] = "Created"
	statusCodeDescMap[StatusLowOnStorageSpace] = "Low on Storage Space"
	statusCodeDescMap[StatusMultipleChoices] = "Multiple Choices"
	statusCodeDescMap[StatusMovedPermanently] = "Moved Permanently"
	statusCodeDescMap[StatusMovedTemporarily] = "Moved Temporarily"
	statusCodeDescMap[StatusSeeOther] = "See Other"
	statusCodeDescMap[StatusNotModified] = "Not Modified"
	statusCodeDescMap[StatusUseProxy] = "Use Proxy"
	statusCodeDescMap[StatusBadRequest] = "Bad Request"
	statusCodeDescMap[StatusUnauthorized] = "Unauthorized"
	statusCodeDescMap[StatusPaymentRequired] = "Payment Required"
	statusCodeDescMap[StatusForbidden] = "Forbidden"
	statusCodeDescMap[StatusNotFound] = "Not Found"
	statusCodeDescMap[StatusMethodNotAllowed] = "Method Not Allowed"
	statusCodeDescMap[StatusNotAcceptable] = "Not Acceptable"
	statusCodeDescMap[StatusProxyAuthenticationRequired] = "Proxy Authentication Required"
	statusCodeDescMap[StatusRequestTimeout] = "Request Timeout"
	statusCodeDescMap[StatusGone] = "Gone"
	statusCodeDescMap[StatusLengthRequired] = "Length Required"
	statusCodeDescMap[StatusPreconditionFaile] = "Precondition Failed"
	statusCodeDescMap[StatusRequestEntityTooLarge] = "Request Entity Too Large"
	statusCodeDescMap[StatusRequestURITooLarge] = "Request-URI Too Long"
	statusCodeDescMap[StatusUnsupportedMediaType] = "Unsupported Media Type"
	statusCodeDescMap[StatusParameterNotUnderstood] = "Invalid parameter"
	statusCodeDescMap[StatusConferenceNotFound] = "Illegal Conference Identifier"
	statusCodeDescMap[StatusNotEnoughBandwidth] = "Not Enough Bandwidth"
	statusCodeDescMap[StatusSessionNotFound] = "Session Not Found"
	statusCodeDescMap[StatusMethodNotValidInThisState] = "Method Not Valid In This State"
	statusCodeDescMap[StatusHeaderFieldNotValidForResource] = "Header Field Not Valid"
	statusCodeDescMap[StatusInvalidRange] = "Invalid Range"
	statusCodeDescMap[StatusParameterIsReadOnly] = "Parameter Is Read-Only"
	statusCodeDescMap[StatusAggregateOperationNotAllowed] = "Aggregate Operation Not Allowed"
	statusCodeDescMap[StatusOnlyAggregateOperationAllowed] = "Only Aggregate Operation Allowed"
	statusCodeDescMap[StatusUnsupportedTransport] = "Unsupported Transport"
	statusCodeDescMap[StatusDestinationUnreachable] = "Destination Unreachable"
	statusCodeDescMap[StatusInternalServerError] = "Internal Server Error"
	statusCodeDescMap[StatusNotImplemented] = "Not Implemented"
	statusCodeDescMap[StatusBadGateway] = "Bad Gateway"
	statusCodeDescMap[StatusServiceUnavailable] = "Service Unavailable"
	statusCodeDescMap[StatusGatewayTimeout] = "Gateway Timeout"
	statusCodeDescMap[StatusRTSPVersionNotSupported] = "RTSP Version Not Supported"
	statusCodeDescMap[StatusOptionNotSupported] = "Option not support"
}

// GetRTSPStatusDesc ...
// param statusCode : RTSP status code
// return desc :RTSP status desc
// return ok :true for ok,false for not default status
func GetRTSPStatusDesc(statusCode int) (desc string, ok bool) {
	desc, ok = statusCodeDescMap[statusCode]
	return
}

func parseURL(url string) (addr string, path string, err error) {
	if !strings.HasPrefix(url, "rtsp://") {
		err = errors.New("a rtsp url must start with rtsp://")
		log.Println(err)
		return
	}
	urlpayload := strings.TrimPrefix(url, "rtsp://")
	log.Println(urlpayload)
	return
}
