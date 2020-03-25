package master

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/s1m0n21/go-crontab/common"
)

const jobPrefix string = "/cron/jobs/"

var JobMgr *jobMgr

type jobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

func InitJobMgr() (err error) {
	config := clientv3.Config{
		Endpoints:   Config.ETCD.Endpoints,
		DialTimeout: time.Duration(Config.ETCD.DialTimeout) * time.Millisecond,
	}

	client, err := clientv3.New(config)
	if err != nil {
		return
	}

	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	JobMgr = &jobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}

	return
}

func (jm *jobMgr) SaveJob(job *common.Job) (*common.Job, error) {
	var oldJob *common.Job

	jobKey := jobPrefix + job.Name
	jobValue, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	resp, err := jm.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	if resp.PrevKv != nil {
		if err := json.Unmarshal(resp.PrevKv.Value, &oldJob); err != nil {
			log.Println(err.Error())
			return nil, nil
		}
	}

	return oldJob, nil
}
