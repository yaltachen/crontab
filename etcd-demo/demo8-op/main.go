package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	var (
		err    error
		cfg    clientv3.Config
		cli    *clientv3.Client
		kv     clientv3.KV
		opPut  clientv3.Op
		opGet  clientv3.Op
		opResp clientv3.OpResponse
	)

	cfg = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if cli, err = clientv3.New(cfg); err != nil {
		panic(err)
	}

	kv = clientv3.NewKV(cli)
	opPut = clientv3.OpPut("/cron/jobs/job4", "this is job4")

	if opResp, err = kv.Do(context.TODO(), opPut); err != nil {
		log.Printf("OpPut failed. Error: %v\r\n", err)
		return
	}
	fmt.Println("写入Revision", opResp.Put().Header.Revision)

	opGet = clientv3.OpGet("/cron/jobs/job4")
	if opResp, err = kv.Do(context.TODO(), opGet); err != nil {
		log.Printf("OpGet failed. Error: %v\r\n", err)
		return
	}
	fmt.Println("数据Revision", opResp.Get().Header.Revision)
	fmt.Println("数据value", string(opResp.Get().Kvs[0].Value))
}
