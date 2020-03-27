package master

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/s1m0n21/go-crontab/common"
)

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

	jobKey := common.JOB_PREFIX + job.Name
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

func (jm *jobMgr) DeleteJob(name string) (*common.Job, error) {
	jobKey := common.JOB_PREFIX + name

	var oldJob *common.Job

	resp, err := jm.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	log.Println(resp.PrevKvs)

	if len(resp.PrevKvs) != 0 {
		if err := json.Unmarshal(resp.PrevKvs[0].Value, &oldJob); err != nil {
			log.Println(err.Error())
			return nil, nil
		}
		log.Println(oldJob)
	}

	return oldJob, nil
}

func (jm *jobMgr) ListJobs() ([]common.Job, error) {
	var job = common.Job{}
	var jobList = make([]common.Job, 0)

	resp, err := jm.kv.Get(context.TODO(), common.JOB_PREFIX, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kvPair := range resp.Kvs {
		if err := json.Unmarshal(kvPair.Value, &job); err != nil {
			log.Println(err.Error())
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}

	return jobList, err
}

func (jm *jobMgr) killJob(name string) error {
	key := common.JOB_KILLER_PREFIX + name

	lease, err := jm.lease.Grant(context.TODO(), 1)
	if err != nil {
		return err
	}

	_, err = jm.kv.Put(context.TODO(), key, "", clientv3.WithLease(lease.ID))

	return err
}
