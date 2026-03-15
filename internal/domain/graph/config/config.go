package graphconfig

type NodeType string

const (
	NodeTypeDefault     NodeType = "default"
	NodeTypeTransparent NodeType = "transparent"
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
}

type NodeConfig struct {
	ID           string
	Type         NodeType
	EndpointID   string
	Dependencies []DependencyConfig
}

type ServicesConfig struct {
	Endpoints map[string]EndpointDef
}

type EndpointDef struct {
	URL        string
	TimeoutMs  int
	RetriesNum int
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
	Path       string `yaml:"path"`
	TimeoutMs  int    `yaml:"timeout-ms"`
	RetriesNum int    `yaml:"retries-num"`
}

type GraphFileConfig struct {
	Nodes           map[string]NodeFileConfig `yaml:"nodes"`
	InputNode       string                    `yaml:"input-node"`
	OutputNode      string                    `yaml:"output-node"`
	FailureStrategy string                    `yaml:"failure-strategy"`
}

type NodeFileConfig struct {
	Type         NodeType           `yaml:"type"`
	EndpointID   string             `yaml:"endpoint-id"`
	Dependencies []DependencyConfig `yaml:"dependencies"`
}

type DependencyConfig struct {
	NodeID          string `yaml:"node-id"`
	ItemID          string `yaml:"item-id"`
	OverridenItemID string `yaml:"overriden-item-id"`
}
