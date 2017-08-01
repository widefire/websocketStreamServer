package HTTPMUX

import (
	"net/http"
	"logger"
)

var ports map[string]*http.ServeMux

func init(){
	ports=make(map[string]*http.ServeMux)
}


func AddRoute(port,route string,handler func(w http.ResponseWriter, req *http.Request))  {
	mux,exist:=ports[port]
	if false==exist{
		ports[port]=http.NewServeMux()
		mux=ports[port]
	}
	mux.HandleFunc(route,handler)
}


func Start()  {
	for k,v:=range ports  {
		go func(addr string,handler http.Handler){
			err:=http.ListenAndServe(addr,handler)
			if err!=nil{
				logger.LOGE(err.Error())
			}
		}(k,v)
	}
}

