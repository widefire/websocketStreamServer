package wssAPI

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/ioutil"
	"logger"
	"net"
	"os"
	"reflect"
	"time"
)

func CheckDirectory(dir string) (bool, error) {
	_, err := os.Stat(dir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, err
	}
	return false, err
}

func CreateDirectory(dir string) (bool, error) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		logger.LOGE(err.Error())
		return false, err
	}
	return true, nil
}

func ReadFileAll(filename string) (data []byte, err error) {
	fp, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer fp.Close()
	data, err = ioutil.ReadAll(fp)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	return
}

func TcpRead(conn net.Conn, size int) (data []byte, err error) {
	data = make([]byte, size)
	received := 0
	for received < int(size) {
		ret, err := conn.Read(data[received:])
		if err != nil {
			logger.LOGE(err.Error())
			return data[:received], err
		}
		received += ret
	}
	return data, err
}

func TcpReadTimeout(conn net.Conn, size int, millSec int) (data []byte, err error) {
	if millSec > 0 {
		err = conn.SetReadDeadline(time.Now().Add(time.Duration(millSec) * time.Millisecond))
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		defer func() {
			conn.SetReadDeadline(time.Time{})
		}()
	}
	data = make([]byte, size)
	received := 0
	for received < int(size) {
		ret, err := conn.Read(data[received:])
		if err != nil {
			logger.LOGE(err.Error())
			return data[:received], err
		}
		received += ret
	}
	return data, err
}

func TcpReadTimeDuration(conn net.Conn, size int, duration time.Duration) (data []byte, err error) {
	if duration > 0 {
		err = conn.SetReadDeadline(time.Now().Add(duration))
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		defer func() {
			conn.SetReadDeadline(time.Time{})
		}()
	}
	data = make([]byte, size)
	received := 0
	for received < int(size) {
		ret, err := conn.Read(data[received:])
		if err != nil {
			logger.LOGE(err.Error())
			return data[:received], err
		}
		received += ret
	}
	return data, err
}

func TcpWrite(conn net.Conn, data []byte) (writed int, err error) {
	err = conn.SetReadDeadline(time.Now().Add(time.Hour))
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer func() {
		conn.SetReadDeadline(time.Time{})
	}()
	for writed < len(data) {
		ret, err := conn.Write(data[writed:])
		if err != nil {
			//logger.LOGE(err.Error())
			return writed, err
		}
		writed += ret
	}
	return
}

func TcpWriteTimeOut(conn net.Conn, data []byte, millSec int) (writed int, err error) {
	err = conn.SetWriteDeadline(time.Now().Add(time.Millisecond * time.Duration(millSec)))
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer func() {
		conn.SetWriteDeadline(time.Time{})
	}()
	for writed < len(data) {
		ret, err := conn.Write(data[writed:])
		if err != nil {
			logger.LOGE(err.Error())
			return writed, err
		}
		writed += ret
	}
	return
}

func TcpWriteTimeDuration(conn net.Conn, data []byte, duration time.Duration) (writed int, err error) {
	err = conn.SetWriteDeadline(time.Now().Add(duration))
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer func() {
		conn.SetWriteDeadline(time.Time{})
	}()
	for writed < len(data) {
		ret, err := conn.Write(data[writed:])
		if err != nil {
			logger.LOGE(err.Error())
			return writed, err
		}
		writed += ret
	}
	return
}

func GenerateGUID() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getMd5String(base64.URLEncoding.EncodeToString(b))
}

func getMd5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func InterfaceIsNil(val interface{}) bool {
	if nil == val {
		return true
	}
	return reflect.ValueOf(val).IsNil()
}

func InterfaceValid(val interface{}) bool {
	if nil == val {
		return false
	}
	return !reflect.ValueOf(val).IsNil()
}
