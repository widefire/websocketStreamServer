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
	return
}
