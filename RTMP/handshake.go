package RTMP

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"log"
	"math/rand"

	"github.com/widefire/websocketStreamServer/core"
)

const (
	RTMP_HANDSHAKE_PACKET_SIZE = 1536
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

func (r *RTMPFormat) handShake() error {
	if r.isInput {
		return r.handShakeAsServer()
	}
	return r.handShakeAsClient()
}

func (r *RTMPFormat) handShakeAsClient() error {
	return errors.New("NOT IMPLETE")
}

func (r *RTMPFormat) handShakeAsServer() error {
	buf := make([]uint8, RTMP_HANDSHAKE_PACKET_SIZE)
	//s1buf := make([]uint8, RTMP_HANDSHAKE_PACKET_SIZE)
	c1buf := make([]uint8, RTMP_HANDSHAKE_PACKET_SIZE)
	_, err := r.io.Read(buf[:1])
	if err != nil {
		return err
	}
	// c0
	if buf[0] != 3 {
		log.Println("RTMP version mismatch")
		return errors.New("RTMP version mismatch")
	}
	// s1
	count, err := r.io.Read(c1buf)
	if count != RTMP_HANDSHAKE_PACKET_SIZE {
		return errors.New("read s1 error")
	}

	var time uint32
	var zero uint32

	memio := core.NewWSSMEM(c1buf)

	core.WSSIOReadB(memio, &time)

	core.WSSIOReadB(memio, &zero)

	if zero != 0 {
	} else {
		return r.sampleHandleShake(c1buf)
	}

	return errors.New("NOT IMPLETE")
}

func (r *RTMPFormat) sampleHandleShake(c1buf []uint8) error {
	s0s1buf := make([]uint8, RTMP_HANDSHAKE_PACKET_SIZE+1)
	s0s1buf[0] = 0x03
	for i := 9; i < RTMP_HANDSHAKE_PACKET_SIZE+1; i++ {
		s0s1buf[i] = uint8(rand.Int() % 256)
	}

	count, err := r.io.Write(s0s1buf)
	if err != nil {
		return err
	}
	if count != len(s0s1buf) {
		return errors.New("send s0 s1 error")
	}

	count, err = r.io.Write(c1buf)
	if err != nil {
		return err
	}
	if count != len(s0s1buf) {
		return errors.New("send s2 error")
	}

	c2buf := make([]uint8, RTMP_HANDSHAKE_PACKET_SIZE)
	count, err = r.io.Read(c2buf)
	if err != nil {
		return err
	}
	if count != len(c2buf) {
		return errors.New("recive c2 error")
	}
	for index := 0; index < RTMP_HANDSHAKE_PACKET_SIZE; index++ {
		if c2buf[index] != s0s1buf[index+1] {
			return errors.New(" c2buf != s1buf")
		}
	}
	return nil
}

func (r *RTMPFormat) complexHandleShake(c1buf []uint8) error {
	keyOffset, digestOffset := getKeyDigetstOffset(c1buf, 0)
	ok := checkClient(c1buf, keyOffset, digestOffset)
	if !ok {
		keyOffset, digestOffset = getKeyDigetstOffset(c1buf, 1)
		ok = checkClient(c1buf, keyOffset, digestOffset)
		if !ok {
			return errors.New("handle shake check c1 error")
		}
	}

}

func generateS1Buf(c1buf []uint8, s1buf []uint8, keyOffset int) error {
	copy(s1, c1buf[keyOffset:])
}

func checkClient(c1buf []uint8, keyOffset, digestOffset int) bool {
	digest := make([]uint8, 32)
	key := make([]uint8, 128)
	copy(digest, c1buf[digestOffset:])
	copy(key, c1buf[keyOffset:])
	p := make([]uint8, RTMP_HANDSHAKE_PACKET_SIZE-32)
	copy(p[:digestOffset], c1buf[:digestOffset])
	copy(p[digestOffset:], c1buf[digestOffset+32:])

	h := hmac.New(sha256.New, GENUINE_FP_KEY[:30])
	h.Write(p)
	result := h.Sum(nil)
	for index := 0; index < 32; index++ {
		if result[index] != c1buf[digestOffset+index] {
			return false
		}
	}
	return true
}

func getKeyDigetstOffset(c1buf []uint8, scheme int) (keyOffset, digestOffset int) {
	keyOffset = -1
	digestOffset = -1

	if scheme == 1 {
		offset := int(c1buf[8]) + int(c1buf[9]) + int(c1buf[10]) + int(c1buf[11])
		if offset > 728 {
			return
		}
		digestOffset = offset + 8 + 4

		offset = int(c1buf[1532]) + int(c1buf[1533]) + int(c1buf[1534]) + int(c1buf[1535])
		if offset > 632 {
			return
		}
		keyOffset = offset + 772 + 4

	} else {
		offset := int(c1buf[772]) + int(c1buf[773]) + int(c1buf[774]) + int(c1buf[775])
		if offset > 728 {
			return
		}
		digestOffset = offset + 772

		offset = int(c1buf[768]) + int(c1buf[769]) + int(c1buf[770]) + int(c1buf[771])
		if offset > 632 {
			return
		}
		keyOffset = offset + 8

	}
	return
}
