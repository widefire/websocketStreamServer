package rtsp

//RangeType ...
type RangeType int32

//All Range types
const (
	_ RangeType = iota
	RangeSmpte
	RangeSmpte30Drop
	RangeNpt
	RangeClock
)
