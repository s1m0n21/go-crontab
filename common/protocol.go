package common

import (
	"context"
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

type JobExecuteInfo struct {
	Job      *Job
	PlanTime time.Time
	RealTime time.Time
	Ctx      context.Context
	Cancel   context.CancelFunc
}

type HTTPResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type JobExecuteResult struct {
	ExecInfo  *JobExecuteInfo
	Output    []byte
	Err       error
	StartTime time.Time
	EndTime   time.Time
}

type JobEvent struct {
	Typ int
	Job *Job
}

type Log struct {
	Name         string `json:"name" bson: "name"`
	Command      string `json:"command" bson: "command"`
	Output       string `json:"output" bson: "output"`
	Err          string `json:"err" bson: "err"`
	PlanTime     int64  `json:"plan_time" bson: "plan_time"`
	ScheduleTime int64  `json:"schedule_time" bson: "schedule_time"`
	StartTime    int64  `json:"start_time" bson: "start_time"`
	EndTime      int64  `json:"end_time" bson: "end_time"`
}

type LogBatch struct {
	Logs []interface{}
}

type LogFilter struct {
	Name string `bson: "name"`
}

type SortLogByStartTime struct {
	SortOrder int `bson: "start_time"`
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

func ExtractKillerName(key string) string {
	return strings.TrimPrefix(key, JOB_KILLER_PREFIX)
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

func BuildJobExecuteInfo(plan *JobSchedulePlan) *JobExecuteInfo {
	ctx, cancel := context.WithCancel(context.TODO())

	return &JobExecuteInfo{
		Job:      plan.Job,
		PlanTime: plan.Next,
		RealTime: time.Now(),
		Ctx:      ctx,
		Cancel:   cancel,
	}
}
