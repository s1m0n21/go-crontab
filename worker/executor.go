package worker

import (
	"context"
	"github.com/s1m0n21/go-crontab/common"
	"os/exec"
	"time"
)

type executor struct {
}

var Executor *executor

func (e *executor) ExecuteJob(info *common.JobExecuteInfo) {

	go func() {
		result := &common.JobExecuteResult{
			ExecInfo:  info,
			Output:    make([]byte, 0),
			Err:       nil,
			StartTime: time.Now(),
		}

		lock := JobMgr.CreateLock(info.Job.Name)

		cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", info.Job.Command)
		output, err := cmd.CombinedOutput()

		result.EndTime = time.Now()
		result.Output = output
		result.Err = err

		Scheduler.PushJobResult(result)
	}()
}

func InitExecutor() error {
	Executor = &executor{}
	return nil
}
