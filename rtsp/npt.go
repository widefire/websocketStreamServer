package rtsp

import (
	"errors"
	"strconv"
	"strings"
)

//Normal play time

//IsNPT ...
func IsNPT(line string) bool {
	return strings.HasPrefix(line, "npt=")
}

//NptTime ...
type NptTime struct {
	HH          int
	MM          int
	SS          int
	SSfractions int
}

//NptRange ...
type NptRange struct {
	Begin *NptTime //nil for now
	End   *NptTime //nil for no end
}

//ParseNPT ...
func ParseNPT(line string) (nptRange *NptRange, err error) {
	if !IsNPT(line) {
		err = errors.New("not npt")
		return
	}

	rawRange := strings.TrimPrefix(line, "npt=")
	if len(rawRange) < 0 {
		err = errors.New("invalid npt range")
		return
	}

	if 1 != strings.Count(rawRange, "-") {
		err = errors.New("need one - in npt range")
		return
	}

	subRanges := strings.Split(rawRange, "-")
	if len(subRanges) != 2 {
		err = errors.New("bad npt ranges")
		return
	}

	nptRange = &NptRange{}

	if subRanges[0] == "now" {
		nptRange.Begin = nil
	} else {
		nptRange.Begin, err = parseNptTime(subRanges[0])
		if err != nil {
			return
		}
	}

	if len(subRanges[1]) > 0 {
		nptRange.End, err = parseNptTime(subRanges[1])
		if err != nil {
			return
		}
	}

	return
}

func parseNptSec(npttime string) (sec, fraction int, err error) {
	dotCount := strings.Count(npttime, ".")
	if dotCount == 0 {
		sec, err = strconv.Atoi(npttime)
		if err != nil {
			return
		}
	} else if dotCount == 1 {
		ssfract := strings.Split(npttime, ".")
		if len(ssfract) != 2 || len(ssfract[0]) == 0 || len(ssfract[1]) == 0 {
			err = errors.New("invalid npt time sec fraction")
			return
		}
		sec, err = strconv.Atoi(ssfract[0])
		if err != nil {
			return
		}
		fraction, err = strconv.Atoi(ssfract[1])
		if err != nil {
			return
		}
		if fraction < 0 {
			err = errors.New("second fraction <0 ")
			return
		}
	} else {
		err = errors.New("invalid npt time sec")
		return
	}

	return
}

func parseNptTime(npttime string) (nptTime *NptTime, err error) {

	hmsCount := strings.Count(npttime, ":")
	if hmsCount == 0 {
		//npt-sec
		nptTime = &NptTime{}
		nptTime.SS, nptTime.SSfractions, err = parseNptSec(npttime)
	} else if hmsCount == 2 {
		//npt-hhmmss
		nptTime = &NptTime{}
		hhmmss := strings.Split(npttime, ":")
		if len(hhmmss) != 3 {
			err = errors.New("invalid hh mm ss")
			return
		}
		nptTime.HH, err = strconv.Atoi(hhmmss[0])
		if err != nil {
			return
		}
		if nptTime.HH < 0 {
			err = errors.New("npt hh should positive number")
			return
		}
		nptTime.MM, err = strconv.Atoi(hhmmss[1])
		if err != nil {
			return
		}
		if nptTime.MM < 0 || nptTime.MM > 59 {
			err = errors.New("npt mm should 0-59")
			return
		}
		nptTime.SS, nptTime.SSfractions, err = parseNptSec(hhmmss[2])
		if err != nil {
			return
		}
		if nptTime.SS < 0 || nptTime.SS > 59 {
			err = errors.New("npt ss should 0-59")
			return
		}
	} else {
		err = errors.New("bad npt-time")
		return
	}

	return
}
