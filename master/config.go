package master

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	ApiPort         int      `json:"api_port,omitempty"`
	ReadTimeOut     int64    `json:"read_time_out,omitempty"`
	WriteTimeOut    int64    `json:"write_time_out,omitempty"`
	EtcdDialTimeOut int64    `json:"etcd_dial_time_out,omitempty"`
	EtcdEndPoints   []string `json:"etcd_end_points,omitempty"`
	WebRoot         string   `json:"web_root,omitempty"`
}

var G_cfg *Configuration

func InitCfg(cfgPath string) error {
	var (
		content []byte
		err     error
		cfg     Configuration
	)
	if content, err = ioutil.ReadFile(cfgPath); err != nil {
		return err
	}
	if err = json.Unmarshal(content, &cfg); err != nil {
		return err
	}

	if G_cfg == nil {
		G_cfg = &cfg
	}

	return nil
}
