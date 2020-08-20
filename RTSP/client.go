package rtsp

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/writefire/websocketStreamServer/sdp"
)

//Client RTSP client
type Client struct {
	url                *url.URL
	cseq               int
	conn               *net.TCPConn
	userAgent          string
	methods            map[string]bool
	sessionDescription *sdp.SessionDescription
	session            string
	timeout            int
	replyTransport     *Transport
	replyRTPInfo       *RTPInfo
}

//NewClient create a rtsp client instance
func NewClient() (client *Client) {
	client = &Client{
		cseq:    1,
		methods: make(map[string]bool),
		timeout: 60,
	}
	return
}

//SetUserAgent ...
func (client *Client) SetUserAgent(userAgent string) {
	client.userAgent = userAgent
}

//Dial connect to RTSP server
func (client *Client) Dial(rawURL string) (err error) {
	client.url, err = parseURL(rawURL)
	if err != nil {
		log.Println(err)
		return
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", client.url.Host)
	if err != nil {
		log.Println(err)
		return
	}
	client.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

//AddCSeq ...
func (client *Client) AddCSeq() {
	client.cseq++
}

func (client *Client) prepareHeader(header http.Header) http.Header {
	ret := http.Header{}
	if header != nil {
		for k, v := range header {
			ret[k] = v
		}
	}
	userAgentInHeader := ret.Get("User-Agent")
	if len(userAgentInHeader) == 0 && len(client.userAgent) > 0 {
		ret.Set("User-Agent", client.userAgent)
	}
	ret.Set(RTSPCSeq, strconv.Itoa(client.cseq))
	return ret
}

func (client *Client) checkResponse(response *Response) (err error) {

	if response.StatusCode != StatusOK {
		err = fmt.Errorf("%s", response.StatusDesc)
		log.Println(err)
		return
	}
	if response.CSeq != client.cseq {
		err = fmt.Errorf("response cseq %d,clent cseq %d", response.CSeq, client.cseq)
		log.Println(err)
		return
	}

	return
}

//ReadResponse ...
func (client *Client) ReadResponse() (response *Response, err error) {
	response, err = ReadResponse(bufio.NewReader(client.conn))
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(response)
	return
}

//SendOptions send OPTION
func (client *Client) SendOptions(header http.Header) (err error) {
	header = client.prepareHeader(header)
	request := NewRequest(MethodOPTIONS, client.url, header, nil)
	_, err = client.conn.Write([]byte(request.String()))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(request)

	response, err := client.ReadResponse()
	if err != nil {
		log.Println(err)
		return
	}

	err = client.checkResponse(response)
	if err != nil {
		log.Println(err)
		return
	}

	methods := strings.Split(response.Header.Get("Public"), ", ")
	client.methods = make(map[string]bool)
	for _, m := range methods {
		client.methods[m] = true
	}

	return
}

//SendDescribe send DESCRIBE
func (client *Client) SendDescribe(header http.Header) (err error) {
	b, _ := client.methods[MethodDESCRIBE]
	if !b {
		err = errors.New("not support DESCRIBE")
		log.Println(err)
		return
	}
	header = client.prepareHeader(header)
	//support sdp only now
	header.Set("Accept", "application/sdp")
	request := NewRequest(MethodDESCRIBE, client.url, header, nil)
	log.Println(request)
	_, err = client.conn.Write([]byte(request.String()))
	if err != nil {
		log.Println(err)
		return
	}

	response, err := client.ReadResponse()
	if err != nil {
		log.Println(err)
	}

	err = client.checkResponse(response)
	if err != nil {
		log.Println(err)
		return
	}

	if len(response.Body) == 0 {
		err = errors.New("response no body")
		log.Println(err)
		return
	}
	client.sessionDescription, err = sdp.ParseSDP(string(response.Body))
	if err != nil {
		log.Println(err)
		return
	}

	return
}

//SendSetup send setup
func (client *Client) SendSetup(header http.Header, tryUDOFirst bool) (err error) {
	if tryUDOFirst {
		return client.setupUDP(header)
	}
	return client.setupTCP(header)
}

func (client *Client) setupUDP(header http.Header) (err error) {
	log.Println("todo,try udp ")
	return client.setupTCP(header)
}

func (client *Client) setupTCP(header http.Header) (err error) {
	b, _ := client.methods[MethodSETUP]
	if !b {
		err = errors.New("not support SETUP")
		log.Println(err)
		return
	}
	header = client.prepareHeader(header)
	header.Set("Transport", "RTP/AVP/TCP;interleaved=0-1")
	request := NewRequest(MethodSETUP, client.url, header, nil)
	log.Println(request)
	_, err = client.conn.Write([]byte(request.String()))
	if err != nil {
		log.Println(err)
		return
	}
	response, err := client.ReadResponse()
	if err != nil {
		log.Println(err)
	}

	err = client.checkResponse(response)
	if err != nil {
		log.Println(err)
		return
	}

	//decode session
	sessionandtimeout := response.Header.Get("Session")
	if len(sessionandtimeout) > 0 {
		subs := strings.SplitN(sessionandtimeout, ";", 2)
		client.session = subs[0]
		if len(subs[1]) > 0 {
			if strings.HasPrefix(subs[1], "timeout=") {
				client.timeout, err = strconv.Atoi(strings.TrimPrefix(subs[1], "timeout="))
				if err != nil {
					log.Println(err)
					//ignore
					err = nil
					return
				}
			} else {
				log.Printf("%s invalid timeout \r\n", subs[1])

			}
		}
	}

	client.replyTransport, err = ParseTransport(response.Header)
	if err != nil {
		log.Println(err)
		return
	}

	if client.replyTransport == nil {
		err = fmt.Errorf("setup response no transport")
		log.Println(err)
		return
	}

	return
}

//SendPlay send play
func (client *Client) SendPlay(header http.Header) (err error) {
	b, _ := client.methods[MethodPLAY]
	if !b {
		err = errors.New("not support PLAY")
		log.Println(err)
		return
	}
	header = client.prepareHeader(header)
	request := NewRequest(MethodPLAY, client.url, header, nil)
	log.Println(request)
	_, err = client.conn.Write([]byte(request.String()))
	if err != nil {
		log.Println(err)
		return
	}
	response, err := client.ReadResponse()
	if err != nil {
		log.Println(err)
	}

	err = client.checkResponse(response)
	if err != nil {
		log.Println(err)
		return
	}

	client.replyRTPInfo, err = ParseRTPInfo(response.Header)
	if err != nil {
		log.Println(err)
		return
	}
	if client.replyRTPInfo == nil {
		err = fmt.Errorf("no RTPInfo in Play response")
		return
	}

	return
}

//readTCPStream  rtsp rtp rtcp all in tcp
func (client *Client) readTCPStream() (err error) {
	reader := bufio.NewReader(client.conn)
	if reader == nil {
		err = errors.New("NewReader failed")
		log.Println(err)
		return
	}
	dallor, err := reader.ReadByte()
	if err != nil {
		log.Println(err)
		return
	}
	if dallor == '$' {
		log.Println("get $")
	} else {
		err = reader.UnreadByte()
		if err != nil {
			log.Println(err)
			return
		}
		var response *Response
		response, err = ReadResponse(reader)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(response)
	}
	return
}
