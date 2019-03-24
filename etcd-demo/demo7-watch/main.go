package main

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

const (
	QUIT = "QUIT"
	RUN  = "RUN"
)

func main() {
	var (
		err         error
		config      clientv3.Config
		client      *clientv3.Client
		kv          clientv3.KV
		ticker      *time.Ticker
		control     chan string
		flag        string
		watchID     int64
		watcher     clientv3.Watcher
		watcherChan <-chan clientv3.WatchResponse
		watcherResp clientv3.WatchResponse
		getResp     *clientv3.GetResponse
		event       *clientv3.Event
	)

	control = make(chan string, 1)

	config = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		panic(err)
	}

	kv = clientv3.NewKV(client)
	ticker = time.NewTicker(1 * time.Second)
	watcher = clientv3.NewWatcher(client)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	watcherChan = watcher.Watch(ctx, "/cron/jobs/job3", clientv3.WithRev(watchID))

	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/job3"); err != nil {
		log.Printf("get failed. Error: %v\r\n", err)
	}

	watchID = getResp.Header.Revision
	go func() {
		for {
			select {
			case <-ticker.C:
				kv.Put(context.TODO(), "/cron/jobs/job3", "this is job3")
				kv.Delete(context.TODO(), "/cron/jobs/job3")
			case flag = <-control:
				if flag == QUIT {
					log.Println("stop updating")
					goto END
				}
			}
		}
	END:
	}()
	go func() {
		time.AfterFunc(8*time.Second, func() {
			control <- QUIT
		})
		time.AfterFunc(5*time.Second, func() {
			cancelFunc()
		})
	}()
	go func() {
		// for watcherResp = range watcherChan {
		// 	for _, event = range watcherResp.Events {
		// 		switch event.Type {
		// 		case mvccpb.PUT:
		// 			log.Println("修改", string(event.Kv.Value), event.Kv.CreateRevision, event.Kv.ModRevision)
		// 		case mvccpb.DELETE:
		// 			log.Println("删除", event.Kv.ModRevision)
		// 		}
		// 	}
		// }
		// log.Println("stop watching")
	FORLOOP:
		for {
			select {
			case watcherResp = <-watcherChan:
				if len(watcherResp.Events) == 0 {
					log.Println("stop watching")
					break FORLOOP
				}
				for _, event = range watcherResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						log.Println("修改", string(event.Kv.Value), event.Kv.CreateRevision, event.Kv.ModRevision)
					case mvccpb.DELETE:
						log.Println("删除", event.Kv.ModRevision)
					}
				}
			}
		}
	}()
	time.Sleep(10 * time.Second)
}
