package app

import configutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/config"

type Config struct {
	Env        configutils.Env  `yaml:"env"`
	HTTPServer HTTPServerConfig `yaml:"http-server"`
	Namespace  map[string]NamespaceConfig  `yaml:"namespace"`
}

type HTTPServerConfig struct {
	Port uint `yaml:"port"`
}

type NamespaceConfig struct {
	Services []string `yaml:"services"`
	Graphs   []string `yaml:"graphs"`
}

type ServiceConfig struct {
	Host string `yaml:"host"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}

type EndpointConfig struct {
	Path string `yaml:"path"`
	TimeoutMs int `yaml:"timeout_ms"`
	RetriesNum int `yaml:"retries_num"`
	ResultIDs []string `yaml:"result_ids"`
}

type NodeConfig struct {
	EndpointID string `yaml:"endpoint_id"`
	Dependencies map[string][]string `yaml:"dependencies"`
}

type GraphConfig struct {
	Nodes map[string]NodeConfig `yaml:"nodes"`
	FailureStrategy string `yaml:"failure_strategy"`
}
