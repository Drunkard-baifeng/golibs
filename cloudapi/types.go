package cloudapi

import "time"

// ==================== 通用响应 ====================

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ==================== 类别相关 ====================

// ConfigPostReq 添加养号类别请求
type ConfigPostReq struct {
	Mode string `json:"mode"` // 类别名称（必填）
}

// ConfigPostResp 添加养号类别响应
type ConfigPostResp struct {
	ID uint `json:"id"`
}

// ==================== 数据相关 ====================

// DataPostReq 添加养号数据请求
type DataPostReq struct {
	ConfigID uint   `json:"config_id"` // 类别ID（必填）
	AccData  string `json:"acc_data"`  // 账号数据（必填）
	Notes1   string `json:"notes1"`    // 备注（可选）
	ExMode   string `json:"ex_mode"`   // 去重模式：不传=全行，1=第1列，2=第2列，1,3=第1和第3列组合
	Sep      string `json:"sep"`       // 列分隔符，默认 ----
}

// DataPostResp 添加养号数据响应
type DataPostResp struct {
	ID uint `json:"id"`
}

// DataGetReq 获取养号数据请求
type DataGetReq struct {
	ConfigID       uint   `json:"config_id"`       // 类别ID（必填）
	NextTimeMode   string `json:"next_time_mode"`  // 时间模式：1表示按下次执行时间获取
	IncludeStatus2 string `json:"include_status2"` // 是否包含状态2的数据
}

// DataGetResp 获取养号数据响应
type DataGetResp struct {
	ID          uint      `json:"id"`
	ConfigID    uint      `json:"config_id"`
	AccData     string    `json:"acc_data"`
	Notes1      string    `json:"notes1"`
	Status      int       `json:"status"`
	NextUseTime time.Time `json:"next_use_time"`
	TimeLogs    string    `json:"time_logs"`
}

// DataSetReq 修改养号数据请求（指针类型：不传=不修改）
type DataSetReq struct {
	ID          uint    `json:"id"`            // 数据ID（必填）
	Status      *int    `json:"status"`        // 状态：0:未养号,1:养号中 2:养号完成, 3:密码错误，4:账号封禁, 5:未实名, 6:停止养号
	Notes1      *string `json:"notes1"`        // 备注
	AccData     *string `json:"acc_data"`      // 账号数据
	NextUseTime *string `json:"next_use_time"` // 下次执行时间 格式：2006-01-02 15:04:05
}

// DataDelReq 删除养号数据请求
type DataDelReq struct {
	ID uint `json:"id"` // 数据ID（必填）
}

// ==================== 重置相关 ====================

// ResettingAllReq 重置全部养号任务请求
type ResettingAllReq struct {
	ConfigID uint `json:"config_id"` // 类别ID（必填）
}

// ResettingAllResp 重置全部养号任务响应
type ResettingAllResp struct {
	Count int64 `json:"count"`
}

// ResettingOneReq 重置单条养号任务请求
type ResettingOneReq struct {
	ID uint `json:"id"` // 数据ID（必填）
}

// ==================== 日志相关 ====================

// TimeLogGetReq 查询养号日志请求
type TimeLogGetReq struct {
	ID   uint    `json:"id"`   // 数据ID（必填）
	Date *string `json:"date"` // 日期 格式：yyyy-MM-dd（不传=返回所有日志）
}

// TimeLogGetResp 查询养号日志响应（单条）
type TimeLogGetResp struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

// TimeLogGetAllResp 查询养号日志响应（全部）
type TimeLogGetAllResp struct {
	Logs map[string]string `json:"logs"`
}

// TimeLogPostReq 添加/修改养号日志请求
type TimeLogPostReq struct {
	ID      uint    `json:"id"`      // 数据ID（必填）
	Date    *string `json:"date"`    // 日期 格式：yyyy-MM-dd（不传=今天）
	Content *string `json:"content"` // 日志内容
}

// TimeLogDelReq 删除养号日志请求
type TimeLogDelReq struct {
	ID   uint   `json:"id"`   // 数据ID（必填）
	Date string `json:"date"` // 日期 格式：yyyy-MM-dd（必填）
}
