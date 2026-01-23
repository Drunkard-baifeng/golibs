package proxypool

import "sync"

var (
	defaultPool *ProxyPool
	once        sync.Once
)

// Default 获取默认代理池（单例）
func Default() *ProxyPool {
	once.Do(func() {
		defaultPool = New(Config{})
	})
	return defaultPool
}

// InitDefault 初始化默认代理池
func InitDefault(cfg Config) *ProxyPool {
	once.Do(func() {
		defaultPool = New(cfg)
	})
	return defaultPool
}

// ResetDefault 重置默认代理池（用于测试）
func ResetDefault() {
	once = sync.Once{}
	defaultPool = nil
}

