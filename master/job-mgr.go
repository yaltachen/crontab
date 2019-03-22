package master

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/coreos/etcd/clientv3"
	"github.com/yaltachen/crontab/common"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	G_jobMgr *JobMgr
)

func InitJobMgr() error {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		err    error
	)
	config = clientv3.Config{
		Endpoints:   G_cfg.EtcdEndPoints,
		DialTimeout: time.Duration(G_cfg.EtcdDialTimeOut) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return err
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	if G_jobMgr == nil {
		G_jobMgr = &JobMgr{
			client: client,
			kv:     kv,
			lease:  lease,
		}
	}

	return nil
}

// save job
func (jm *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {

	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
	)

	jobKey = common.JOB_SAVE_DIR + job.Name

	if jobValue, err = json.Marshal(job); err != nil {
		return nil, err
	}

	if putResp, err = jm.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return nil, err
	}

	// 如果是更新
	if putResp.PrevKv != nil {
		// 更新
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}

	return
}

// del job
func (jm *JobMgr) DeleteJob(jobName string) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		oldJobObj common.Job
		delResp   *clientv3.DeleteResponse
	)
	jobKey = common.JOB_SAVE_DIR + jobName

	if delResp, err = jm.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return nil, err
	}

	if len(delResp.PrevKvs) == 1 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
		}
		oldJob = &oldJobObj
	}

	return
}

// list job
func (jm *JobMgr) ListJobs() (jobs []*common.Job, err error) {
	var (
		job     *common.Job
		getResp *clientv3.GetResponse
		kvpair  *mvccpb.KeyValue
	)

	if getResp, err = jm.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return nil, err
	}

	jobs = make([]*common.Job, 0)

	for _, kvpair = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvpair.Value, job); err != nil {
			log.Printf("unmarshal job: %s from etcd failed. Error: %v\r\n", string(kvpair.Value), err)
			err = nil
			continue
		} else {
			jobs = append(jobs, job)
			log.Printf("job: %v\r\n", job)
		}
	}

	return
}

// kill job
func (jm *JobMgr) KillJob(jobName string) (err error) {
	var (
		killerKey      string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseID        clientv3.LeaseID
	)
	killerKey = common.JOB_KILL_DIR + jobName

	// 让worker监听到一次put操作，创建一个租约让其稍后自动过期
	if leaseGrantResp, err = jm.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	// 租约ID
	leaseID = leaseGrantResp.ID

	// 设置killer标记
	if _, err = jm.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseID)); err != nil {
		return
	}

	return
}
