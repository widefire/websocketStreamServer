package webSocketService

import (
	"encoding/json"
	"errors"
	"logger"
	"wssAPI"
)

func (this *websocketHandler) ctrlPlay(data []byte) (err error) {
	st := &stPlay{}
	defer func() {
		if err != nil {
			logger.LOGE("play failed")
			err = this.sendWsStatus(this.conn, WS_status_error, NETSTREAM_PLAY_FAILED, st.Req)
		} else {
			this.lastCmd = WSC_play
		}
	}()
	err = json.Unmarshal(data, st)
	if err != nil {
		logger.LOGE("invalid params")
		return err
	}
	if false == supportNewCmd(this.lastCmd, WSC_play) {
		logger.LOGE("bad cmd")
		err = errors.New("bad cmd")
		return
	}
	//清除之前的

	switch this.lastCmd {
	case WSC_close:
		err = this.doPlay(st)
	case WSC_play:
		err = this.doClose()
		if err != nil {
			logger.LOGE("close failed ")
			return
		}
		err = this.doPlay(st)
	case WSC_play2:
		err = this.doClose()
		if err != nil {
			logger.LOGE("close failed ")
			return
		}
		err = this.doPlay(st)
	case WSC_pause:
		err = this.doClose()
		if err != nil {
			logger.LOGE("close failed ")
			return
		}
		err = this.doPlay(st)
	default:
		logger.LOGW("invalid last cmd")
		err = errors.New("invalid last cmd")
	}
	return
}

func (this *websocketHandler) ctrlPlay2(data []byte) (err error) {
	st := &stPlay2{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}

	return
}

func (this *websocketHandler) ctrlResume(data []byte) (err error) {

	st := &stResume{}
	defer func() {
		if err != nil {
			logger.LOGE("resume failed do nothing")
			this.sendWsStatus(this.conn, WS_status_status, NETSTREAM_FAILED, st.Req)

		} else {
			this.lastCmd = WSC_play
		}
	}()
	if false == supportNewCmd(this.lastCmd, WSC_resume) {
		logger.LOGE("bad cmd")
		err = errors.New("bad cmd")
		return
	}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	//only pase support resume
	switch this.lastCmd {
	case WSC_pause:
		err = this.doResume(st)
	default:
		err = errors.New("invalid last cmd")
		logger.LOGE(err.Error())
	}
	return
}

func (this *websocketHandler) ctrlPause(data []byte) (err error) {
	st := &stPause{}
	defer func() {
		if err != nil {
			logger.LOGE("pause failed")
			this.sendWsStatus(this.conn, WS_status_status, NETSTREAM_FAILED, st.Req)
		} else {
			this.lastCmd = WSC_pause
		}
	}()
	if false == supportNewCmd(this.lastCmd, WSC_pause) {
		logger.LOGE("bad cmd")
		err = errors.New("bad cmd")
		return
	}

	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	switch this.lastCmd {
	case WSC_play:
		this.doPause(st)
	case WSC_play2:
		this.doPause(st)
	default:
		err = errors.New("invalid last cmd in pause")
		logger.LOGE(err.Error())
	}

	return
}

func (this *websocketHandler) ctrlSeek(data []byte) (err error) {
	st := &stSeek{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlClose(data []byte) (err error) {
	st := &stClose{}
	defer func() {
		if err != nil {
			this.sendWsStatus(this.conn, WS_status_error, NETSTREAM_FAILED, st.Req)
		} else {

			this.lastCmd = WSC_close
		}
	}()
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	err = this.doClose()
	return
}

func (this *websocketHandler) ctrlStop(data []byte) (err error) {
	logger.LOGW("stop do the same as close now")
	st := &stStop{}
	err = json.Unmarshal(data, st)
	defer func() {
		if err != nil {
			this.sendWsStatus(this.conn, WS_status_error, NETSTREAM_FAILED, st.Req)
		} else {
			this.lastCmd = WSC_close
		}
	}()
	if err != nil {
		return err
	}
	err = this.doClose()
	return
}

func (this *websocketHandler) ctrlPublish(data []byte) (err error) {
	st := &stPublish{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	logger.LOGE("publish not coded")
	return
}

func (this *websocketHandler) ctrlOnMetadata(data []byte) (err error) {
	logger.LOGT(string(data))
	logger.LOGW("on metadata not processed")
	return
}

func (this *websocketHandler) doClose() (err error) {
	if this.isPlaying {
		this.stopPlay()
	}
	if this.hasSink {
		this.delSink(this.streamName, this.clientId)
	}
	if this.isPublish {
		this.stopPublish()
	}
	if this.hasSource {
		this.delSource(this.streamName, this.sourceIdx)
	}
	return
}

func (this *websocketHandler) doPlay(st *stPlay) (err error) {

	logger.LOGT("play")
	this.clientId = wssAPI.GenerateGUID()
	if len(this.app) > 0 {
		this.streamName = this.app + "/" + st.Name
	} else {
		this.streamName = st.Name
	}

	err = this.addSink(this.streamName, this.clientId, this)
	if err != nil {
		logger.LOGE("add sink failed: " + err.Error())
		return
	}

	err = this.sendWsStatus(this.conn, WS_status_status, NETSTREAM_PLAY_START, st.Req)
	return
}

func (this *websocketHandler) doPlay2() (err error) {
	logger.LOGW("play2 not coded")
	err = errors.New("not processed")
	return
}

func (this *websocketHandler) doResume(st *stResume) (err error) {
	logger.LOGT("resume play start")
	err = this.sendWsStatus(this.conn, WS_status_status, NETSTREAM_PLAY_START, st.Req)
	return
}

func (this *websocketHandler) doPause(st *stPause) (err error) {
	logger.LOGT("pause do nothing")

	this.sendWsStatus(this.conn, WS_status_status, NETSTREAM_PAUSE_NOTIFY, st.Req)
	return
}

func (this *websocketHandler) doSeek() (err error) {
	return
}

func (this *websocketHandler) doPublish() (err error) {
	return
}
