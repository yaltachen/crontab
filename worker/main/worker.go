package main

import (
	"flag"
	"runtime"

	"github.com/yaltachen/crontab/worker"
)

var (
	cfgPath string
)

func InitEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitArgs() {
	flag.StringVar(&cfgPath, "config", "./worker.json", "woker 配置")
	flag.Parse()
}

func main() {
	var (
		running chan int
		err     error
	)

	running = make(chan int)

	InitEnv()
	InitArgs()

	if err = worker.InitCfg(cfgPath); err != nil {
		panic(err)
	}

	// if err = worker.InitApiServer(); err != nil {
	// 	panic(err)
	// }

	// 初始化调度器
	worker.InitScheduler()

	// 初始化执行器
	worker.InitExecutor()

	// 初始化JobMgr
	if err = worker.InitJobMgr(); err != nil {
		panic(err)
	}

	// 初始化LogSink
	if err = worker.InitLogSink(); err != nil {
		panic(err)
	}

	// 初始化Register
	if err = worker.InitRegister(); err != nil {
		panic(err)
	}

	<-running
}
