package worker

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/yaltachen/crontab/common"
	"go.etcd.io/etcd/clientv3"
)

type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIP string
}

var (
	G_register *Register
)

func InitRegister() error {
	var (
		err     error
		client  *clientv3.Client
		config  clientv3.Config
		localIP string
	)

	if G_register != nil {
		return nil
	}

	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return err
	}

	if localIP, err = getLocalIP(); err != nil {
		return err
	}

	G_register = &Register{
		client:  client,
		kv:      clientv3.NewKV(client),
		lease:   clientv3.NewLease(client),
		localIP: localIP,
	}

	go G_register.keepOnline()

	return nil
}

func getLocalIP() (string, error) {
	var (
		ipv4 string
		err  error

		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
	)
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return "", err
	}

	for _, addr = range addrs {
		// ipv4 OR ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return ipv4, nil
			}
		}
	}

	return "", common.ERR_NO_LOCAL_IP_FOUND
}

// 注册到/cron/workers/{ip}，并自动续租
func (r *Register) keepOnline() {
	var (
		regKey                 string
		leaseID                clientv3.LeaseID
		leaseGrantResp         *clientv3.LeaseGrantResponse
		leaseKeepAliveChan     <-chan *clientv3.LeaseKeepAliveResponse
		leaseKeepAliveResponse *clientv3.LeaseKeepAliveResponse
		cancelCtx              context.Context
		cancelFunc             func()
		err                    error
	)

	for {

		regKey = common.WORKER_DIR + r.localIP

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())
		// 创建租约
		if leaseGrantResp, err = r.lease.Grant(cancelCtx, 10); err != nil {
			goto RETRY
		}
		leaseID = leaseGrantResp.ID

		// 自动续租
		if leaseKeepAliveChan, err = r.lease.KeepAlive(cancelCtx, leaseID); err != nil {
			goto RETRY
		}

		// 注册到etcd
		if _, err = r.kv.Put(context.TODO(), regKey, "", clientv3.WithLease(leaseID)); err != nil {
			// 取消租约
			cancelFunc()
			goto RETRY
		}

		log.Printf("主机: %s,注册成功\r\n", r.localIP)

		// 处理续租应答
		for {
			select {
			case leaseKeepAliveResponse = <-leaseKeepAliveChan:
				if leaseKeepAliveResponse == nil {
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
	}
}
