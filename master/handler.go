package master

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/yaltachen/crontab/common"
)

func handleHome(ctx *gin.Context) {
	var (
		t   *template.Template
		err error
	)
	if t, err = template.ParseFiles(G_cfg.WebRoot + "home.html"); err != nil {
		log.Printf("home.html parse failed. Error: %v\r\n", err)
		ctx.JSON(500, ErrTemplateParse)
		return
	}

	if err = t.Execute(ctx.Writer, nil); err != nil {
		log.Printf("home.html template execute failed. Error: %v\r\n", err)
		ctx.JSON(500, ErrTemplateExecute)
		return
	}
}

// save job
// post /job/:job-name job={"name": "job1", "command": "echo hello", "cronExpr": "*****"}
func handleJobSave(ctx *gin.Context) {
	var (
		jobStr string
		job    common.Job
		oldJob *common.Job
		err    error
	)

	jobStr = ctx.PostForm("job")

	if err = json.Unmarshal([]byte(jobStr), &job); err != nil {
		log.Printf("unmarshal job: %s failed. Error: %v\r\n", jobStr, err)
		ctx.JSON(500, ErrJsonUnmarshal)
		return
	}

	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		log.Printf("save job: %s failed. Error: %v\r\n", jobStr, err)
		ctx.JSON(500, ErrSaveJob)
		return
	}

	ctx.JSON(200, common.Resp{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data:    oldJob,
	})

}

// del jon
// delete /job/:job-name
func handleJobDel(ctx *gin.Context) {
	var (
		jobName string
		err     error
		oldJob  *common.Job
	)
	if jobName = ctx.Param("job-name"); len(jobName) == 0 {
		ctx.JSON(http.StatusBadRequest, ErrEmptyJobName)
		return
	}

	if oldJob, err = G_jobMgr.DeleteJob(jobName); err != nil {
		log.Printf("delete job: %s failed. Error: %v\r\n", jobName, err)
		ctx.JSON(http.StatusInternalServerError, ErrDeleteJob)
		return
	}

	ctx.JSON(200, common.Resp{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data:    oldJob,
	})
	return
}

// list jobs
// get /job
func handleJobList(ctx *gin.Context) {
	var (
		err  error
		jobs []*common.Job
	)

	if jobs, err = G_jobMgr.ListJobs(); err != nil {
		log.Printf("List job failed. Error: %v\r\n", err)
		ctx.JSON(500, ErrListJob)
	}

	ctx.JSON(200, common.Resp{
		Data:    jobs,
		ErrCode: 0,
		ErrMsg:  "ok",
	})

	return
}

// kill job
// post /job/kill job-name=job
func handleJobKill(ctx *gin.Context) {
	var (
		err     error
		jobName string
	)

	if jobName = ctx.PostForm("job-name"); len(jobName) == 0 {
		ctx.JSON(500, ErrEmptyJobName)
		return
	}

	if err = G_jobMgr.KillJob(jobName); err != nil {
		log.Printf("Kill <job: %s> failed. Error: %v\r\n", jobName, err)
		ctx.JSON(500, ErrKillJob)
		return
	}
	ctx.JSON(200, common.Resp{ErrCode: 0, ErrMsg: "ok"})
}
