package common

import "encoding/json"

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cron_expr"`
}

type HTTPResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewResponse(code int, msg string, data interface{}) ([]byte, error) {
	var r HTTPResponse

	r.Code = code
	r.Msg = msg
	r.Data = data

	resp, err := json.Marshal(r)

	return resp, err
}
