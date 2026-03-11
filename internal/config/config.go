package config

import (
	"time"
)

var C Config

type ServerConf struct {
	Host            string `json:",default=0.0.0.0"`
	Port            int
	CertFile        string `json:",optional"`
	KeyFile         string `json:",optional"`
	Verbose         bool   `json:",optional"`
	EnableAccessLog bool   `json:",optional,default=true"` // enable/disable access log.
	MaxConns        int    `json:",default=10000"`
	MaxBytes        int64  `json:",default=1048576"`
	// milliseconds
	// nolint:all
	Timeout int64 `json:",default=3000"`
	// nolint:all
	CpuThreshold int64 `json:",default=0,range=[0:1000)"`
	// TraceIgnorePaths is paths blacklist for trace middleware.
	TraceIgnorePaths []string `json:",optional"`
	NacosDiscovery   bool     `json:",default=false"`
}

type MySQLConfig struct {
	DSN         string
	ReplicasDSN string
	Mock        bool
}

type RedisConf struct {
	Host             string
	ReadOnly         bool   `json:",optional"`
	RouteByLatency   bool   `json:",optional"`
	RouteRandomly    bool   `json:",optional"`
	SingleReplicaSet bool   `json:",optional"`
	Type             string `json:",default=node,options=node|cluster"`
	Password         string `json:",optional"`
	EnableTls        bool   `json:",optional"`
	EnableBreaker    bool   `json:",default=true"`
	NonBlock         bool   `json:",default=true"`
	// PingTimeout is the timeout for ping redis.
	PingTimeout    time.Duration `json:",default=1s"`
	DB             int           `json:",default=0"`
	PoolSize       int           `json:",optional"`
	MaxActiveConns int           `json:",optional"`
	MaxIdleConns   int           `json:",optional"`
	MinIdleConns   int           `json:",optional"`
}

type Config struct {
	ServerConf   `mapstructure:",squash"`
	MysqlExample MySQLConfig
	Redis        RedisConf
	Env          string
}
