package core

type WSSFormat interface {
	Open() error
	ReadMetadata() error
	WriteMetadata() error
	ReadPacket(*WSSPacket) error
	WritePacket(WSSPacket) error
	Seek(time int64) error
	Close() error
}
