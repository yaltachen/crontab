package worker

import (
	"context"

	"github.com/yaltachen/crontab/common"
	"go.etcd.io/etcd/clientv3"
)

type JobLocker struct {
	JobName    string
	Kv         clientv3.KV
	Lease      clientv3.Lease
	LeaseID    clientv3.LeaseID
	CancelFunc func()
	IsLocked   bool
}

func NewJobLocker(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLocker) {
	return &JobLocker{
		JobName: jobName,
		Kv:      kv,
		Lease:   lease,
	}
}

func (j *JobLocker) TryLock() (err error) {
	var (
		cancelCtx          context.Context
		cancelFunc         func()
		leaseID            clientv3.LeaseID
		leaseGrantResp     *clientv3.LeaseGrantResponse
		leaseKeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
		txn                clientv3.Txn
		lockPath           string
		txnResponse        *clientv3.TxnResponse
	)
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	// 申请租约
	if leaseGrantResp, err = j.Lease.Grant(context.TODO(), 5); err != nil {
		goto FAIL
	}
	leaseID = leaseGrantResp.ID
	// 续租
	if leaseKeepAliveChan, err = j.Lease.KeepAlive(cancelCtx, leaseID); err != nil {
		goto FAIL
	}
	// 处理续租应答
	go func() {
		var (
			leaseKeepAliveResponse *clientv3.LeaseKeepAliveResponse
		)
	forloop:
		for {
			select {
			case leaseKeepAliveResponse = <-leaseKeepAliveChan:
				if leaseKeepAliveResponse == nil {
					// 租约到期
					break forloop
				} else {
					// 续租成功
				}
			}
		}
	}()

	// 锁路径
	lockPath = common.JOB_LOCK_DIR + j.JobName

	// 事务
	txn = j.Kv.Txn(context.TODO())
	if txnResponse, err = txn.If(clientv3.Compare(
		clientv3.CreateRevision(lockPath), "=", 0)).
		Then(clientv3.OpPut(lockPath, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(lockPath)).Commit(); err != nil {
		goto FAIL
	}

	if !txnResponse.Succeeded {
		err = common.ERR_KEY_ALREADY_REQUIRED
		goto FAIL
	}

	j.CancelFunc = cancelFunc
	j.LeaseID = leaseID
	j.IsLocked = true
	return nil

FAIL: // 立刻释放租约
	cancelFunc()
	j.Lease.Revoke(context.TODO(), leaseID)
	j.IsLocked = false
	return
}

func (j *JobLocker) Unlock() {
	if j.IsLocked {
		j.CancelFunc()
		j.Lease.Revoke(context.TODO(), j.LeaseID)
	}
}
