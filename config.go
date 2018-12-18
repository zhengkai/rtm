package rtm

import (
	"time"
)

// Config 配置文件结构
type Config struct {
	ProjectID          int32
	SignatureSecretKey string
	ServerGate         string
	ClientGate         string
	Timeout            time.Duration
	IgnorePing         bool
}

// ConfigClient client config 结构
type ConfigClient struct {
	ProjectID  int32
	UID        int64
	ClientGate string
	IgnorePing bool
}

var globalConfig *Config

// SetConfig 设置缺省配置
func SetConfig(c *Config) {
	globalConfig = c
}
