package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/yaltachen/crontab/common"
	"go.etcd.io/etcd/clientv3"
)

type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var G_jobMgr *JobMgr

func InitJobMgr() (err error) {
	var (
		cfg     clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		watcher clientv3.Watcher
		lease   clientv3.Lease
	)
	cfg = clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(cfg); err != nil {
		return err
	}

	kv = clientv3.NewKV(client)
	watcher = clientv3.NewWatcher(client)
	lease = clientv3.NewLease(client)

	if G_jobMgr == nil {
		G_jobMgr = &JobMgr{
			client:  client,
			kv:      kv,
			lease:   lease,
			watcher: watcher,
		}
	}

	G_jobMgr.WatchJobs()
	G_jobMgr.WatchKiller()
	return nil
}

func (jm *JobMgr) WatchJobs() (err error) {
	// 1.获取Jobs
	// 2.从当前Revision向后监听变化
	var (
		getResponse        *clientv3.GetResponse
		job                *common.Job
		kvPair             *mvccpb.KeyValue
		watchStartRevision int64
		watchResp          clientv3.WatchResponse
		watchChan          <-chan clientv3.WatchResponse
		event              *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
	)

	if getResponse, err = jm.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return err
	}

	for _, kvPair = range getResponse.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		} else {
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			G_scheduler.PushJobEvent(jobEvent)
		}
	}

	watchStartRevision = getResponse.Header.Revision + 1

	go func() {
		watchChan = jm.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())

	forloop:
		for {
			select {
			case watchResp = <-watchChan:

				if len(watchResp.Events) == 0 {
					break forloop
				}

				for _, event = range watchResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						if job, err = common.UnpackJob(event.Kv.Value); err != nil {
							continue
						}
						// 构建JobEvent
						jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
						// log.Println("put", string(event.Kv.Value))
					case mvccpb.DELETE:
						jobName = common.ExtractJobName(string(event.Kv.Key))
						job = &common.Job{Name: jobName}
						jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
						// log.Println("delete", string(event.Kv.Key))
					}
					// 推送jobEvent给调度器
					// G_Sechduler.PushJobEvent
					G_scheduler.PushJobEvent(jobEvent)
				}

			}
		}

	}()

	return
}

func (jm *JobMgr) NewLocker(jobName string) (jobLocker *JobLocker) {
	return &JobLocker{
		JobName:  jobName,
		Kv:       G_jobMgr.kv,
		Lease:    G_jobMgr.lease,
		IsLocked: false,
	}
}

func (jm *JobMgr) WatchKiller() {
	go func() {
		var (
			watchChan     <-chan clientv3.WatchResponse
			watchResponse clientv3.WatchResponse
			event         *clientv3.Event
			jobName       string
			job           *common.Job
			jobEvent      *common.JobEvent
		)
		watchChan = jm.watcher.Watch(context.TODO(), common.JOB_KILL_DIR, clientv3.WithPrefix())

	watchLoop:
		for {
			select {
			case watchResponse = <-watchChan:
				if len(watchResponse.Events) == 0 {
					// watchChan 关闭
					break watchLoop
				}
				for _, event = range watchResponse.Events {
					switch event.Type {
					case mvccpb.PUT:
						// 杀死任务
						jobName = common.ExtractKillerName(string(event.Kv.Key))
						job = &common.Job{Name: jobName}
						jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
						// 事件推给scheduler
						G_scheduler.PushJobEvent(jobEvent)
					}
				}
			}
		}
	}()
	return
}
