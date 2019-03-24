package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	WriteTimeOut          int64    `json:"write_time_out,omitempty"`
	ReadTimeOut           int64    `json:"read_time_out,omitempty"`
	ApiServerPort         int      `json:"api_server_port,omitempty"`
	EtcdDialTimeout       int      `json:"etcd_dial_timeout,omitempty"`
	EtcdEndPoints         []string `json:"etcd_end_points,omitempty"`
	Bin                   string   `json:"bin,omitempty"`
	JobLogBatchSize       int      `json:"job_log_batch_size,omitempty"`
	MongodbUri            string   `json:"mongodb_uri,omitempty"`
	MongodbConnectTimeout int      `json:"mongodb_connect_timeout,omitempty"`
	JobLogCommitTimeout   int      `json:"job_log_commit_timeout,omitempty"`
}

var G_config *Configuration

func InitCfg(cfgPath string) (err error) {

	var (
		content []byte
		cfg     Configuration
	)

	if content, err = ioutil.ReadFile(cfgPath); err != nil {
		return err
	}

	if err = json.Unmarshal(content, &cfg); err != nil {
		return err
	}

	if G_config == nil {
		G_config = &cfg
	}

	return nil
}
