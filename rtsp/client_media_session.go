package rtsp

import (
	"bytes"
	"errors"
	"log"
	"net/url"

	"github.com/writefire/websocketStreamServer/rtcp"
	"github.com/writefire/websocketStreamServer/rtp"
)

//ClientMediaSession build from sessionDescription for every media level
type ClientMediaSession struct {
	Control          string //from sdp
	RequestTransport *TransportItem
	ReplyTransport   *TransportItem
	RTPInfo          *RTPInfoItem
	URL              *url.URL
	SR               *rtcp.SenderReport
	rtpSSRC          uint32
	rtpPacketCount   uint32
	rtpSize          uint64
	src              *rtp.SourceStateInfo
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

//客户端通过发送getparam保活
func (desc *ClientMediaSession) handleByChannelAndData(channel byte, packet []byte) (reply []byte, err error) {
	if !desc.canHandle(channel) {
		err = errors.New("can't handle")
		log.Println(err)
		return
	}
	if int(channel) == desc.ReplyTransport.Interleaved.From {
		//rtp
		rtpHeader := rtp.NewHeader()
		offset := 0
		offset, err = rtpHeader.Decode(packet)
		if err != nil {
			log.Println(err)
			return
		}
		if offset > len(packet) {
			err = errors.New("header size big thran packet")
			log.Println(err)
		}
		desc.rtpPacketCount++
		desc.rtpSize += uint64(len(packet) - offset)
		desc.rtpSSRC = rtpHeader.SSRC

		if desc.src == nil {
			desc.src = &rtp.SourceStateInfo{}
			desc.src.InitSeq(rtpHeader.SequenceNumber)
		}
		desc.src.UpdateSeq(rtpHeader.SequenceNumber)

	} else {
		//rtcp

		compound := &rtcp.CompoundPacket{}
		err = compound.Decode(packet)
		if err != nil {
			log.Println(err)
			return
		}
		for _, rtcpPacket := range compound.Packet {
			header := rtcpPacket.Header()
			if header == nil {
				err = errors.New("rtcp packet get header failed")
				log.Println(err)
				return
			}
			if header.PacketType == rtcp.TypeSendReport {
				sr, ok := rtcpPacket.(*rtcp.SenderReport)
				if ok {
					desc.SR = sr
					log.Printf("%d %d %d %d", sr.SSRC, sr.SendersPacketCount, sr.SendersOctetCount, len(sr.Reports))
					log.Printf("%d %d", desc.rtpPacketCount, int(desc.rtpSize))
					rr := &rtcp.ReceiverReport{}
					rr.SSRC = desc.rtpSSRC
					rr.Reports = make([]*rtcp.ReceptionReport, 1)
					rr.Reports[0] = &rtcp.ReceptionReport{}
					rr.Reports[0].SSRC = sr.SSRC
					rr.Reports[0].FractionLost, rr.Reports[0].CumulativeLost, rr.Reports[0].ExtendedHighestSequenceNumber = desc.src.GetLostValues()
					rr.Reports[0].Jitter = 0 //to do
					rr.Reports[0].LSR = uint32(sr.NTPTime >> 16)
					rr.Reports[0].DLSR = 0 //send right now
					rrBuf := new(bytes.Buffer)
					err = rr.Encode(rrBuf)
					if err != nil {
						log.Println(err)
						return
					}
					reply = rrBuf.Bytes()
				} else {
					log.Println(ok)
				}
			} else {
				log.Printf("rtcp type %d not handle now", header.PacketType)
			}
		}
		log.Println("get sr")
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
