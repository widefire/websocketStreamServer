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

func TestCheckSMPTE(t *testing.T) {
	t1 := "smpte=10:12:33:20-"
	if !IsSMPTE(t1) {
		t.Error("check is smpte failed")
	}

	

	return
}
