package rtsp

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

/*
smpte=10:12:33:20-
smpte=10:07:33-
smpte=10:07:00-10:07:33:05.01
smpte-25=10:07:00-10:07:33:05.01
*/

//! framte rate 107892 / hour ,every minute fast first 2 frame ,except 0,10,20,30,40,50 minute.
const SMPTE_30_drop_frame_rate = 29.97

const smpte_prefix = "smpte"

var regType *regexp.Regexp
var regInt *regexp.Regexp
var regTwoDigit *regexp.Regexp
var regNumber *regexp.Regexp

func init() {
	regType = regexp.MustCompile("smpte-[0-9]+=")
	regInt = regexp.MustCompile("[1-9][0-9]*")
	regTwoDigit = regexp.MustCompile("[0-9][0-9]")
	regNumber = regexp.MustCompile("[0-9]+")
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

type Smpte_Range struct {
	Drop      bool
	FrameRate int
	Begin     *Smpte_timestamp
	End       *Smpte_timestamp
}

func ParseSMPTE(line string) (err error, smpteRange *Smpte_Range) {

	if !IsSMPTE(line) {
		err = errors.New("not smpte timestamp")
		return
	}

	eqIndex := strings.Index(line, "=")
	hyphenIndex := strings.Index(line, "-")

	smpteRange = &Smpte_Range{}

	prefix := "smpte="
	if hyphenIndex < eqIndex {
		smpteRange.Drop = false
		prefix = regType.FindString(line)
		if len(prefix) == 0 {
			err = errors.New("bad framerate")
			return
		}
		strFrameRate := regInt.FindString(prefix)
		smpteRange.FrameRate, err = strconv.Atoi(strFrameRate)
		if err != nil {
			return
		}
	} else {
		smpteRange.Drop = true
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
		err, smpteRange.Begin = parseSampteRange(fromToArr[0])
		if err != nil {
			return
		}
	} else {
		err = errors.New("empty sampte from range")
		return
	}
	//to
	if len(fromToArr[1]) > 0 {
		err, smpteRange.End = parseSampteRange(fromToArr[1])
		if err != nil {
			return
		}
	} else {
		smpteRange.End = nil
	}

	return
}

func parseSampteRange(strRange string) (err error, ts *Smpte_timestamp) {
	ts = &Smpte_timestamp{Hours: 0, Minutes: 0, Seconds: 0, Frames: 0, Subframes: 0}

	subValues := strings.Split(strRange, ":")

	c := len(subValues)

	if c != 3 && c != 4 {
		err = errors.New("too many or less : for range")
		return
	}

	if len(subValues[0]) != 2 || len(subValues[1]) != 2 || len(subValues[2]) != 2 {
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
		frame_subFrame := regNumber.FindAllString(subValues[3], -1)
		count_frame_subFrame := len(frame_subFrame)
		if count_frame_subFrame == 0 || count_frame_subFrame > 2 {
			err = errors.New("smpte range invalid frame subframe")
			return
		}
		if count_frame_subFrame > 0 {
			if len(frame_subFrame[0]) < 2 {
				err = errors.New("at least need two digit hour")
				return
			}
			ts.Frames, err = strconv.Atoi(frame_subFrame[0])
			if err != nil {
				return
			}
		}
		if count_frame_subFrame > 1 {
			if len(frame_subFrame[1]) != 2 {
				err = errors.New("need two digit")
				return
			}
			ts.Subframes, err = strconv.Atoi(frame_subFrame[1])
			if err != nil {
				return
			}
		}
	}

	return
}
