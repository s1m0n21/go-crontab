package master

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/s1m0n21/go-crontab/common"
)

var APISrv *API

type API struct {
	httpSrv *http.Server
}

func handleJobSave(w http.ResponseWriter, r *http.Request) {
	var job *common.Job

	if err := r.ParseForm(); err != nil {
		log.Println(err.Error())
		resp, _ := common.NewResponse(-1, err.Error(), nil)
		w.Write(resp)
		return
	}

	jobRaw := r.PostForm.Get("job")

	if err := json.Unmarshal([]byte(jobRaw), &job); err != nil {
		log.Println(err.Error())
		resp, _ := common.NewResponse(-1, err.Error(), nil)
		w.Write(resp)
		return
	}

	oldJob, err := JobMgr.SaveJob(job)
	if err != nil {
		log.Println(err.Error())
		resp, _ := common.NewResponse(-1, err.Error(), nil)
		w.Write(resp)
		return
	}

	resp, err := common.NewResponse(0, "success", oldJob)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(resp)

	return
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
