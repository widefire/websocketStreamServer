package RTSP

import (
	"testing"
)

/*
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

	err, drop, frameRate, from, to := ParseSMPTE(t1)
	if err != nil {
		t.Error(err)
	}

	if !drop {
		t.Error("should drop")
	}

	ffr, ok := frameRate.(float64)
	if !ok {
		t.Error("not float32")
	}
	if ffr != SMPTE_30_drop_frame_rate {
		t.Error("bad frame rate")
	}

	if from == nil {
		t.Error("from should not nil")
	}

	if from.Hours != 10 || from.Minutes != 12 || from.Seconds != 33 || from.Frames != 20 || from.Subframes != 0 {
		t.Error("parse time failed ")
	}

	if to != nil {
		t.Error("t0 should nil")
	}

	return
}

func TestCheckSMPTEFromTo(t *testing.T) {
	t1 := "smpte-25=10:07:00-10:07:33:05.01"
	if !IsSMPTE(t1) {
		t.Error("check is smpte failed")
	}

	err, drop, frameRate, from, to := ParseSMPTE(t1)
	if err != nil {
		t.Error(err)
	}

	if drop {
		t.Error("should not drop")
	}

	ffr, ok := frameRate.(int)
	if !ok {
		t.Error("not float32")
	}
	if ffr != 25 {
		t.Error("bad frame rate")
	}

	if from == nil {
		t.Error("from should not nil")
	}

	if from.Hours != 10 || from.Minutes != 7 || from.Seconds != 0 || from.Frames != 0 || from.Subframes != 0 {
		t.Error("parse time failed ")
	}

	if to == nil {
		t.Error("t0 should not nil")
	}

	return
}
