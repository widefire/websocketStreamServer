package rtsp

import (
	"errors"
	"log"
	"net/url"
)

//ClientMediaSession build from sessionDescription for every media level
type ClientMediaSession struct {
	Control          string //from sdp
	RequestTransport *TransportItem
	ReplyTransport   *TransportItem
	RTPInfo          *RTPInfoItem
	URL              *url.URL
}

//canHandle check can hand this channel packet,just for tcp,udp is by port
func (desc *ClientMediaSession) canHandle(channel byte) bool {
	if desc.ReplyTransport == nil {
		log.Println("no reply transport ")
		return false
	}

	if desc.ReplyTransport.Interleaved == nil {
		log.Println("transport interleaved is nil")
		return false
	}

	return desc.ReplyTransport.Interleaved.isInRange(channel)
}

func (desc *ClientMediaSession) handleByChannelAndData(channel byte, packet []byte) (err error) {
	if !desc.canHandle(channel) {
		err = errors.New("can't handle")
		log.Println(err)
		return
	}
	if int(channel) == desc.ReplyTransport.Interleaved.From {
		//rtp
		log.Println("todo parse RTP")
	} else {
		//rtcp
		log.Println("todo parse RTCP")
	}
	return
}

func (desc *ClientMediaSession) setRTPInfo(info *RTPInfo) {
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
