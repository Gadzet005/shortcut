package app

import (
	"time"

	"github.com/Gadzet005/shortcut/pkg/app/di"
)

type Config struct {
	di.AppConfig     `yaml:"app"`
	di.HTTPConfig    `yaml:"http"`
	di.LogConfig     `yaml:"logs"`
	di.MetricsConfig `yaml:"metrics"`
	MongoConfig      `yaml:"mongo"`
	CacheConfig      `yaml:"cache"`
	PostgresConfig   `yaml:"postgres"`
	FailureWorker    FailureWorkerConfig `yaml:"failure-worker"`
	Trace            TraceConfig         `yaml:"trace"`
}

type MongoConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

type CacheConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type PostgresConfig struct {
	URL string `yaml:"url"`
}

type FailureWorkerConfig struct {
	Interval          time.Duration `yaml:"interval"`
	BatchSize         int           `yaml:"batch-size"`
	VisibilityTimeout time.Duration `yaml:"visibility-timeout"`
}

type TraceConfig struct {
	TTL time.Duration `yaml:"ttl"`
}
