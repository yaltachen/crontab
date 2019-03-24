package worker

import (
	"log"
	"sync"
	"time"

	"github.com/yaltachen/crontab/common"
)

// 任务调度
type Scheduler struct {
	JobEventChan      chan *common.JobEvent         // etcd任务事件队列
	JobResultChan     chan *common.JobExecuteResult // 执行结果队列
	JobPlanTable      *sync.Map
	JobExecutingTable *sync.Map
}

var (
	G_scheduler *Scheduler
)

func InitScheduler() {
	G_scheduler = &Scheduler{
		JobEventChan:      make(chan *common.JobEvent, 1000),
		JobResultChan:     make(chan *common.JobExecuteResult, 1000),
		JobPlanTable:      &sync.Map{},
		JobExecutingTable: &sync.Map{},
	}
	// 启动调度协程
	go G_scheduler.schedulerLoop()
}

// 推送任务变化事件
func (s *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	s.JobEventChan <- jobEvent
}

func (s *Scheduler) tryScheduler() (scheduleAfter time.Duration) {

	var (
		now      time.Time
		nearTime *time.Time
		count    int
	)
	// 获取当前时间
	now = time.Now()

	// 1. 遍历所有任务
	s.JobPlanTable.Range(func(name, plan interface{}) bool {
		count++
		// 2. 任务到期立即执行
		if plan.(*common.JobSchedulePlan).NextTime.Before(now) ||
			plan.(*common.JobSchedulePlan).NextTime.Equal(now) {
			// 尝试执行任务
			s.tryStartJob(plan.(*common.JobSchedulePlan))
			// 更新下一次执行时间
			plan.(*common.JobSchedulePlan).NextTime = plan.(*common.JobSchedulePlan).Expr.Next(now)
		}
		// 统计最近要过期的时间
		if nearTime == nil || plan.(*common.JobSchedulePlan).NextTime.Before(*nearTime) {
			nearTime = &plan.(*common.JobSchedulePlan).NextTime
		}
		return true
	})

	if count == 0 {
		// 当前任务为空，睡眠随机时间
		scheduleAfter = 1 * time.Second
	} else {
		// 下次调度间隔（最近要执行任务调度时间-当前时间）
		scheduleAfter = (*nearTime).Sub(now)
	}
	return
}

func (s *Scheduler) tryStartJob(jobPlan *common.JobSchedulePlan) {
	// 调度和执行是两件事
	// 执行的任务可能很久，1分调度60次，但只能执行1次，去重放置并发！
	// 如果job正在执行，跳过
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)
	if _, jobExecuting = s.JobExecutingTable.Load(jobPlan.Job.Name); jobExecuting {
		// job正在执行，跳过本次执行
		log.Printf("job: %s is running, skip this execution.\r\n", jobPlan.Job.Name)
		return
	}
	// 执行job
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)
	// 保存执行状态
	s.JobExecutingTable.Store(jobPlan.Job.Name, jobExecuteInfo)

	// 执行任务
	// log.Printf("start executing job: %s. PlanExeTime: %v, RealExeTime: %v.\r\n",
	// 	jobPlan.Job.Name, jobExecuteInfo.PlanExeTime, jobExecuteInfo.RealExeTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

// 调度协程
func (s *Scheduler) schedulerLoop() {
	// 定时任务
	var (
		jobEvent      *common.JobEvent
		jobResult     *common.JobExecuteResult
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
	)

	// 初始化一次(1秒)
	scheduleAfter = s.tryScheduler()

	// 调度的延时定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	for {
		select {
		case jobEvent = <-s.JobEventChan: // 监听任务变化
			// 对内存中维护的任务列表做增删改查
			s.handleJobEvent(jobEvent)
		case jobResult = <-s.JobResultChan:
			s.handleJobResult(jobResult)
		case <-scheduleTimer.C: // 最近的任务到期了
		}
		// 重新调度任务
		scheduleAfter = s.tryScheduler()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 处理JobEvent
func (s *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExcuteInfo   interface{}
		jobExcuting     bool
		err             error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		// 保存任务事件
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			log.Printf("build job: %s schedule plan failed. Error: %s\r\n", jobEvent.Job.Name, err)
		} else {
			s.JobPlanTable.Store(jobEvent.Job.Name, jobSchedulePlan)
		}
	case common.JOB_EVENT_DELETE:
		// 删除任务事件
		s.JobPlanTable.Delete(jobEvent.Job.Name)
	case common.JOB_EVENT_KILL:
		// 强杀任务事件
		// 取消掉command执行
		if jobExcuteInfo, jobExcuting = s.JobExecutingTable.Load(jobEvent.Job.Name); jobExcuting {
			// 强杀任务
			log.Printf("kill job: %s\r\n", jobExcuteInfo.(*common.JobExecuteInfo).Job.Name)
			jobExcuteInfo.(*common.JobExecuteInfo).CancelFunc()
		} else {
		}
	}
}

// 处理JobResult
func (s *Scheduler) handleJobResult(jobResult *common.JobExecuteResult) {
	var (
		jobLog *common.JobLog
	)
	// 从JobExecutingTable中删除
	s.JobExecutingTable.Delete(jobResult.ExecuteInfo.Job.Name)

	// if jobResult.Err != nil {
	// 	// 任务执行出错
	// 	log.Printf("Execute Job: %s failed. Error: %v\r\n", jobResult.ExecuteInfo.Job.Name, jobResult.Err)
	// } else {
	// 	log.Printf("Job: %s finished. startTime: %v, endTime: %v, output: %s",
	// 		jobResult.ExecuteInfo.Job.Name, jobResult.StartTime, jobResult.EndTime, string(jobResult.Output))
	// }

	if jobResult.Err == nil {
		jobLog = &common.JobLog{
			JobName:      jobResult.ExecuteInfo.Job.Name,
			Command:      jobResult.ExecuteInfo.Job.Command,
			Err:          "",
			Output:       string(jobResult.Output),
			PlanTime:     jobResult.ExecuteInfo.PlanExeTime.UnixNano() / 1000 / 1000,
			ScheduleTime: jobResult.ExecuteInfo.RealExeTime.UnixNano() / 1000 / 1000,
			StartTime:    jobResult.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      jobResult.EndTime.UnixNano() / 1000 / 1000,
		}
	}
	if jobResult.Err != nil {
		jobLog.Err = jobResult.Err.Error()
	}
	// TODO: 存储到mongodb
	G_logSink.logChan <- jobLog
}
