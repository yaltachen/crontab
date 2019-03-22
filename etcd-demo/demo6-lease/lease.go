package main

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

func main() {
	var (
		err            error
		cfg            clientv3.Config
		cli            *clientv3.Client
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseID        clientv3.LeaseID
		kv             clientv3.KV
		getResp        *clientv3.GetResponse
		ticker         *time.Ticker
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		keepResp       *clientv3.LeaseKeepAliveResponse
	)
	cfg = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if cli, err = clientv3.New(cfg); err != nil {
		panic(err)
	}

	kv = clientv3.NewKV(cli)
	lease = clientv3.NewLease(cli)

	if leaseGrantResp, err = lease.Grant(context.TODO(), 10); err != nil {
		log.Printf("grant failed. Error: %v\r\n", err)
		return
	}

	leaseID = leaseGrantResp.ID
	fiveSecond, _ := context.WithTimeout(context.TODO(), 5*time.Second)

	if keepRespChan, err = lease.KeepAlive(fiveSecond, leaseID); err != nil {
		log.Printf("keepAlive failed. Error: %v\r\n", err)
		return
	}

	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					log.Println("续租已经失效")
					goto END
				} else {
					// 每秒续租一次
					log.Println("收到自动续租应答：", keepResp.ID)
				}
			}
		}
	END:
	}()

	if _, err = kv.Put(context.TODO(), "/cron/lock/job1", "this is job1", clientv3.WithLease(leaseID)); err != nil {
		log.Printf("put failed. Error: %v\r\n", err)
		return
	}

	log.Println("写入成功")

	ticker = time.NewTicker(1 * time.Second)

forloop:
	for {
		select {
		case <-ticker.C:
			if getResp, err = kv.Get(context.TODO(), "/cron/lock/job1", clientv3.WithPrevKV()); err != nil {
				log.Printf("get failed. Error: %v\r\n", err)
			}
			if getResp.Count == 0 {
				break forloop
			} else {
				log.Printf("get k-v: %v", getResp.Kvs)
			}
		}
	}

}
