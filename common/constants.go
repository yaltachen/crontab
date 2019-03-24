package common

import "errors"

const (
	// 任务保存目录
	JOB_SAVE_DIR = "/cron/jobs/"
	// 任务杀死目录
	JOB_KILL_DIR = "/cron/kill/"
	// 锁
	JOB_LOCK_DIR = "/cron/lock/"
	// 注册
	WORKER_DIR = "/cron/workers/"

	// 保存任务事件
	JOB_EVENT_SAVE = 1
	// 删除任务事件
	JOB_EVENT_DELETE = 2
	// 强杀任务
	JOB_EVENT_KILL = 3
)

var (
	ERR_KEY_ALREADY_REQUIRED = errors.New("锁已被占用")
	ERR_KILL_WATCHER_CLOSE   = errors.New("killer watcher channel close.")
	ERR_NO_LOCAL_IP_FOUND    = errors.New("没有找到网卡IP")
)
