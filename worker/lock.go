package worker

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/s1m0n21/go-crontab/common"
)

type JobLock struct {
	Kv    clientv3.KV
	Lease clientv3.Lease

	JobName string
	Cancel  context.CancelFunc
}

func InitJobLock(name string, kv clientv3.KV, lease clientv3.Lease) *JobLock {
	return &JobLock{
		Kv:      kv,
		Lease:   lease,
		JobName: name,
	}
}

func (jl *JobLock) TryLock() error {
	lease, err := jl.Lease.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.TODO())
	leaseID := lease.ID

	respChan, err := jl.Lease.KeepAlive(ctx, leaseID)
	if err != nil {
		cancel()
		jl.Lease.Revoke(context.TODO(), leaseID)
		return err
	}

	go func() {
		select {
		case resp := <-respChan:
			if resp == nil {
				break
			}
		}
	}()

	txn := jl.Kv.Txn(context.TODO(), leaseID)
	key := common.JOB_LOCK_PREFIX + jl.JobName

	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(key))

	return nil
}

func (jl *JobLock) Unlock() {

}
