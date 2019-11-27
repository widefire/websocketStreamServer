package core

import (
	"errors"
	"os"
)

type WSSFile struct {
	path string
	file *os.File
}

func NewWSSFile() WSSIO {
	return &WSSFile{}
}

func (f *WSSFile) Read(data []uint8) (num int, err error) {
	if f.file == nil {
		return 0, errors.New("must open file first")
	}
	count, err := f.file.Read(data)
	return count, err
}

func (f *WSSFile) Write(data []uint8) (num int, err error) {
	count, err := f.file.Write(data)
	return count, err
}

func (f *WSSFile) Seek(pos int64, whence int) error {
	_, err := f.file.Seek(pos, whence)
	return err
}

func (f *WSSFile) Open(path string) error {
	f.path = path
	file, err := os.OpenFile(path, os.O_RDWR, os.ModeExclusive)
	f.file = file
	return err
}

func (f *WSSFile) Close() error {
	return f.file.Close()
}
