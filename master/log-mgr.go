package master

import (
	"context"
	"log"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
	"github.com/yaltachen/crontab/common"
)

// mongodb存储日志
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var G_logMgr *LogMgr

func InitLogSink() (err error) {
	var (
		client *mongo.Client
	)

	// 建立mongodb连接
	if client, err = mongo.Connect(
		context.TODO(),
		G_cfg.MongodbUri,
		clientopt.ConnectTimeout(time.Duration(G_cfg.MongodbConnectTimeout)*time.Millisecond)); err != nil {
		return
	}

	//   选择db和collection
	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}

	return
}

func (lm *LogMgr) ListLog(jobName string, skip, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  mongo.Cursor
		jobLog  *common.JobLog
	)

	// 初始化logArr
	logArr = make([]*common.JobLog, 0)

	// 过滤条件
	filter = &common.JobLogFilter{JobName: jobName}
	// 按照任务开始事件倒序
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	if cursor, err = lm.logCollection.Find(context.TODO(), filter, findopt.Sort(logSort),
		findopt.Skip(int64(skip)), findopt.Limit(int64(limit))); err != nil {
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		// 反序列化
		if err = cursor.Decode(jobLog); err != nil {
			log.Printf("cursor decode job log failed. Error: %v\r\n", err)
			continue
		}

		logArr = append(logArr, jobLog)
	}

	return logArr, nil
}
