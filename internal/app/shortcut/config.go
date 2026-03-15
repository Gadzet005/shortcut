package app

import "github.com/Gadzet005/shortcut/pkg/app/di"

type Config struct {
	di.AppConfig
	di.HTTPConfig
	MetricsPort uint   `yaml:"metrics-port"`
	LogPath     string `yaml:"logs"`
}
