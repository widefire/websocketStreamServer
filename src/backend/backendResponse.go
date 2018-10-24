package backend

import (
	"encoding/json"
	"net/http"
)

const (
	WSS_RequestMethodError = 1
	WSS_UserAuthError      = 101
	WSS_ParamError         = 102
	WSS_NotLogin           = 103
	WSS_RequestOK          = 200
	WSS_SeverHandleError   = 501
	WSS_SeverError         = 500
)

type BadRequestData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func SendResponse(responseData []byte, err error, w http.ResponseWriter) {
	if err != nil {
		w.Write([]byte("{\"code\":500,\"msg\":\"sevr error!\"}"))
	} else {
		w.Write(responseData)
	}
}

func BadRequest(code int, msg string) ([]byte, error) {
	result := &BadRequestData{}
	result.Code = code
	result.Msg = msg
	bytes, err := json.Marshal(result)
	return bytes, err
}
