package worker

import (
	"github.com/s1m0n21/go-crontab/common"
	"log"
	"time"
)

type scheduler struct {
	Event chan *common.JobEvent
	Plan  map[string]*common.JobSchedulePlan
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
			log.Println("exec ", plan.Job.Name)
			plan.Next = plan.Expr.Next(now)
		}

		if near == nil || plan.Next.Before(*near) {
			near = &plan.Next
		}

	}

	nextSchedule := (*near).Sub(now)

	return nextSchedule
}

func (s *scheduler) scheduleLoop() {
	next := s.TrySchedule()
	timer := time.NewTimer(next)

	for {
		select {
		case event := <-s.Event:
			s.handleJobEvent(event)
		case <-timer.C:

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
		_, exist := s.Plan[event.Job.Name]
		if exist {
			delete(s.Plan, event.Job.Name)
		}
	}
}

func InitScheduler() error {
	Scheduler = &scheduler{
		Event: make(chan *common.JobEvent, 1000),
		Plan:  make(map[string]*common.JobSchedulePlan),
	}

	go Scheduler.scheduleLoop()

	return nil
}
