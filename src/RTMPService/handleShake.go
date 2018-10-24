package RTMPService

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"logger"
	"math/rand"
	"net"
	"strconv"
	"time"
	"wssAPI"
)

//RTMP握手协议，包括简单握手和复杂握手
const (
	rtmp_randomsize    = 1536
	rtmp_digestsize    = 32
	rtmp_serverVersion = 0x5033029
)

var GENUINE_FMS_KEY = []byte{
	0x47, 0x65, 0x6e, 0x75, 0x69, 0x6e, 0x65, 0x20,
	0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x46, 0x6c,
	0x61, 0x73, 0x68, 0x20, 0x4d, 0x65, 0x64, 0x69,
	0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Media Server 001
	0xf0, 0xee, 0xc2, 0x4a, 0x80, 0x68, 0xbe, 0xe8,
	0x2e, 0x00, 0xd0, 0xd1, 0x02, 0x9e, 0x7e, 0x57,
	0x6e, 0xec, 0x5d, 0x2d, 0x29, 0x80, 0x6f, 0xab,
	0x93, 0xb8, 0xe6, 0x36, 0xcf, 0xeb, 0x31, 0xae,
}
var GENUINE_FP_KEY = []byte{
	0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
	0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
	0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79,
	0x65, 0x72, 0x20, 0x30, 0x30, 0x31, /* Genuine Adobe Flash Player 001 */
	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
	0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
	0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
}

func rtmpHandleshake(conn net.Conn) (err error) {

	err = conn.SetReadDeadline(time.Now().Add(time.Duration(serviceConfig.TimeoutSec) * time.Second))
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer func() {
		conn.SetReadDeadline(time.Time{})
	}()
	buf, err := wssAPI.TcpRead(conn, rtmp_randomsize+1)
	if err != nil {
		return err
	}
	if buf[0] != 0x03 {
		return errors.New("invalid handleshake version")
	}
	//判断简单握手还是复杂握手
	var version uint32
	version, err = AMF0DecodeInt32(buf[5:])
	if version == 0 {
		err = sampleHandleShake(conn, buf[1:])
		if err != nil {
			return err
		}
	} else {
		err = complexHandleShake(conn, buf[1:])
		if err != nil {
			return nil
		}
	}

	return err
}

func sampleHandleShake(conn net.Conn, c1 []byte) (err error) {
	//s0
	//s1
	s01 := make([]byte, rtmp_randomsize+1)
	s01[0] = 0x03
	s01[1] = 0x00
	s01[2] = 0x00
	s01[3] = 0x00
	s01[4] = 0x00
	s01[5] = 0x00
	s01[6] = 0x00
	s01[7] = 0x00
	s01[8] = 0x00
	for i := 9; i <= rtmp_randomsize; i++ {
		s01[i] = byte(rand.Int() % 256)
	}

	sendRet, err := conn.Write(s01)
	if err != nil {
		return err
	}
	if sendRet != len(s01) {
		return errors.New("send s0 s1 failed")
	}
	//s2
	sendRet, err = conn.Write(c1)
	if err != nil {
		return err
	}
	if sendRet != len(c1) {
		return errors.New("send s2 failed")
	}
	//c2
	c2, err := wssAPI.TcpRead(conn, rtmp_randomsize)
	if err != nil {
		return err
	}
	for i := 0; i < rtmp_randomsize; i++ {
		if c2[i] != s01[i+1] {
			return errors.New("rtmp handleshake error:c2 != s1 ")
		}
	}
	return err
}

func complexHandleShake(conn net.Conn, c1 []byte) (err error) {

	scheme, _, digest := validClient(c1)
	if scheme != 0 && scheme != 1 {
		return errors.New("complex handleshake failed,not valid c1 data")
	}
	//s1
	s1 := createComplexS1()
	off := getDigestOffset(s1, scheme)
	s1p := make([]byte, rtmp_randomsize-rtmp_digestsize)
	copy(s1p, s1[0:off])
	copy(s1p[off:], s1[off+rtmp_digestsize:])
	h := hmac.New(sha256.New, GENUINE_FMS_KEY[:36])
	h.Write(s1p)
	s1digest := h.Sum(nil)
	copy(s1[off:off+rtmp_digestsize], s1digest)
	//s2
	h2 := hmac.New(sha256.New, GENUINE_FMS_KEY[:68])
	h2.Write(digest)
	s2tmpHash := h2.Sum(nil)
	s2Random := createComplexS2()
	h2 = hmac.New(sha256.New, s2tmpHash)
	h2.Write(s2Random)
	s2Hash := h2.Sum(nil)
	//send s0 s1 s2
	tmp8 := make([]byte, 1)
	tmp8[0] = 0x03
	_, err = conn.Write(tmp8)
	if err != nil {
		return err
	}
	_, err = conn.Write(s1)
	if err != nil {
		return err
	}
	_, err = conn.Write(s2Random)
	if err != nil {
		return err
	}
	_, err = conn.Write(s2Hash)
	if err != nil {
		return err
	}

	//recv c2
	c2, err := wssAPI.TcpRead(conn, rtmp_randomsize)
	if err != nil {
		return err
	}
	if len(c2) != rtmp_randomsize {
		return errors.New("invalid c2 recved")
	}
	//for i := 0; i < rtmp_randomsize; i++ {
	//	if c2[i] != s1[i] {
	//		return errors.New("rtmp handleshake error:c2 != s1 ")
	//	}
	//}
	//cal c2

	return err
}

