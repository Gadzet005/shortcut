package graphconfig

import "time"

type NodeType string

const (
	NodeTypeDefault     NodeType = "default"
	NodeTypeTransparent NodeType = "transparent"
	NodeTypeHTTPAdapter NodeType = "http-adapter"
)

type Config struct {
	Namespaces map[string]NamespaceConfig
}

type NamespaceConfig struct {
	HTTPRoutes map[string]HTTPRouteConfig
	Services   ServicesConfig
	Graphs     map[string]GraphConfig
}

type GraphConfig struct {
	Nodes           map[string]NodeConfig
	InputNode       string
	OutputNode      string
	FailureStrategy string
	TimeoutMs       int
}

type NodeConfig struct {
	ID           string
	Type         NodeType
	EndpointID   string
	Dependencies []DependencyConfig
	Cache        *CacheNodeConfig
}

type CacheNodeConfig struct {
	Enabled bool
	TTL     time.Duration
}

type ServicesConfig struct {
	Endpoints map[string]EndpointDef
}

type EndpointDef struct {
	URL               string
	TimeoutMs         int
	RetriesNum        int
	InitialIntervalMs int
	BackoffMultiplier float64
	MaxIntervalMs     int
}

type HTTPRouterConfig struct {
	Routes map[string]HTTPRouteConfig `yaml:"routes"`
}

type HTTPRouteConfig struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
	Graph  string `yaml:"graph"`
}

type ServiceConfig struct {
	Host      string                    `yaml:"host"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}

type EndpointConfig struct {
	Path              string  `yaml:"path"`
	TimeoutMs         int     `yaml:"timeout-ms"`
	RetriesNum        int     `yaml:"retries-num"`
	InitialIntervalMs int     `yaml:"initial-interval-ms"`
	BackoffMultiplier float64 `yaml:"backoff-multiplier"`
	MaxIntervalMs     int     `yaml:"max-interval-ms"`
}

type GraphFileConfig struct {
	Nodes           map[string]NodeFileConfig `yaml:"nodes"`
	InputNode       string                    `yaml:"input-node"`
	OutputNode      string                    `yaml:"output-node"`
	FailureStrategy string                    `yaml:"failure-strategy"`
	TimeoutMs       int                       `yaml:"timeout-ms"`
}

type NodeFileConfig struct {
	Type         NodeType              `yaml:"type"`
	EndpointID   string                `yaml:"endpoint-id"`
	Dependencies []DependencyConfig    `yaml:"dependencies"`
	Cache        *CacheNodeFileConfig  `yaml:"cache"`
}

type CacheNodeFileConfig struct {
	Enabled bool `yaml:"enabled"`
	TTLMS   int  `yaml:"ttl-ms"`
}

type DependencyConfig struct {
	NodeID          string `yaml:"node-id"`
	ItemID          string `yaml:"item-id"`
	OverridenItemID string `yaml:"overriden-item-id"`
}
