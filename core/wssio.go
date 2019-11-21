package core

type WSSIO interface {
	Read(data []uint8, len int64) (num int64, err error)
	Write(data []uint8) (num int64, err error)
	Seek(pos int64) (err error)
	Open(path string) (err error)
	Close()
}
