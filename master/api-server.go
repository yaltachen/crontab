package master

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ApiServer struct {
	httpServer *http.Server
}

var G_apiServer *ApiServer

func InitApiServer() error {

	var (
		listener   net.Listener
		httpServer *http.Server
		router     *gin.Engine
		err        error
	)

	router = gin.Default()

	router.StaticFS("/statics", http.Dir(G_cfg.WebRoot))

	router.GET("/home", handleHome)

	// 任务增删改查
	router.POST("/job/:job-name", handleJobSave)
	router.DELETE("/job/:job-name", handleJobDel)
	router.GET("/job", handleJobList)

	// 查看任务日志
	router.GET("/log/:job-name/:skip/:limit", handleLogList)

	// 强杀任务
	router.POST("/kill/:job-name", handleJobKill)

	// 查看健康节点
	router.GET("/worker", handleGetOnlineWorkers)

	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_cfg.ApiPort)); err != nil {
		return err
	}

	httpServer = &http.Server{
		Handler:      router,
		WriteTimeout: time.Duration(G_cfg.WriteTimeOut) * time.Millisecond,
		ReadTimeout:  time.Duration(G_cfg.ReadTimeOut) * time.Millisecond,
	}

	if G_apiServer == nil {
		G_apiServer = &ApiServer{
			httpServer: httpServer,
		}
	}

	// 启动服务端
	go httpServer.Serve(listener)

	return nil
}
