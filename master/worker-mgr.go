package master

import (
	"context"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/yaltachen/crontab/common"
	"go.etcd.io/etcd/clientv3"
)

type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
}

var G_workerMgr *WorkerMgr

func InitWorkerMgr() error {
	var (
		err    error
		cfg    clientv3.Config
		client *clientv3.Client
	)
	cfg = clientv3.Config{
		Endpoints:   G_cfg.EtcdEndPoints,
		DialTimeout: time.Duration(G_cfg.EtcdDialTimeOut) * time.Millisecond,
	}

	if client, err = clientv3.New(cfg); err != nil {
		return err
	}

	if G_workerMgr == nil {
		G_workerMgr = &WorkerMgr{
			client: client,
			kv:     clientv3.NewKV(client),
		}
	}

	return nil
}

// 从etcd中获取健康节点
func (w *WorkerMgr) GetOnlineWorkers() ([]*common.OnlineWorker, error) {
	var (
		workers     []*common.OnlineWorker
		worker      *common.OnlineWorker
		err         error
		getResponse *clientv3.GetResponse
		kvPair      *mvccpb.KeyValue
	)

	workers = make([]*common.OnlineWorker, 0)

	if getResponse, err = w.kv.Get(context.TODO(), common.WORKER_DIR, clientv3.WithPrefix()); err != nil {
		return nil, err
	}

	if getResponse.Count == 0 {
		return workers, nil
	}

	for _, kvPair = range getResponse.Kvs {
		worker = &common.OnlineWorker{
			WorkerIP: common.ExtractWorkerIp(string(kvPair.Key)),
		}
		workers = append(workers, worker)
	}

	return workers, nil
}
