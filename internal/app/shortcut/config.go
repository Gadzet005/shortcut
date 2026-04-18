package app

import "github.com/Gadzet005/shortcut/pkg/app/di"

type Config struct {
	di.AppConfig     `yaml:"app"`
	di.HTTPConfig    `yaml:"http"`
	di.LogConfig     `yaml:"logs"`
	di.MetricsConfig `yaml:"metrics"`
	TracingConfig    `yaml:"tracing"`
	MongoConfig      `yaml:"mongo"`
	PostgresConfig   `yaml:"postgres"`
}

type TracingConfig struct {
	Enabled bool `yaml:"enabled"`
}

type MongoConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

type PostgresConfig struct {
	URI      string `yaml:"uri"`
}
