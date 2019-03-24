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
		getResp *clientv3.GetResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		panic(err)
	}

	kv = clientv3.NewKV(client)

	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/job1"); err != nil {
		log.Printf("get key:/cron/jobs/job1 failed. error: %v\r\n", err)
	}

	log.Println(getResp.Kvs)
}
