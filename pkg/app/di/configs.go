package di

import "time"

type AppConfigProvider interface {
	GetAppConfig() AppConfig
}

type AppConfig struct {
	ShutdownTimeout time.Duration `yaml:"shutdown-timeout"`
}

func (c AppConfig) GetAppConfig() AppConfig {
	return c
}

type HTTPConfigProvider interface {
	GetHTTPConfig() HTTPConfig
}

type HTTPConfig struct {
	Port int `yaml:"port"`
}

func (c HTTPConfig) GetHTTPConfig() HTTPConfig {
	return c
}
