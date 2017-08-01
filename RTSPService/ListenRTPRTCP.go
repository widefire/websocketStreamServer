//监听RTP RTCP
package RTSPService

import (
	"logger"
	"net"
	"sync"
)

//这两个函数 直到track里面的svr conn关闭后，自动退出？
func listenRTP(track *trackInfo, handler *RTSPHandler, wait *sync.WaitGroup, chAddr chan *net.UDPAddr) {
	defer func() {
		wait.Done()
	}()
	mtu := 1500
	data := make([]byte, mtu)
	needAddr := true
	for handler.isPlaying {
		ret, addr, err := track.RTPSvrConn.ReadFromUDP(data)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		if needAddr {
			needAddr = false
			chAddr <- addr
		}
		logger.LOGD(data)
		logger.LOGD(ret, addr)
	}
}

func listenRTCP(track *trackInfo, handler *RTSPHandler, wait *sync.WaitGroup, chAddr chan *net.UDPAddr) {
	defer func() {
		wait.Done()
	}()
	mtu := 1500
	data := make([]byte, mtu)
	needAddr := true
	for handler.isPlaying {
		ret, addr, err := track.RTCPSvrConn.ReadFromUDP(data)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		if needAddr {
			needAddr = false
			chAddr <- addr
		}
		//处理RTCP信息
		//logger.LOGD(data)
		rtcpPkt, err := parseRTCP(data[0:ret])
		if err != nil {
			logger.LOGE(err.Error())
			continue
		}
		if nil == rtcpPkt {
			continue
		}

	}
}
