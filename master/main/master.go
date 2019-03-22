package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/yaltachen/crontab/master"
)

var (
	cfgPath string
)

func InitEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitArgs() {
	flag.StringVar(&cfgPath, "config", "master.json", "指定master.json")
	flag.Parse()
}

func main() {
	var (
		running chan int
		err     error
	)

	running = make(chan int)

	// init args
	InitArgs()

	// init env
	InitEnv()

	// init cfg
	master.InitCfg(cfgPath)

	// init job-mgr
	master.InitJobMgr()

	// init api-server
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	<-running

ERR:
	fmt.Println(err)
}
