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

	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/", clientv3.WithPrefix()); err != nil {
		log.Printf("put <k: /cron/jobs/job1 v: hello> failed. error: %v\r\n", err)
		return
	} else {
		log.Println(getResp.Kvs)
	}

}
