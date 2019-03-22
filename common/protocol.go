package common

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
