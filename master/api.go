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

func handleJobDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err.Error())
		resp, _ := common.NewResponse(-1, err.Error(), nil)
		w.Write(resp)
		return
	}

	name := r.PostForm.Get("name")

	if len(name) == 0 {
		resp, _ := common.NewResponse(-1, "invalid param", nil)
		w.Write(resp)
		return
	}

	log.Println(name)

	oldJob, err := JobMgr.DeleteJob(name)
	if err != nil {
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

func handleJobList(w http.ResponseWriter, r *http.Request) {
	jobList, err := JobMgr.ListJobs()
	if err != nil {
		resp, _ := common.NewResponse(-1, err.Error(), nil)
		w.Write(resp)
		return
	}

	resp, _ := common.NewResponse(0, "success", jobList)
	w.Write(resp)

	return
}

func handleJobKill(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		resp, _ := common.NewResponse(-1, err.Error(), nil)
		w.Write(resp)
		return
	}

	name := r.PostForm.Get("name")

	if err := JobMgr.killJob(name); err != nil {
		resp, _ := common.NewResponse(-1, err.Error(), nil)
		w.Write(resp)
		return
	}

	resp, _ := common.NewResponse(0, "success", nil)
	w.Write(resp)

	return
}

func InitAPI() (err error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)

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
