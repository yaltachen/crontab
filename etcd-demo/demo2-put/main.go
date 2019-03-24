package main

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	var (
		err     error
		config  clientv3.Config
		client  *clientv3.Client // 建立客户端
		kv      clientv3.KV      // 读写etcd键值对
		putResp *clientv3.PutResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		panic(err)
	}

	kv = clientv3.NewKV(client)

	if putResp, err = kv.Put(context.TODO(), "/cron/jobs/job2", "This is job2", clientv3.WithPrevKV()); err != nil {
		log.Printf("put <k: /cron/jobs/job1 v: hello> failed. error: %v\r\n", err)
	}

	if putResp.PrevKv != nil {
		log.Printf("PrevValue: %s", string(putResp.PrevKv.Value))
	}

	log.Println("Revision:", putResp.Header.Revision)
}
