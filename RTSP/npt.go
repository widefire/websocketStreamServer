package RTSP

import (
	"errors"
	"strings"
)

//Normal play time

func IsNPT(line string) bool {
	return strings.HasPrefix(line, "npt=")
}

type NPT_Time struct {
	HH          int
	MM          int
	SS          int
	SSfractions int
}

type NPT_Range struct {
	Begin *NPT_Time //nil for now
	End   *NPT_Time //nil for no end
}

func ParseNPT(line string) (err error, nptRange *NPT_Range) {
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

	nptRange = &NPT_Range{}

	if subRanges[0] == "now" {
		nptRange.Begin = nil
	} else {
		err, nptRange.Begin = parse_NPT_Time(subRanges[0])
		if err != nil {
			return
		}
	}

	if len(subRanges[1]) > 0 {
		err, nptRange.End = parse_NPT_Time(subRanges[1])
		if err != nil {
			return
		}
	}

	return
}

func parse_npt_sec(npttime string) (err error, sec, fraction int) {
	return
}

func parse_NPT_Time(npttime string) (err error, nptTime *NPT_Time) {

	hmsCount := strings.Count(npttime, ":")
	if hmsCount == 1 {
		//npt-sec
	} else if hmsCount == 3 {
		//npt-hhmmss
	} else {
		err = errors.New("bad npt-time")
		return
	}

	// hms := strings.Split(subRanges[0], ":")
	// chms := len(hms)
	// if chms == 0 || chms > 3 {
	// 	err = errors.New("bad npt time")
	// 	return
	// }
	// nptRange.Begin = &NPT_Time{}
	// if chms == 3 {
	// 	nptRange.Begin.HH, err = strconv.Atoi(hms[0])
	// 	if err != nil {
	// 		return
	// 	}
	// 	if nptRange.Begin.HH < 0 {
	// 		err = errors.New("npt-hh must positive ")
	// 		return
	// 	}

	// 	nptRange.Begin.MM, err = strconv.Atoi(hms[1])
	// 	if err != nil {
	// 		return
	// 	}
	// 	if nptRange.Begin.MM < 0 || nptRange.Begin.MM > 59 {
	// 		err = errors.New("npt-mm must 0-59")
	// 		return
	// 	}

	// }
	return
}
