package rtsp

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

//RangeType ...
type RangeType int32

//All Range types
const (
	_ RangeType = iota
	RangeSmpte
	RangeNpt
	RangeClock //utc
)

//Range ...
type Range struct {
	RangeType       RangeType
	RangesSpecifier interface{}
	UtcTime         *UTCDateTime
}

//RangeKey ...
const RangeKey = "Range"

//ParseRange may reutrn both r and err are nil
func ParseRange(header http.Header) (r *Range, err error) {
	RangeValue := header.Get(RangeKey)
	if len(RangeValue) == 0 {
		return nil, nil
	}
	r = &Range{}
	subStr := strings.SplitN(RangeValue, ";", 2)
	if IsUTC(subStr[0]) {
		r.RangesSpecifier, err = ParseUTC(subStr[0])
		if err != nil {
			log.Println(err)
			return
		}
		r.RangeType = RangeClock
	} else if IsSMPTE(subStr[0]) {
		r.RangesSpecifier, err = ParseSMPTE(subStr[0])
		if err != nil {
			log.Println(err)
			return
		}
		r.RangeType = RangeSmpte
	} else if IsNPT(subStr[0]) {
		r.RangesSpecifier, err = ParseNPT(subStr[0])
		if err != nil {
			log.Println(err)
			return
		}
		r.RangeType = RangeNpt
	} else {
		err = fmt.Errorf("%s not support range type", subStr[1])
		log.Println(err)
		return
	}

	if len(subStr) == 2 {
		if strings.HasPrefix(subStr[1], "time=") {
			r.UtcTime, err = parseDateTime(strings.TrimPrefix(RangeValue, "time="))
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			err = fmt.Errorf("invalid range %s", RangeValue)
			log.Println(err)
			return
		}
	}

	return
}
