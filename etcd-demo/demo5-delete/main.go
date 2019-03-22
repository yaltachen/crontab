// delete
package main

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/coreos/etcd/clientv3"
)

func main() {
	var (
		err     error
		config  clientv3.Config
		client  *clientv3.Client // 建立客户端
		kv      clientv3.KV      // 读写etcd键值对
		delResp *clientv3.DeleteResponse
		kvpair  *mvccpb.KeyValue
	)
	config = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		panic(err)
	}

	kv = clientv3.NewKV(client)

	if delResp, err = kv.Delete(context.TODO(), "/cron/jobs/job2", clientv3.WithPrevKV()); err != nil {
		log.Printf("delete key:/cron/jobs/job2 failed. error: %v\r\n", err)
		return
	}

	if len(delResp.PrevKvs) != 0 {
		for _, kvpair = range delResp.PrevKvs {
			log.Println("删除了", string(kvpair.Key), string(kvpair.Value))
		}
	}
}
