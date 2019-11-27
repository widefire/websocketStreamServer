package rtsp

import (
	"testing"
)

/*
test smpte
Examples:
     smpte=10:12:33:20-
     smpte=10:07:33-
     smpte=10:07:00-10:07:33:05.01
     smpte-25=10:07:00-10:07:33:05.01
*/
func TestCheckSMPTEOnlyFrom(t *testing.T) {
	t1 := "smpte=10:12:33:20-"
	if !IsSMPTE(t1) {
		t.Error("check is smpte failed")
	}

	err, smpteRange := ParseSMPTE(t1)
	if err != nil {
		t.Error(err)
	}

	if !smpteRange.Drop {
		t.Error("should drop")
	}

	if smpteRange.Begin == nil {
		t.Error("from should not nil")
	}

	from := smpteRange.Begin
	if from.Hours != 10 || from.Minutes != 12 || from.Seconds != 33 || from.Frames != 20 || from.Subframes != 0 {
		t.Error("parse time failed ")
	}

	if smpteRange.End != nil {
		t.Error("t0 should nil")
	}

	return
}

func TestCheckSMPTEFromTo(t *testing.T) {
	t1 := "smpte-25=10:07:00-10:07:33:05.01"
	if !IsSMPTE(t1) {
		t.Error("check is smpte failed")
	}

	err, smpteRange := ParseSMPTE(t1)
	if err != nil {
		t.Error(err)
	}

	if smpteRange.Drop {
		t.Error("should not drop")
	}

	if smpteRange.FrameRate != 25 {
		t.Error("bad frame rate")
	}

	from := smpteRange.Begin
	if from == nil {
		t.Error("from should not nil")
	}

	if from.Hours != 10 || from.Minutes != 7 || from.Seconds != 0 || from.Frames != 0 || from.Subframes != 0 {
		t.Error("parse time failed ")
	}

	if smpteRange.End == nil {
		t.Error("t0 should not nil")
	}

	return
}

// test npt
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

//test utc
func TestCheckClock(t *testing.T) {
	err, utc := ParseUTC("clock=19961108T143720.25Z-19961108T143720Z")
	if err != nil {
		t.Error(err)
		return
	}
	if utc == nil || utc.BeginDateTime == nil || utc.EndDateTime == nil {
		return
	}
}
