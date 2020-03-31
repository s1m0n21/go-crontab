package worker

import (
	"context"

	"github.com/s1m0n21/go-crontab/common"

	"github.com/coreos/etcd/clientv3"
)

type JobLock struct {
	Kv    clientv3.KV
	Lease clientv3.Lease

	JobName  string
	Cancel   context.CancelFunc
	LeaseID  clientv3.LeaseID
	IsLocked bool
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
		for {
			select {
			case resp := <-respChan:
				if resp == nil {
					break
				}
			}
		}
	}()

	txn := jl.Kv.Txn(context.TODO())
	key := common.JOB_LOCK_PREFIX + jl.JobName

	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(key))

	resp, err := txn.Commit()
	if err != nil {
		cancel()
		jl.Lease.Revoke(context.TODO(), leaseID)
		return err
	}

	if !resp.Succeeded {
		cancel()
		jl.Lease.Revoke(context.TODO(), leaseID)
		return common.ERR_LOCK_ALREADY_REQUIRED
	}

	jl.LeaseID = leaseID
	jl.Cancel = cancel
	jl.IsLocked = true

	return nil
}

func (jl *JobLock) Unlock() {
	if jl.IsLocked {
		jl.Cancel()
		jl.Lease.Revoke(context.TODO(), jl.LeaseID)
	}
}
