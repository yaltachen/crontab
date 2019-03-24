package worker

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

func InitApiServer() (err error) {
	var (
		httpServer *http.Server
		listener   net.Listener
		router     *gin.Engine
	)

	router = gin.Default()

	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiServerPort)); err != nil {
		return err
	}

	httpServer = &http.Server{
		Handler:      router,
		ReadTimeout:  time.Duration(G_config.ReadTimeOut) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.WriteTimeOut) * time.Microsecond,
	}

	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	go httpServer.Serve(listener)

	return
}
