package master

import (
	"net"
	"net/http"
	"strconv"
	"time"
)

var APISrv *API

type API struct {
	httpSrv *http.Server
}

func handleJobSave(w http.ResponseWriter, r *http.Request) {

}

func InitAPI() (err error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/job/save", handleJobSave)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(Config.API.ListenPort))
	if err != nil {
		return
	}

	httpSrv := http.Server{
		ReadTimeout:  time.Duration(Config.API.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(Config.API.WriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	APISrv = &API{
		httpSrv: &httpSrv,
	}

	go httpSrv.Serve(listener)

	return
}
