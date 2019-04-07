package worker

import (
	"math/rand"
	"os/exec"
	"time"

	"github.com/yaltachen/crontab/common"
)

// 任务执行器
type Executor struct {
}

var G_executor *Executor

func InitExecutor() {
	if G_executor == nil {
		G_executor = &Executor{}
	}
}

func (e *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			cmd       *exec.Cmd
			err       error
			output    []byte
			result    *common.JobExecuteResult
			startTime time.Time
			endTime   time.Time
			jobLocker *JobLocker
		)

		// 上锁
		jobLocker = G_jobMgr.NewLocker(info.Job.Name)

		// 随机睡眠(0~1s)，尽量保证任务均匀分配
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		startTime = time.Now()
		if err = jobLocker.TryLock(); err != nil {
			// 上锁失败
			endTime = time.Now()
			result = &common.JobExecuteResult{
				ExecuteInfo: info,
				Err:         err,
				StartTime:   startTime,
				EndTime:     endTime,
			}
		} else {
			// 上锁成功
			startTime = time.Now()
			// 执行shell命令
			cmd = exec.CommandContext(info.CancelCtx, G_config.Bin, "-c", info.Job.Command)
			// 执行并捕获输出
			output, err = cmd.CombinedOutput()
			endTime = time.Now()

			// 任务执行完毕之后，把执行结果返回给scheduler，scheduler从executingTable中删除任务
			result = &common.JobExecuteResult{
				ExecuteInfo: info,
				Output:      output,
				Err:         err,
				StartTime:   startTime,
				EndTime:     endTime,
			}
			// 解锁
			jobLocker.Unlock()
		}

		G_scheduler.JobResultChan <- result
	}()
}
