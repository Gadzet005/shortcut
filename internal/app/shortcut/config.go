package app

import "github.com/Gadzet005/shortcut/pkg/app/di"

type Config struct {
	di.AppConfig     `yaml:"app"`
	di.HTTPConfig    `yaml:"http"`
	di.LogConfig     `yaml:"logs"`
	di.MetricsConfig `yaml:"metrics"`
	MongoConfig      `yaml:"mongo"`
	CacheConfig      `yaml:"cache"`
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
