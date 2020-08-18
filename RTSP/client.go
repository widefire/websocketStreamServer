package rtsp

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

//Client RTSP client
type Client struct {
	url       *url.URL
	cseq      int
	conn      *net.TCPConn
	userAgent string
}

//NewClient create a rtsp client instance
func NewClient() (client *Client) {
	client = &Client{
		cseq: 1,
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

//SendOptions send OPTION
func (client *Client) SendOptions(header http.Header) (err error) {
	userAgentInHeader := header.Get("User-Agent")
	if len(userAgentInHeader) == 0 && len(client.userAgent) > 0 {
		header.Add("User-Agent", client.userAgent)
	}
	header.Add(RTSPCSeq, strconv.Itoa(client.cseq))
	request := NewRequest(MethodOPTIONS, client.url, header, nil)
	_, err = client.conn.Write([]byte(request.String()))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(request)
	return
}

//AddSeq ...
func (client *Client) AddSeq() {
	client.cseq++
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
