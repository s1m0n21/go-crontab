package master

import (
	"time"

	"github.com/coreos/etcd/clientv3"
)

type jobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var JobMgr *jobMgr

func InitJobMgr() (err error) {
	config := clientv3.Config{
		Endpoints:   Config.ETCD.Endpoints,
		DialTimeout: time.Duration(Config.ETCD.DialTimeout) * time.Microsecond,
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
