package master

import "github.com/yaltachen/crontab/common"

var (
	ErrJsonUnmarshal   = common.Resp{ErrCode: 1000, ErrMsg: "json unmarshal failed."}
	ErrTemplateParse   = common.Resp{ErrCode: 1001, ErrMsg: "template parse failed."}
	ErrTemplateExecute = common.Resp{ErrCode: 1002, ErrMsg: "template execute failed."}

	ErrSaveJob     = common.Resp{ErrCode: 2000, ErrMsg: "save job failed."}
	ErrDeleteJob   = common.Resp{ErrCode: 2001, ErrMsg: "delete job failed."}
	ErrListJob     = common.Resp{ErrCode: 2002, ErrMsg: "list job failed."}
	ErrKillJob     = common.Resp{ErrCode: 2003, ErrMsg: "kill job failed."}
	ErrListJobLog  = common.Resp{ErrCode: 2004, ErrMsg: "list job logs failed."}
	ErrListWorkers = common.Resp{ErrCode: 2005, ErrMsg: "list online workers failed."}

	ErrEmptyJobName  = common.Resp{ErrCode: 3000, ErrMsg: "emtpy job name"}
	ErrBadSkipValue  = common.Resp{ErrCode: 3001, ErrMsg: "skip shoubld be number"}
	ErrBadLimitValue = common.Resp{ErrCode: 3002, ErrMsg: "limit should be number"}
)
