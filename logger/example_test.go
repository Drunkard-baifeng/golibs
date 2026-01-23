package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {

	Debug("这是调试信息")
	Info("这是普通信息")
	Warn("这是警告信息")
	Error("这是错误信息")
	Success("这是成功信息")

	// 格式化
	Infof("代理获取成功: %s:%d", "1.2.3.4", 8080)

	// 带字段（key-value 对）
	Info("获取代理",
		"ip", "1.2.3.4",
		"port", 8080,
		"remaining", 3,
	)

	Info("任务完成",
		"taskID", 1,
		"duration", "2.5s",
		"status", "success",
	)

	// 带错误
	Error("请求失败",
		"url", "http://example.com",
		"error", "连接超时",
	)
}
