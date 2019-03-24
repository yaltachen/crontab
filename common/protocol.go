package common

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

type Job struct {
	Name     string `json:"name,omitempty"`
	Command  string `json:"command,omitempty"`
	CronExpr string `json:"cron_expr,omitempty"`
}

type Resp struct {
	Data    interface{} `json:"data,omitempty"`
	ErrCode int         `json:"err_code"`
	ErrMsg  string      `json:"err_msg"`
}

type JobEvent struct {
	EventType int
	Job       *Job
}

type JobSchedulePlan struct {
	Job      *Job
	Expr     *cronexpr.Expression // job解析好的 cron expr表达式
	NextTime time.Time            // 下次调度时间
}

type JobExecuteInfo struct {
	Job         *Job            `json:"job,omitempty"`
	PlanExeTime time.Time       `json:"plan_exe_time,omitempty"`
	RealExeTime time.Time       `json:"real_exe_time,omitempty"`
	CancelCtx   context.Context `json:"cancel_ctx,omitempty"`
	CancelFunc  func()          `json:"cancel_func,omitempty"`
}

type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo `json:"execute_info,omitempty"`
	Err         error           `json:"err,omitempty"`
	Output      []byte          `json:"output,omitempty"`
	StartTime   time.Time       `json:"start_time,omitempty"`
	EndTime     time.Time       `json:"end_time,omitempty"`
}

// 任务执行日志
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`           // 任务名字
	Command      string `json:"command" bson:"command"`           // 脚本命令
	Err          string `json:"err" bson:"err"`                   // 错误原因
	Output       string `json:"output" bson:"output"`             // 脚本输出
	PlanTime     int64  `json:"planTime" bson:"planTime"`         // 计划开始时间
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"` // 实际调度时间
	StartTime    int64  `json:"startTime" bson:"startTime"`       // 任务执行开始时间
	EndTime      int64  `json:"endTime" bson:"endTime"`           // 任务执行结束时间
}

// 日志批次
type LogBatch struct {
	Logs []interface{} // 多条日志
}

// 任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

// 任务日志排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"` // {startTime: -1}
}

type OnlineWorker struct {
	WorkerIP string `json:"worker_ip,omitempty"`
}

func ExtractWorkerIp(key string) string {
	return strings.TrimPrefix(key, WORKER_DIR)
}

func ExtractJobName(jobkey string) string {
	return strings.TrimPrefix(jobkey, JOB_SAVE_DIR)
}

func ExtractKillerName(jobkey string) string {
	return strings.TrimPrefix(jobkey, JOB_KILL_DIR)
}

func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

func UnpackJob(data []byte) (job *Job, err error) {
	job = &Job{}
	if err = json.Unmarshal(data, job); err != nil {
		return nil, err
	}
	return job, nil

}

// 构造执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expression *cronexpr.Expression
	)
	if expression, err = cronexpr.Parse(job.CronExpr); err != nil {
		return nil, err
	}

	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expression,
		NextTime: expression.Next(time.Now()),
	}

	return
}

// 构造执行状态细信息
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:         jobSchedulePlan.Job,
		PlanExeTime: jobSchedulePlan.NextTime,
		RealExeTime: time.Now(),
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return jobExecuteInfo
}
