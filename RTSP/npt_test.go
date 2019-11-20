package RTSP

import (
	"testing"
)

func TestCheckNpt(t *testing.T) {
	{
		npt := "npt=123.45-125"
		err, nptRange := ParseNPT(npt)
		if err != nil {
			t.Error(err)
		}
		if nptRange == nil || nptRange.Begin == nil || nptRange.End == nil {
			t.Error("result error")
			return
		}
	}
	{
		npt := "npt=12:05:35.3-"
		err, nptRange := ParseNPT(npt)
		if err != nil {
			t.Error(err)
		}
		if nptRange == nil || nptRange.Begin == nil || nptRange.End != nil {
			t.Error("result error")
			return
		}
	}

	{
		npt := "npt=now-"
		err, nptRange := ParseNPT(npt)
		if err != nil {
			t.Error(err)
		}
		if nptRange == nil || nptRange.Begin != nil || nptRange.End != nil {
			t.Error("result error")
			return
		}
	}
}
