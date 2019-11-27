package core

// WSSIO is WSS io interface
type WSSIO interface {
	Read(data []uint8, len int64) (num int64, err error)
	Write(data []uint8) (num int64, err error)
	// Seek to offset
	// offset the offset you seek
	// whence 0 for relative to the origin of the stream, 1 for relative to the current
	// offset, 2 for relative to the end
	Seek(offset int64, whence int) (err error)
	Open(path string) (err error)
	Close() (err error)
}
