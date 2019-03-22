package master

import "github.com/yaltachen/crontab/common"

var (
	ErrJsonUnmarshal   = common.Resp{ErrCode: 1000, ErrMsg: "json unmarshal failed."}
	ErrTemplateParse   = common.Resp{ErrCode: 1001, ErrMsg: "template parse failed."}
	ErrTemplateExecute = common.Resp{ErrCode: 1002, ErrMsg: "template execute failed."}

	ErrSaveJob   = common.Resp{ErrCode: 2000, ErrMsg: "job save failed."}
	ErrDeleteJob = common.Resp{ErrCode: 2001, ErrMsg: "job delete failed."}
	ErrListJob   = common.Resp{ErrCode: 2002, ErrMsg: "job list failed."}
	ErrKillJob   = common.Resp{ErrCode: 2003, ErrMsg: "job kill failed."}

	ErrEmptyJobName = common.Resp{ErrCode: 3000, ErrMsg: "emtpy job name"}
)
