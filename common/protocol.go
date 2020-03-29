package common

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cron_expr"`
}

type JobSchedulePlan struct {
	Job  *Job
	Expr *cronexpr.Expression
	Next time.Time
}

type HTTPResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type JobEvent struct {
	Typ int
	Job *Job
}

func NewResponse(code int, msg string, data interface{}) ([]byte, error) {
	var r HTTPResponse

	r.Code = code
	r.Msg = msg
	r.Data = data

	resp, err := json.Marshal(r)

	return resp, err
}

func UnpackJob(val []byte) (*Job, error) {
	job := &Job{}

	if err := json.Unmarshal(val, job); err != nil {
		return nil, err
	}

	return job, nil
}

func ExtractJobName(key string) string {
	return strings.TrimPrefix(key, JOB_PREFIX)
}

func BuildJobEvent(typ int, job *Job) *JobEvent {
	return &JobEvent{
		Typ: typ,
		Job: job,
	}
}

func BuildJobSchedulePlan(job *Job) (*JobSchedulePlan, error) {
	expr, err := cronexpr.Parse(job.CronExpr)
	if err != nil {
		return nil, err
	}

	plan := &JobSchedulePlan{
		Job:  job,
		Expr: expr,
		Next: expr.Next(time.Now()),
	}

	return plan, nil
}
