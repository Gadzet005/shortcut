package config

type Config struct {
	Namespace  map[string]NamespaceConfig  `yaml:"namespace"`
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
