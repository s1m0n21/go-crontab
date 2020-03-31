package worker

import (
	"math/rand"
	"os/exec"
	"time"

	"github.com/s1m0n21/go-crontab/common"
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

		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		err := lock.TryLock()
		defer lock.Unlock()

		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			result.StartTime = time.Now()

			cmd := exec.CommandContext(info.Ctx, "/bin/bash", "-c", info.Job.Command)
			output, err := cmd.CombinedOutput()

			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}

		Scheduler.PushJobResult(result)
	}()
}

func InitExecutor() error {
	Executor = &executor{}
	return nil
}
