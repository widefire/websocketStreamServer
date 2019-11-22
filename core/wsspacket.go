package core

// WSSPacket stores compressed data
type WSSPacket struct {
	Data     []uint8   //packet data
	SideData [][]uint8 //side data
	Pts      int64     //presentation timestamp
	Dts      int64     //decompression timestamp
	Duration int64     //duration
	Size     int       //packet size
	StreamID int       //packet stream id
	Flags    int       // packet flag
	Pos      int64     //packet posion in stream
}
