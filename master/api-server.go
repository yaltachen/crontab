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

	router.POST("/job/:job-name", handleJobSave)
	router.DELETE("/job/:job-name", handleJobDel)
	router.GET("/job", handleJobList)

	router.POST("/kill/:job-name", handleJobKill)

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
