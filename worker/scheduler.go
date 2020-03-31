package worker

import (
	"log"
	"time"

	"github.com/s1m0n21/go-crontab/common"
)

type scheduler struct {
	Event      chan *common.JobEvent
	Plan       map[string]*common.JobSchedulePlan
	Executing  map[string]*common.JobExecuteInfo
	ExecResult chan *common.JobExecuteResult
}

var Scheduler *scheduler

func (s *scheduler) TrySchedule() time.Duration {
	var near *time.Time
	now := time.Now()

	if len(s.Plan) == 0 {
		return 1 * time.Second
	}

	for _, plan := range s.Plan {
		if plan.Next.Before(now) || plan.Next.Equal(now) {
			// TODO: Execute command
			s.TryStartJob(plan)
			plan.Next = plan.Expr.Next(now)
		}

		if near == nil || plan.Next.Before(*near) {
			near = &plan.Next
		}

	}

	nextSchedule := (*near).Sub(now)

	return nextSchedule
}

func (s *scheduler) TryStartJob(plan *common.JobSchedulePlan) {
	if _, executing := s.Executing[plan.Job.Name]; executing {
		log.Printf("last task[%v] has not ended, skip\n", plan.Job.Name)
		return
	}

	execInfo := common.BuildJobExecuteInfo(plan)
	s.Executing[plan.Job.Name] = execInfo

	Executor.ExecuteJob(execInfo)
	//log.Printf("exec: %v | pt: %v | rt: %v\n", execInfo.Job.Name, execInfo.PlanTime, execInfo.RealTime)
}

func (s *scheduler) scheduleLoop() {
	next := s.TrySchedule()
	timer := time.NewTimer(next)

	for {
		select {
		case event := <-s.Event:
			s.handleJobEvent(event)
		case <-timer.C:
		case result := <-s.ExecResult:
			s.handleJobResult(result)
		}

		next = s.TrySchedule()
		timer.Reset(next)
	}
}

func (s *scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	s.Event <- jobEvent
}

func (s *scheduler) handleJobEvent(event *common.JobEvent) {
	switch event.Typ {
	case common.JOB_EVENT_SAVE:
		plan, err := common.BuildJobSchedulePlan(event.Job)
		if err != nil {
			return
		}
		s.Plan[event.Job.Name] = plan
	case common.JOB_EVENT_DELETE:
		if _, exist := s.Plan[event.Job.Name]; exist {
			delete(s.Plan, event.Job.Name)
		}
	case common.JOB_EVENT_KILL:
		if execInfo, executing := s.Executing[event.Job.Name]; executing {
			execInfo.Cancel()
		}
	}
}

func (s *scheduler) handleJobResult(result *common.JobExecuteResult) {
	if result.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		log := &common.Log{
			Name:         result.ExecInfo.Job.Name,
			Command:      result.ExecInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			log.Err = result.Err.Error()
		} else {
			log.Err = ""
		}
	}

	log.Printf("res: %v | st: %v | et: %v\n | err: %v", string(result.Output), result.StartTime, result.EndTime, result.Err)
	delete(s.Executing, result.ExecInfo.Job.Name)
}

func InitScheduler() error {
	Scheduler = &scheduler{
		Event:      make(chan *common.JobEvent, 1000),
		Plan:       make(map[string]*common.JobSchedulePlan),
		Executing:  make(map[string]*common.JobExecuteInfo),
		ExecResult: make(chan *common.JobExecuteResult, 1000),
	}

	go Scheduler.scheduleLoop()

	return nil
}

func (s *scheduler) PushJobResult(result *common.JobExecuteResult) {
	s.ExecResult <- result
}
