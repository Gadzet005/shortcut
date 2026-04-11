package app

import "github.com/Gadzet005/shortcut/pkg/app/di"

type Config struct {
	di.AppConfig     `yaml:"app"`
	di.HTTPConfig    `yaml:"http"`
	di.LogConfig     `yaml:"logs"`
	di.MetricsConfig `yaml:"metrics"`
	TracingConfig    `yaml:"tracing"`
	MongoConfig      `yaml:"mongo"`
}

type TracingConfig struct {
	Enabled bool `yaml:"enabled"`
}

type MongoConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}
