package app

import configutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/config"

type Config struct {
	Env        configutils.Env  `yaml:"env"`
	HTTPServer HTTPServerConfig `yaml:"http-server"`
}

type HTTPServerConfig struct {
	Port uint `yaml:"port"`
}
