package rtsp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//DefaultPort RTSP default port
const DefaultPort = 554

//RTSPVeraion1 ...
const RTSPVeraion1 = "RTSP/1.0"

//RTSPEndl ...
const RTSPEndl = "\r\n"

//RTSPCSeq ...
const RTSPCSeq = "CSeq"

//MaxRtspBodyLen ...
const MaxRtspBodyLen = 1024 * 1024 * 10

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

func parseURL(rawURL string) (u *url.URL, err error) {
	u, err = url.Parse(rawURL)
	if err != nil {
		log.Println(err)
		return
	}
	if u.Scheme != "rtsp" {
		err = fmt.Errorf("invalid rtsp url:%s", u.Scheme)
		log.Println(err)
		return
	}

	port := u.Port()
	if len(port) == 0 {
		u.Host = fmt.Sprintf("%s:%d", u.Host, DefaultPort)
	}
	return
}

//Request RTSP request
//https://github.com/beatgammit/rtsp/blob/master/rtsp.go
type Request struct {
	Method string
	URL    *url.URL
	Header http.Header
	Body   []byte
}

//NewRequest ...
func NewRequest(method string, urlIn *url.URL, header http.Header, body []byte) *Request {
	return &Request{
		Method: method,
		URL:    urlIn,
		Header: header,
		Body:   body,
	}
}

//String Request to string
func (r Request) String() string {
	ret := fmt.Sprintf("%s %s %s%s", MethodOPTIONS, r.URL.String(), RTSPVeraion1, RTSPEndl)
	if len(r.Body) > 0 {
		r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))
	}
	for k, v := range r.Header {
		for _, v := range v {
			ret += fmt.Sprintf("%s: %s%s", k, v, RTSPEndl)
		}
	}

	ret += RTSPEndl
	if len(r.Body) > 0 {

		ret += string(r.Body)
	}
	return ret
}

//Response RTSP response
type Response struct {
	Version    string
	StatusCode int
	StatusDesc string
	CSeq       int
	Header     http.Header
	Body       []byte
}

//String Response to string
func (r Response) String() string {
	ret := fmt.Sprintf("%s %d %s%s", r.Version, r.StatusCode, r.StatusDesc, RTSPEndl)
	if len(r.Body) > 0 {
		r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))
	}
	for k, v := range r.Header {
		for _, sv := range v {
			ret += fmt.Sprintf("%s: %s%s", k, sv, RTSPEndl)
		}
	}
	ret += RTSPEndl
	if len(r.Body) > 0 {
		ret += string(r.Body)
	}
	return ret
}

//ReadResponse ...
func ReadResponse(r io.Reader) (response *Response, err error) {
	response = &Response{
		Header: http.Header{},
		Body:   make([]byte, 0),
	}
	reader := bufio.NewReader(r)
	line, err := readline(reader)
	if err != nil {
		log.Println(err)
		return
	}
	subStr := strings.SplitN(line, " ", 3)
	response.Version = subStr[0]
	if response.Version != RTSPVeraion1 {
		err = fmt.Errorf("invalid rtsp version %s", response.Version)
		log.Println(err)
		return
	}
	response.StatusCode, err = strconv.Atoi(subStr[1])
	if err != nil {
		log.Println(err)
		return
	}
	response.StatusDesc = strings.TrimSpace(subStr[2])
	for {
		line, err = readline(reader)
		if err != nil {
			log.Println(err)
			return
		}
		if len(line) == 0 {
			break
		}
		subStr = strings.SplitN(line, ":", 2)
		response.Header.Add(strings.TrimSpace(subStr[0]), strings.TrimSpace(subStr[1]))
	}

	response.CSeq, err = strconv.Atoi(response.Header.Get("CSeq"))
	if err != nil {
		log.Println(err)
		return
	}

	strContentLength := response.Header.Get("Content-Length")
	if len(strContentLength) == 0 {
		return
	}
	contentLength, err := strconv.Atoi(strContentLength)
	if err != nil {
		log.Println(err)
		return
	}

	if contentLength > MaxRtspBodyLen {
		err = fmt.Errorf("Content-Length: %d => too big", contentLength)
		log.Println(err)
		return
	}
	if contentLength < 0 {
		err = fmt.Errorf("Content-Length: %d => too small", contentLength)
		log.Println(err)
		return
	}

	if contentLength == 0 {
		log.Printf("Content-Length: %d => zero?", contentLength)
		return
	}

	response.Body, err = readbytes(reader, contentLength)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func readline(reader *bufio.Reader) (string, error) {
	l, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return l, err
	}
	if strings.HasSuffix(l, RTSPEndl) {
		l = strings.TrimSuffix(l, RTSPEndl)
		return l, nil
	}
	return l, errors.New("invalid line end")
}

func readbytes(reader *bufio.Reader, count int) (buf []byte, err error) {
	if count < 0 {
		err = fmt.Errorf("read %d ", count)
		return
	}
	buf = make([]byte, count)
	if count == 0 {
		return
	}

	cur := 0
	for cur < count {
		n := 0
		n, err = reader.Read(buf[cur:])
		if err != nil {
			log.Println(err)
			return
		}
		cur += n
	}
	return
}
