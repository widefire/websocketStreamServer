package rtsp

import (
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
}

//NewClient create a rtsp client instance
func NewClient() (client *Client) {
	client = &Client{
		cseq:    1,
		methods: make(map[string]bool),
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
	_, err = client.conn.Write([]byte(request.String()))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(request)

	response, err := client.ReadResponse()
	if err != nil {
		log.Println(err)
	}
	log.Println(response)
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

//ReadResponse ...
func (client *Client) ReadResponse() (response *Response, err error) {
	response, err = ReadResponse(client.conn)
	if err != nil {
		log.Println(err)
		return
	}
	if response.CSeq != client.cseq {
		err = fmt.Errorf("CSeq error,get %d,local %d", response.CSeq, client.cseq)
		return
	}
	return
}
