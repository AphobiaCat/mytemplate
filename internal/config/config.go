package config

import (
	"time"
)

var C Config

type ServerConf struct {
	Host             string `json:"host"`
	Port             int
	CertFile         string `json:"cert_file"`
	KeyFile          string `json:"key_file"`
	Verbose          bool   `json:"verbose"`
	EnableAccessLog  bool   `json:"enable_access_log"` // enable/disable access log.
	MaxConns         int    `json:"max_conns"`
	MaxBytes         int64  `json:"max_bytes"`
	Timeout          int64
	CpuThreshold     int64
	TraceIgnorePaths []string
	NacosDiscovery   bool
	TrustedProxies   []string
}

type MySQLConfig struct {
	DSN         string
	ReplicasDSN string
	Mock        bool
}

type RedisConf struct {
	Host             string
	ReadOnly         bool
	RouteByLatency   bool
	RouteRandomly    bool
	SingleReplicaSet bool
	Type             string
	Password         string
	EnableTls        bool
	EnableBreaker    bool
	NonBlock         bool
	PingTimeout      time.Duration
	DB               int
	PoolSize         int
	MaxActiveConns   int
	MaxIdleConns     int
	MinIdleConns     int
}

type Config struct {
	ServerConf   `mapstructure:",squash"`
	MysqlExample MySQLConfig
	Redis        RedisConf
	Env          string
	ServerJwtKey string
}
