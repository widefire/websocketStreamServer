package core

type WSSFrame struct {
	Data       [][8]uint8
	LineSize   [][8]uint8
	With       int
	Height     int
	Format     int
	Channels   int
	SampleRate int
	StreamID   int
}
