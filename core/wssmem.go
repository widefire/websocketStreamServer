package core

import "errors"

type WSSMEM struct {
	buf      []uint8
	readPos  int
	writePos int
}

func NewWSSMEM(buf []uint8) WSSIO {
	mem := &WSSMEM{}
	mem.buf = buf
	return mem
}

func (m *WSSMEM) Read(data []uint8) (int, error) {
	if len(data) > len(m.buf)-m.readPos {
		return 0, errors.New("not enough data for read")
	}
	count := copy(data, m.buf[m.readPos:len(data)])
	return count, nil
}

func (m *WSSMEM) Write(data []uint8) (int, error) {
	if len(data) > len(m.buf)-m.writePos {
		return 0, errors.New("not enough space for write")
	}
	count := copy(m.buf[m.writePos:], data)
	return count, nil
}

func (m *WSSMEM) Seek(offset int64, whence int) error {
	switch whence {
	case 0:
		m.writePos = int(offset)
		m.readPos = int(offset)
	case 1:
		m.writePos += int(offset)
		m.readPos += int(offset)
	case 2:
		m.writePos = len(m.buf) - int(offset)
		m.readPos = len(m.buf) - int(offset)
	default:
		return errors.New("whence error")
	}
	return nil
}

func (m *WSSMEM) Open(path string) error {
	return errors.New("NOT IMPLETE")
}

func (m *WSSMEM) Close() error {
	m.buf = nil
	m.readPos = 0
	m.writePos = 0
	return nil
}
