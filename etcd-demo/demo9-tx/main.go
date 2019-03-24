package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

/*
* lease实现锁自动过期
* op操作
* txn事务：if else then
 */
func main() {
	var (
		err            error
		cfg            clientv3.Config
		cli            *clientv3.Client
		lease          clientv3.Lease
		leaseID        clientv3.LeaseID
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseKeepChan  <-chan *clientv3.LeaseKeepAliveResponse
		leaseKeepResp  *clientv3.LeaseKeepAliveResponse
		ctx            context.Context
		cancelFunc     func()
		txn            clientv3.Txn
		kv             clientv3.KV
		txnResp        *clientv3.TxnResponse
	)

	cfg = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if cli, err = clientv3.New(cfg); err != nil {
		panic(err)
	}

	lease = clientv3.NewLease(cli)
	kv = clientv3.NewKV(cli)
	txn = kv.Txn(context.TODO())

	if leaseGrantResp, err = lease.Grant(context.TODO(), 5); err != nil {
		log.Printf("grant failed. Error: %v\r\n", err)
		return
	}
	leaseID = leaseGrantResp.ID

	ctx, cancelFunc = context.WithCancel(context.TODO())

	defer func() {
		defer cancelFunc()
		defer lease.Revoke(context.TODO(), leaseID)
	}()

	if leaseKeepChan, err = lease.KeepAlive(ctx, leaseID); err != nil {
		log.Printf("keep alive failed. Error: %v\r\n", err)
		return
	}

	// 续租
	go func() {
	forloop:
		for {
			select {
			case leaseKeepResp = <-leaseKeepChan:
				if leaseKeepResp == nil {
					log.Println("停止续租")
					break forloop
				} else {
					log.Println("收到续租应答:", leaseKeepResp.ID)
				}
			}
		}
	}()

	if txnResp, err = txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/jobs/job5"), "=", 0)).
		Then(clientv3.OpPut("/cron/jobs/job5", "this is job5")).
		Else(clientv3.OpGet("/cron/jobs/job5")).Commit(); err != nil {
		log.Printf("txn commit failed. Error: %v\r\n", err)
	}

	if !txnResp.Succeeded {
		// 抢锁失败
		log.Println("锁被占用:", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	// 抢锁成功
	// 处理业务
	fmt.Println("开始处理任务")
	time.Sleep(5 * time.Second)
	fmt.Println("任务处理结束")
}
