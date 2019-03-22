package main

import (
	"time"

	"github.com/coreos/etcd/clientv3"
)

var (
	config clientv3.Config
	client *clientv3.Client
)

func init() {
	var err error

	config = clientv3.Config{
		Endpoints:   []string{"211.159.180.190:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		panic(err)
	}
}

func main() {

}
