package backend

import (
	"container/list"
	"encoding/json"
	"errors"
	"events/eLiveListCtrl"
	"events/eRTMPEvent"
	"events/eStreamerEvent"
	"logger"
	"net/http"
	"strconv"
	"strings"
	"wssAPI"
)

type adminStreamManageHandler struct {
	Route string
}

type streamManageRequestData struct {
	Action Action
	LoginResponseData
}

func (asmh *adminStreamManageHandler) init(data *wssAPI.Msg) (err error) {
	asmh.Route = "/admin/stream/manage"
	return
}

func (asmh *adminStreamManageHandler) getRoute() (route string) {
	return asmh.Route
}

func (asmh *adminStreamManageHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == asmh.Route {
		asmh.handleStreamManageRequest(w, req)
	} else {
		badrequestResponse, err := BadRequest(WSS_SeverError, "server error in login")
		SendResponse(badrequestResponse, err, w)
	}
}

func (asmh *adminStreamManageHandler) handleStreamManageRequest(w http.ResponseWriter, req *http.Request) {
	if !LoginHandler.isLogin {
		doManage(w, req)
		//response, err := BadRequest(WSS_NotLogin, "please login")
		//SendResponse(response, err, w)
		return
	} else {
		requestData := getRequestData(req)
		if requestData.Action.ActionToken != LoginHandler.AuthToken {
			response, err := BadRequest(WSS_UserAuthError, "Auth error")
			SendResponse(response, err, w)
			return
		} else { //do manage
			doManage(w, req)
		}
	}
}

func getRequestData(req *http.Request) streamManageRequestData {
	data := streamManageRequestData{}
	code := req.PostFormValue("action_code")
	codeInt, err := strconv.Atoi(code)
	if err != nil {
		data.Action.ActionCode = codeInt
	} else {
		data.Action.ActionCode = -1
	}
	data.Action.ActionToken = req.PostFormValue("action_token")
	return data
}

func doManage(w http.ResponseWriter, req *http.Request) {
	err := parseActionEvent(w, req)
	if err != nil {
		logger.LOGI("parseActionEvent error ", err)
		sendBadResponse(w, "action code error", WSS_ParamError)
	}
}

func parseActionEvent(w http.ResponseWriter, req *http.Request) error {
	actionCode := req.PostFormValue("action_code")
	if len(actionCode) == 0 {
		return errors.New("no action code")
	}
	intCode, err := strconv.Atoi(actionCode)
	var task wssAPI.Task
	if err != nil {
		return errors.New("action code is error")
	}
	switch intCode {
	case WS_SHOW_ALL_STREAM:
		doShowAllStream(w)
	case WS_GET_LIVE_PLAYER_COUNT:
		doGetLivePlayerCount(w, req)
	case WS_ENABLE_BLACK_LIST:
		doEnableBlackList(w, req)
	case WS_SET_BLACK_LIST:
		doSetBlackList(w, req)
	case WS_ENABLE_WHITE_LIST:
		doEnableWhiteList(w, req)
	case WS_SET_WHITE_LIST:
		doSetWhiteList(w, req)
	case WS_SET_UP_STREAM_APP:
		doSetUpStreamApp(w, req)
	case WS_PULL_RTMP_STREAM:
		task = &eRTMPEvent.EvePullRTMPStream{}
	case WS_ADD_SINK:
		task = &eStreamerEvent.EveAddSink{}
	case WS_DEL_SINK:
		task = &eStreamerEvent.EveDelSink{}
	case WS_ADD_SOURCE:
		task = &eStreamerEvent.EveAddSource{}
	case WS_DEL_SOURCE:
		task = &eStreamerEvent.EveDelSource{}
	case WS_GET_SOURCE:
		task = &eStreamerEvent.EveGetSource{}
	default:
		return errors.New("no function")
	}

	if task == nil {

	}

	return nil
}

func doShowAllStream(w http.ResponseWriter) {
	eve := eLiveListCtrl.EveGetLiveList{}
	wssAPI.HandleTask(&eve)
	list := make([]object, 0)
	for item := eve.Lives.Front(); item != nil; item = item.Next() {
		info := item.Value.(*eLiveListCtrl.LiveInfo)
		list = append(list, *info)
	}
	sendSuccessResponse(nil, list, w)
}

// get one stream player
// exp: "ws:127.0.0.1:8080/ws/live/hks"
// 		"live_name should be "live/hks""
func doGetLivePlayerCount(w http.ResponseWriter, req *http.Request) {
	liveName := req.FormValue("live_name")
	if len(liveName) <= 0 {
		sendBadResponse(w, "need live_name", WSS_ParamError)
		return
	}
	eve := eLiveListCtrl.EveGetLivePlayerCount{}
	eve.LiveName = liveName
	wssAPI.HandleTask(&eve)
	sendSuccessResponse(eve, nil, w)
}

