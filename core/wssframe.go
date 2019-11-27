package core

type WSSFrame struct {
	Data       [][]uint8
	LineSize   [][]uint8
	Width      int
	Height     int
	Format     int
	Channels   int
	SampleRate int
	StreamID   int
}
