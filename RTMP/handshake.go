package RTMP

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

const (
	RtmpHandShakePacketSize = 1536
	RtmpServerVersion       = 0x5033029
	RtmpDigestSize          = 32
	RtmpKeySize             = 128
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

func (r *RTMPFormat) HandShakeAsServer() error {
	buf := make([]uint8, 1)
	count, _ := r.io.Read(buf)
	if count != 1 {
		log.Println("read data error")
		return errors.New("handshake error on read version")
	}
	if buf[0] != 3 {
		log.Printf("client version %d is not support!", buf[0])
		return errors.New("version not support")
	}

	c1buf := make([]byte, RtmpHandShakePacketSize)
	count, err := r.io.Read(c1buf)
	if err != nil || count != len(c1buf) {
		log.Println("read c1 error")
		return errors.New("read  c1 error")
	}

	time := binary.BigEndian.Uint32(c1buf)
	version := binary.BigEndian.Uint32(c1buf[4:])
	r.remoteTime = time

	if version != 0 {
		return r.complexHandShake(c1buf)
	} else {
		return r.sampleHandShake(c1buf)
	}
}

// HandShakeAsClient handshake as client use sample handshake to client server
func (r *RTMPFormat) HandShakeAsClient() error {
	c0c1buf := make([]byte, RtmpHandShakePacketSize+1)
	//version
	c0c1buf[0] = 0x03
	for i := 9; i < RtmpHandShakePacketSize+1; i++ {
		c0c1buf[i] = byte(rand.Int() % 256)
	}
	count, err := r.io.Write(c0c1buf)
	if err != nil || count != len(c0c1buf) {
		log.Println("write c0c1 error")
		return errors.New("write c0c1 error")
	}

	s0s1buf := make([]byte, RtmpHandShakePacketSize+1)

	count, err = r.io.Read(s0s1buf)
	if count != len(s0s1buf) || err != nil {
		log.Println("get s0s1 error")
		return errors.New("get s0 s1 error")
	}

	if s0s1buf[0] != 0x03 {
		log.Println("unrecognize server version")
		return errors.New("unrecognized server version")
	}

	count, err = r.io.Write(s0s1buf[1:])

	if count != len(s0s1buf[1:]) || err != nil {
		log.Println("write c2 error")
		return errors.New("write c2 error")
	}

	s2buf := make([]byte, RtmpHandShakePacketSize)

	count, err = r.io.Read(s2buf)
	if count != len(s2buf) || err != nil {
		log.Println("get s2 error")
		return errors.New("get s2 error")
	}

	for i := 0; i < RtmpHandShakePacketSize; i++ {
		if s2buf[i] != c0c1buf[1+i] {
			log.Println("s2 not equal c1")
			return errors.New("s2 not equal c1")
		}
	}

	return nil
}

func (r *RTMPFormat) sampleHandShake(c1buf []byte) error {
	s0s1buf := make([]uint8, RtmpHandShakePacketSize+1)
	s0s1buf[0] = 0x3
	//copy time
	copy(s0s1buf[1:5], c1buf[:4])
	for i := 9; i < RtmpHandShakePacketSize+1; i++ {
		s0s1buf[i] = uint8(rand.Int() % 256)
	}

	count, err := r.io.Write(s0s1buf)
	if count != len(s0s1buf) || err != nil {
		log.Println("write s0s1 error")
		return errors.New("write s0s1 error")
	}

	count, err = r.io.Write(c1buf)
	if count != len(c1buf) || err != nil {
		log.Println("write s2 error")
		return errors.New("write s2 error")
	}

	c2buf := make([]uint8, RtmpHandShakePacketSize)
	count, err = r.io.Read(c2buf)
	if count != len(c2buf) || err != nil {
		log.Println("get c2 error")
		return errors.New("get c2 error")
	}

	for i := 0; i < RtmpHandShakePacketSize; i++ {
		if s0s1buf[i+1] != c2buf[i] {
			return errors.New("c2 not equal s1")
		}
	}
	return nil
}

func (r *RTMPFormat) complexHandShake(c1buf []byte) error {
	scheme := 0
	keyOffset, digestOffset := getKeyDigetstOffset(c1buf, scheme)
	ok, digest := checkClient(c1buf, keyOffset, digestOffset)
	if !ok {
		scheme = 1
		keyOffset, digestOffset = getKeyDigetstOffset(c1buf, scheme)
		ok, digest = checkClient(c1buf, keyOffset, digestOffset)
		if !ok {
			log.Println("complex handshake error on c1 check")
			return errors.New("complex handshake error on c1 check")
		}
	}
	s0s1buf := make([]byte, RtmpHandShakePacketSize+1)
	s0s1buf[0] = 0x03

	//set version
	s0s1buf[5] = byte((RtmpServerVersion >> 24) & 0xff)
	s0s1buf[6] = byte((RtmpServerVersion >> 16) & 0xff)
	s0s1buf[7] = byte((RtmpServerVersion >> 8) & 0xff)
	s0s1buf[8] = byte(RtmpServerVersion & 0xff)
	//s1 content
	for index := 9; index < RtmpHandShakePacketSize; index++ {
		s0s1buf[index] = byte(rand.Int() % 256)
	}

	keyOffset, digestOffset = getKeyDigetstOffset(s0s1buf[1:], scheme)
	s1p := make([]byte, RtmpHandShakePacketSize-RtmpDigestSize)
	copy(s1p, s0s1buf[1:digestOffset+1])
	copy(s1p[digestOffset:], s0s1buf[digestOffset+RtmpDigestSize+1:])
	h := hmac.New(sha256.New, GENUINE_FMS_KEY[:36])
	h.Write(s1p)
	s1digest := h.Sum(nil)
	copy(s0s1buf[1+digestOffset:], s1digest)

	count, err := r.io.Write(s0s1buf)
	if count != len(s0s1buf) || err != nil {
		log.Println("send s0s1 error")
		return errors.New("send s0s1 error")
	}

	//s2 content
	h2 := hmac.New(sha256.New, GENUINE_FMS_KEY[:68])
	h2.Write(digest)
	h2key := h2.Sum(nil)
	h22 := hmac.New(sha256.New, h2key)

	s2buf := make([]byte, RtmpHandShakePacketSize)
	for index := 0; index < RtmpHandShakePacketSize; index++ {
		s2buf[index] = byte(rand.Int() % 256)
	}
	h22.Write(s2buf[:RtmpHandShakePacketSize-RtmpDigestSize])
	s2digest := h22.Sum(nil)
	copy(s2buf[RtmpHandShakePacketSize-RtmpDigestSize:], s2digest)

	count, err = r.io.Write(s2buf)
	if count != len(s2buf) || err != nil {

		log.Println("send s2 error")
		return errors.New("send s2 error")
	}

	c2buf := make([]byte, RtmpHandShakePacketSize)
	count, err = r.io.Read(c2buf)
	if count != RtmpHandShakePacketSize || err != nil {
		log.Println("get c2 error")
		return errors.New("get c2 error")
	}

	//	hc2 := hmac.New(sha256.New, GENUINE_FP_KEY[:62])
	//	hc2.Write(s1digest)
	//	hc2key := hc2.Sum(nil)
	//	hc22 := hmac.New(sha256.New, hc2key)
	//	hc22.Write(c2buf[:1504])
	//	c2digest := hc22.Sum(nil)

	return nil
	//not worke fixme
	for index := 0; index < RtmpHandShakePacketSize; index++ {
		if c2buf[index] != s0s1buf[index+1] {
			log.Println("c2 iligle")
			return errors.New("c2 iligle")
		}
	}

	return nil
}

func checkClient(c1buf []uint8, keyOffset, digestOffset int) (bool, []byte) {
	if keyOffset < 0 || digestOffset < 0 {
		return false, nil
	}
	digest := make([]uint8, 32)
	key := make([]uint8, 128)
	copy(digest, c1buf[digestOffset:])
	copy(key, c1buf[keyOffset:])
	p := make([]uint8, RtmpHandShakePacketSize-32)
	copy(p[:digestOffset], c1buf[:digestOffset])
	copy(p[digestOffset:], c1buf[digestOffset+32:])

	h := hmac.New(sha256.New, GENUINE_FP_KEY[:30])
	h.Write(p)
	result := h.Sum(nil)
	for index := 0; index < 32; index++ {
		if result[index] != c1buf[digestOffset+index] {
			return false, nil
		}
	}
	return true, digest
}

func getKeyDigetstOffset(c1buf []uint8, scheme int) (keyOffset, digestOffset int) {
	keyOffset = -1
	digestOffset = -1

	if scheme == 1 {
		offset := int(c1buf[8]) + int(c1buf[9]) + int(c1buf[10]) + int(c1buf[11])
		offset = (offset % 728)
		digestOffset = offset + 8 + 4

		offset = int(c1buf[1532]) + int(c1buf[1533]) + int(c1buf[1534]) + int(c1buf[1535])
		offset = (offset % 632)
		keyOffset = offset + 772

	} else {
		offset := int(c1buf[772]) + int(c1buf[773]) + int(c1buf[774]) + int(c1buf[775])
		offset = (offset % 728)
		digestOffset = offset + 772 + 4

		offset = int(c1buf[768]) + int(c1buf[769]) + int(c1buf[770]) + int(c1buf[771])
		offset = (offset % 632)
		keyOffset = offset + 8

	}
	return
}
