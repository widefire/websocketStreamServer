package RTSP

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

//https://blog.csdn.net/andrew57/article/details/6752182
/*
smpte=10:12:33:20-
smpte=10:07:33-
smpte=10:07:00-10:07:33:05.01
smpte-25=10:07:00-10:07:33:05.01
*/
//https://developer.gnome.org/gst-plugins-libs/unstable/gst-plugins-base-libs-gstrtsprange.html

//! framte rate 107892 / hour ,every minute fast first 2 frame ,except 0,10,20,30,40,50 minute.
const SMPTE_30_drop_frame_rate = 29.97

const smpte_prefix = "smpte"

var regType *regexp.Regexp
var regInt *regexp.Regexp

func init() {
	regType = regexp.MustCompile("smpte-[0-9]*=")
	regInt = regexp.MustCompile("[1-9][0-9]*")
}

func IsSMPTE(line string) bool {
	if !strings.HasPrefix(line, smpte_prefix) {
		return false
	}

	if strings.Count(line, "=") != 1 || (strings.Count(line, "-") != 1 && strings.Count(line, "-") != 2) {
		return false
	}

	eqIndex := strings.Index(line, "=")
	hyphenIndex := strings.Index(line, "-")

	if eqIndex == -1 || hyphenIndex == -1 {
		return false
	}

	return true
}

type Smpte_timestamp struct {
	Hours     int
	Minutes   int
	Seconds   int
	Frames    int
	Subframes int
}

func ParseSMPTE(line string) (err error, drop bool, frameRate interface{}, from, to *Smpte_timestamp) {

	if !IsSMPTE(line) {
		err = errors.New("not smpte timestamp")
		return
	}

	eqIndex := strings.Index(line, "=")
	hyphenIndex := strings.Index(line, "-")

	prefix := "smpte"
	if hyphenIndex < eqIndex {
		drop = false
		prefix = regType.FindString(line)
		if len(prefix) == 0 {
			err = errors.New("bad framerate")
			return
		}
		strFrameRate := regInt.FindString(prefix)
		frameRate, err = strconv.Atoi(strFrameRate)
		if err != nil {
			return
		}
	} else {
		drop = true
		frameRate = SMPTE_30_drop_frame_rate
	}

	from_to := strings.TrimPrefix(line, prefix)

	if len(from_to) == 0 {
		err = errors.New("no time range")
		return
	}

	fromToArr := strings.Split(from_to, "-")
	if len(fromToArr) != 2 {
		err = errors.New("bad sampte range")
		return
	}

	//from
	if len(fromToArr[0]) > 0 {

	} else {
		err = errors.New("empty sampte from range")
		return
	}
	//to
	if len(fromToArr[1]) > 0 {

	} else {
		to = nil
	}

	return
}

func parseSampteRange(strRange string) (err error, ts *Smpte_timestamp) {
	ts = &Smpte_timestamp{Hours: 0, Minutes: 0, Seconds: 0, Frames: 0, Subframes: 0}

	subValues := strings.Split(strRange, ":")

	c := len(subValues)

	if c <= 3 {
		err = errors.New("smpte range at least has h,m,s")
		return
	}
	if c > 4 {
		err = errors.New("smpte range too many \":\"")
		return
	}

	if len(subValues[0]) == 0 || len(subValues[1]) == 0 || len(subValues[2]) == 0 {
		err = errors.New("smpte range h,m,s can not empty")
		return
	}

	ts.Hours, err = strconv.Atoi(subValues[0])
	if err != nil {
		return
	}

	ts.Minutes, err = strconv.Atoi(subValues[1])
	if err != nil {
		return
	}

	ts.Seconds, err = strconv.Atoi(subValues[2])
	if err != nil {
		return
	}

	if c == 4 && len(subValues[3]) > 0 {
		frame_subFrame := regInt.FindAllString(subValues[3], -1)
		count_frame_subFrame := len(frame_subFrame)
		if count_frame_subFrame == 0 || count_frame_subFrame > 2 {
			err = errors.New("smpte range invalid frame subframe")
			return
		}
		if count_frame_subFrame > 0 {
			ts.Frames, err = strconv.Atoi(frame_subFrame[0])
			if err != nil {
				return
			}
		}
		if count_frame_subFrame > 1 {
			ts.Subframes, err = strconv.Atoi(frame_subFrame[1])
			if err != nil {
				return
			}
		}
	}

	return
}