func doEnableBlackList(w http.ResponseWriter, r *http.Request) {
	enableBlackOrWhiteList(0, w, r)
}

func doSetBlackList(w http.ResponseWriter, req *http.Request) {
	setBlackOrWhiteList(0, w, req)
}

func doEnableWhiteList(w http.ResponseWriter, req *http.Request) {
	enableBlackOrWhiteList(1, w, req)
}

func doSetWhiteList(w http.ResponseWriter, r *http.Request) {
	setBlackOrWhiteList(1, w, r)
}

func doSetUpStreamApp(w http.ResponseWriter, r *http.Request) {
}

func doPullRtmpStream(w http.ResponseWriter, r *http.Request){
}

func doAddSink(w http.ResponseWriter, r *http.Request){
}

func doDelSink(w http.ResponseWriter, r *http.Request){
}

func doAddSource(w http.ResponseWriter, r *http.Request){
}

func doDelSource(w http.ResponseWriter, r *http.Request){
}

func doGetSource(w http.ResponseWriter, r *http.Request){
}


//Enable BlackList
// need form data " opcode = 1
// 					opcode 1 for enable blacklist
//		  			other for disable blacklist
//bwcode : 0 for black
//	  	 : 1 for white
func enableBlackOrWhiteList(bwcode int, w http.ResponseWriter, req *http.Request) {
	code := req.FormValue("opcode")
	eve := eLiveListCtrl.EveEnableBlackList{}

	if code == "1" {
		eve.Enable = true
	} else if code == "0" {
		eve.Enable = false
	} else {
		sendBadResponse(w, "opcode error", WSS_ParamError)
		return
	}

	var err error
	if bwcode == 0 {
		err = wssAPI.HandleTask(&eve)
		wssAPI.HandleTask(&eve)
	} else if bwcode == 1 {
		task := &eLiveListCtrl.EveEnableWhiteList{}
		task.Enable = eve.Enable
		err = wssAPI.HandleTask(task)
	} else {
		sendBadResponse(w, "bwcode error", WSS_SeverError)
		return
	}

	if err != nil {
		sendBadResponse(w, "error in service ", WSS_ParamError)
		return
	}
	sendSuccessResponse("op success", nil, w)
}

//setBlacklist
//need formdata " list=xxx|xxxx|xxxx&opcode=1
//				" | to split names
//				" opcode 1 for add 0 for del
//bwcode : 0 for black
//	  	 : 1 for white
func setBlackOrWhiteList(bwcode int, w http.ResponseWriter, req *http.Request) {
	str := req.FormValue("list")
	opcode := req.FormValue("opcode")
	if str == "" || opcode == "" {
		sendBadResponse(w, "need list data,it is look like 'list=xxx|xxx|xxx|xxx'", WSS_ParamError)
		return
	}

	eve := eLiveListCtrl.EveSetBlackList{}

	if opcode == "0" {
		eve.Add = false
	} else if opcode == "1" {
		eve.Add = true
	} else {
		sendBadResponse(w, "opcode error , 0 for del 1 for add", WSS_ParamError)
		return
	}

	banList := strings.Split(str, "|")
	eve.Names = list.New()
	for _, item := range banList {
		eve.Names.PushBack(item)
	}

	var err error
	if bwcode == 0 {
		err = wssAPI.HandleTask(&eve)
	} else if bwcode == 1 {
		task := &eLiveListCtrl.EveSetWhiteList{}
		task.Add = eve.Add
		task.Names = eve.Names
		err = wssAPI.HandleTask(task)
	} else {
		sendBadResponse(w, "bwcode error", WSS_SeverError)
		return
	}
	if err != nil {
		sendBadResponse(w, "error in service ", WSS_ParamError)
		return
	}

	sendSuccessResponse("op success", nil, w)
}

func sendSuccessResponse(data object, datas []object, w http.ResponseWriter) {
	responseData := ActionResponseData{}
	responseData.Code = WSS_RequestOK
	responseData.Msg = "ok"
	responseData.Data = data
	responseData.Datas = datas
	jbyte, err := json.Marshal(responseData)
	SendResponse(jbyte, err, w)
}

func sendBadResponse(w http.ResponseWriter, msg string, code int) {
	respnse, err := BadRequest(code, msg)
	SendResponse(respnse, err, w)
}
