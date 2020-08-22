package rtsp

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

//RTPInfoItem ...
type RTPInfoItem struct {
	StreamURL string
	Seq       int
	Rtptime   int
}

//RTPInfo ...
type RTPInfo struct {
	Items []*RTPInfoItem
}

//ParseRTPInfo ...
func ParseRTPInfo(header http.Header) (rtpInfo *RTPInfo, err error) {
	rtpInfo = &RTPInfo{}
	rtpInfo.Items = make([]*RTPInfoItem, 0)
	value := header.Get("RTP-Info")
	if len(value) == 0 {
		return
	}
	strItems := strings.Split(value, ",")
	for _, str := range strItems {
		subItems := strings.Split(str, ";")
		item := &RTPInfoItem{}
		for _, subStr := range subItems {
			if strings.HasPrefix(subStr, "url=") {
				item.StreamURL = strings.TrimPrefix(subStr, "url=")
			} else if strings.HasPrefix(subStr, "seq=") {
				item.Seq, err = strconv.Atoi(strings.TrimPrefix(subStr, "seq="))
				if err != nil {
					log.Println(err)
					return
				}
			} else if strings.HasPrefix(subStr, "rtptime=") {
				item.Rtptime, err = strconv.Atoi(strings.TrimPrefix(subStr, "rtptime="))
				if err != nil {
					log.Println(err)
					return
				}
			} else {
				log.Printf("not support rtp info %s\r\n", value)
			}
		}
		rtpInfo.Items = append(rtpInfo.Items, item)
	}
	return
}
