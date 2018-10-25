package backend

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"wssAPI"
)

type adminLoginHandler struct {
	route     string
	AuthToken string
	isLogin   bool
}

type adminLoginData struct {
	username string
	password string
}

var LoginHandler *adminLoginHandler
var loginData *adminLoginData

func (alh *adminLoginHandler) init(data *wssAPI.Msg) (err error) {
	lgdata := data.Param1.(adminLoginData)
	if len(lgdata.username) == 0 || len(lgdata.password) == 0 {
		return errors.New("invalid param!")
	}
	loginData = &lgdata
	alh.route = "/admin/login"
	LoginHandler = alh
	return
}

func (alh *adminLoginHandler) getRoute() (route string) {
	return alh.route
}

func (alh *adminLoginHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == alh.route {
		alh.handleLoginRequest(w, req)
	} else {
		badrequestResponse, err := BadRequest(WSS_SeverError, "server error in login")
		SendResponse(badrequestResponse, err, w)
	}
}

func (alh *adminLoginHandler) handleLoginRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		result, err := BadRequest(WSS_RequestMethodError, "bad request in login ")
		SendResponse(result, err, w)
	} else {
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")
		if len(username) > 0 && len(password) > 0 {
			ispass, authToken := compAuth(username, password)
			if ispass {
				alh.isLogin = true
				alh.AuthToken = authToken
				responseData, err := passAuthResponseData(authToken)
				SendResponse(responseData, err, w)
			} else {
				responseData, err := BadRequest(WSS_UserAuthError, "login auth error")
				SendResponse(responseData, err, w)
			}
		} else {
			responseData, err := BadRequest(WSS_ParamError, "login auth error")
			SendResponse(responseData, err, w)
		}
	}
}

//login sucess response
func passAuthResponseData(authToken string) ([]byte, error) {
	result := &LoginResponseData{}
	result.Code = WSS_RequestOK
	result.Msg = "ok"
	result.Data.UserData.Token = authToken
	result.Data.UserData.Usrname = loginData.username

	resultData, err := json.Marshal(result)
	return resultData, err
}

func compAuth(username, password string) (ispass bool, authToken string) {
	if loginData.username == username && loginData.password == password {
		hash := md5.New()
		hash.Write([]byte(username + password))
		md5data := hash.Sum(nil)
		md5str := hex.EncodeToString(md5data)
		return true, md5str
	} else {
		return false, ""
	}
}