func createComplexS1() (s1 []byte) {

	s1 = make([]byte, rtmp_randomsize)
	s1[0] = 0x00
	s1[1] = 0x00
	s1[2] = 0x00
	s1[3] = 0x00
	s1[4] = byte((rtmp_serverVersion >> 24) & 0xff)
	s1[5] = byte((rtmp_serverVersion >> 16) & 0xff)
	s1[6] = byte((rtmp_serverVersion >> 8) & 0xff)
	s1[7] = byte((rtmp_serverVersion >> 0) & 0xff)
	for i := 8; i < rtmp_randomsize; i++ {
		s1[i] = byte(rand.Int() % 256)
	}
	return s1
}
func createComplexS2() (s2 []byte) {

	s2 = make([]byte, rtmp_randomsize-rtmp_digestsize)
	for i := 0; i < rtmp_randomsize-rtmp_digestsize; i++ {
		s2[i] = byte(rand.Int() % 256)
	}
	return s2
}

func validClient(buf []byte) (scheme int, challenge []byte, digest []byte) {
	challenge, digest, err := validClientScheme(buf, 1)
	if err == nil {
		return 0, challenge, digest
	} else {
		logger.LOGI(err.Error())
	}
	challenge, digest, err = validClientScheme(buf, 0)
	if err == nil {
		return 1, challenge, digest
	} else {
		logger.LOGE(err.Error())
	}
	return -1, challenge, digest
}

func validClientScheme(buf []byte, scheme int) (challenge []byte, digest []byte, err error) {
	digestOffset := getDigestOffset(buf, scheme)
	challengeOffset := getDHOffset(buf, scheme)
	if digestOffset == -1 || challengeOffset == -1 {
		err = errors.New("err scheme")
		return challenge, digest, err
	}
	digest = make([]byte, rtmp_digestsize)
	p := make([]byte, rtmp_randomsize-rtmp_digestsize)
	copy(digest, buf[digestOffset:digestOffset+rtmp_digestsize])
	if digestOffset != 0 {
		copy(p, buf[:digestOffset])
		copy(p[digestOffset:], buf[digestOffset+rtmp_digestsize:])
	} else {
		copy(p, buf)
	}
	h := hmac.New(sha256.New, GENUINE_FP_KEY[:30])
	h.Write(p)
	tmpHash := h.Sum(nil)
	if 0 != bytes.Compare(tmpHash, digest) {
		err = errors.New(string("rtmp schme test:" + strconv.Itoa(scheme) + " failed"))
		return challenge, digest, err
	}
	challenge = make([]byte, 128)
	copy(challenge, buf[challengeOffset:challengeOffset+128])
	return challenge, digest, err
}

func getDigestOffset(buf []byte, scheme int) int {
	if 0 == scheme {
		var offset int
		offset = int(buf[8]) + int(buf[9]) + int(buf[10]) + int(buf[11])
		offset = (offset % 728) + 8 + 4
		if offset+32 > 1536 {
			return -1
		}
		return offset
	} else if 1 == scheme {
		var offset int
		offset = int(buf[772]) + int(buf[773]) + int(buf[774]) + int(buf[775])
		offset = (offset % 728) + 772 + 4
		if offset+32 > 1536 {
			return -1
		}
		return offset
	}
	return -1
}

func getDHOffset(buf []byte, scheme int) int {
	if 0 == scheme {
		var offset int
		offset = int(buf[1532]) + int(buf[1533]) + int(buf[1534]) + int(buf[1535])
		offset = (offset % 632) + 772
		if offset+128 > 1536 {
			return -1
		}
		return offset
	} else if 1 == scheme {
		var offset int
		offset = int(buf[768]) + int(buf[769]) + int(buf[770]) + int(buf[771])
		offset = (offset % 632) + 8
		if offset+128 > 1536 {
			return -1
		}
		return offset
	}
	return -1
}
