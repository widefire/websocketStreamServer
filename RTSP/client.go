package rtsp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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
	url                    *url.URL
	cseq                   int
	conn                   *net.TCPConn
	userAgent              string
	methods                map[string]bool
	sessionDescription     *sdp.SessionDescription
	clientMediaDescription []*ClientMediaDescription
	session                string
	timeout                int
}

//ClientMediaDescription build from sessionDescription for every media level
type ClientMediaDescription struct {
	Control          string //from sdp
	RequestTransport *TransportItem
	ReplyTransport   *TransportItem
	RTPInfo          *RTPInfoItem
	URL              *url.URL
}

func (desc *ClientMediaDescription) setRTPInfo(info *RTPInfo) {
	for _, item := range info.Items {
		itemURL, err := url.Parse(item.StreamURL)
		if err != nil {
			log.Println(err)
			return
		}
		if itemURL.Path == desc.URL.Path {
			desc.RTPInfo = item
		}
	}
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

	client.clientMediaDescription = make([]*ClientMediaDescription, len(client.sessionDescription.MediaDescriptions))
	for i, m := range client.sessionDescription.MediaDescriptions {
		desc := &ClientMediaDescription{}
		for _, attr := range m.Attribute {
			switch attr.AttributeName {
			case "control":
				if attr.Value != nil {
					desc.Control = *attr.Value
				} else {
					log.Println("control is nil")
				}
			default:
				log.Printf("not handle now : %s ", attr.AttributeName)
			}
		}
		desc.RequestTransport = &TransportItem{}
		err = desc.RequestTransport.ParseTransportSpec(m.Proto)
		if err != nil {
			log.Println(err)
			return
		}
		//set unicast
		desc.RequestTransport.Unicast = true
		client.clientMediaDescription[i] = desc

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
	if len(client.clientMediaDescription) == 0 {
		err = errors.New("no media can setup")
		log.Println(err)
		return
	}

	for i := 0; i < len(client.clientMediaDescription); i++ {
		header = client.prepareHeader(header)
		desc := client.clientMediaDescription[i]
		desc.RequestTransport.TransportProtocol = "RTP"
		desc.RequestTransport.Profile = "AVP"
		desc.RequestTransport.LowerTransport = "TCP"
		desc.RequestTransport.Interleaved = &IntRange{}
		desc.RequestTransport.Interleaved.From = 2 * i
		desc.RequestTransport.Interleaved.To = new(int)
		*desc.RequestTransport.Interleaved.To = 2*i + 1
		header.Set("Transport", desc.RequestTransport.String())
		desc.URL, err = url.Parse(client.url.String())
		if err != nil {
			log.Println(err)
			return
		}
		desc.URL.Path += "/"
		desc.URL.Path += desc.Control
		request := NewRequest(MethodSETUP, desc.URL, header, nil)
		log.Println(request)
		_, err = client.conn.Write([]byte(request.String()))
		if err != nil {
			log.Println(err)
			return
		}
		var response *Response
		response, err = client.ReadResponse()
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

		var trans *Transport
		trans, err = ParseTransport(response.Header)
		if err != nil {
			log.Println(err)
			return
		}
		if trans == nil || len(trans.Items) == 0 {
			err = fmt.Errorf("setup response no transport")
			log.Println(err)
			return
		}
		if len(trans.Items) > 1 {
			log.Println("a compound transport,to do")
		}
		desc.ReplyTransport = trans.Items[0]
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
	header.Set("Session", client.session)
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

	rtpInfo, err := ParseRTPInfo(response.Header)
	if err != nil {
		log.Println(err)
		return
	}
	if rtpInfo == nil {
		err = fmt.Errorf("no RTPInfo in Play response")
		log.Println(err)
		return
	}

	if len(rtpInfo.Items) != len(client.clientMediaDescription) {
		err = fmt.Errorf("%d %d rtp info count error", len(rtpInfo.Items), len(client.clientMediaDescription))
		log.Println(err)
		return
	}

	for _, desc := range client.clientMediaDescription {
		desc.setRTPInfo(rtpInfo)
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
		return err
	}
	if dallor == '$' {
		log.Println("get $")
		channel, err := reader.ReadByte()
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println(channel)
		var packetSize uint16
		err = binary.Read(reader, binary.BigEndian, &packetSize)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println(packetSize)
		if packetSize == 0 {
			log.Println("packet size is zero")
		} else {
			packet := make([]byte, packetSize)
			_, err = io.ReadFull(reader, packet)
			if err != nil {
				log.Println(err)
				return err
			}
			log.Println("read packet succeed")
		}
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
