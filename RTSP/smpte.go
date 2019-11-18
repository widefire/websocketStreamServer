package RTSP

import "strings"

//https://blog.csdn.net/andrew57/article/details/6752182
//https://developer.gnome.org/gst-plugins-libs/unstable/gst-plugins-base-libs-gstrtsprange.html

//! framte rate 107892 / hour ,every minute fast first 2 frame ,except 0,10,20,30,40,50 minute.
const SMPTE_30_drop_frame_rate = 29.97

const smpte_prefix = "smpte"

func IsSMPTE(line string) bool {
	if !strings.HasPrefix(line, smpte_prefix) {
		return false
	}

	if strings.Count(line, "=") != 1 || strings.Count(line, "-") != 1 {
		return false
	}

	eqIndex := strings.Index(line, "=")
	hyphenIndex := strings.Index(line, "-")

	if eqIndex == -1 || hyphenIndex == -1 || eqIndex > hyphenIndex {
		return false
	}

	return true
}

func 
