package worker

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/s1m0n21/go-crontab/common"

	"github.com/coreos/etcd/mvcc/mvccpb"
)

var JobMgr *jobMgr

type jobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

func InitJobMgr() error {
	config := clientv3.Config{
		Endpoints:   Config.ETCD.Endpoints,
		DialTimeout: time.Duration(Config.ETCD.DialTimeout) * time.Millisecond,
	}

	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watcher := clientv3.NewWatcher(client)

	JobMgr = &jobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	if err = JobMgr.WatchJobs(); err != nil {
		return err
	}

	if err = JobMgr.watchKiller(); err != nil {
		return err
	}

	return nil
}

func (jm *jobMgr) WatchJobs() error {
	resp, err := jm.kv.Get(context.TODO(), common.JOB_PREFIX, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kvPair := range resp.Kvs {
		if job, err := common.UnpackJob(kvPair.Value); err == nil {
			jobEvent := common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			// TODO: schedule job
			log.Println(jobEvent)
			Scheduler.PushJobEvent(jobEvent)
		}
	}

	go func() {
		startRev := resp.Header.Revision + 1

		watchChan := JobMgr.watcher.Watch(context.TODO(), common.JOB_PREFIX, clientv3.WithRev(startRev), clientv3.WithPrefix())

		for resp := range watchChan {
			for _, event := range resp.Events {
				var jobEvent *common.JobEvent

				switch event.Type {
				case mvccpb.PUT:
					job, _ := common.UnpackJob(event.Kv.Value)
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE:
					jobName := common.ExtractJobName(string(event.Kv.Key))
					job := &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)

				}
				// TODO: schedule job
				log.Println(jobEvent)
				Scheduler.PushJobEvent(jobEvent)
			}
		}
	}()

	return nil
}

func (jm *jobMgr) CreateLock(name string) *JobLock {
	return InitJobLock(name, jm.kv, jm.lease)
}

func (jm jobMgr) watchKiller() error {
	go func() {
		watchChan := JobMgr.watcher.Watch(context.TODO(), common.JOB_KILLER_PREFIX, clientv3.WithPrefix())

		for resp := range watchChan {
			var jobEvent *common.JobEvent

			for _, event := range resp.Events {
				switch event.Type {
				case mvccpb.PUT:
					name := common.ExtractKillerName(string(event.Kv.Key))
					job := &common.Job{Name: name}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					Scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:

				}

			}
		}
	}()

	return nil
}
