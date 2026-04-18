package graphconfig

import (
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gadzet005/shortcut/pkg/app/config"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	httpRouterName = "http_router.yaml"
	graphsDir      = "graphs"
	servicesDir    = "services"

	defaultInputNode  = "input"
	defaultOutputNode = "output"
)

func Load(dir string) (Config, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Config{}, errors.Wrapf(err, "read config dir %s", dir)
	}

	namespaces := make(map[string]NamespaceConfig)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		namespaceID := e.Name()
		namespaceDir := filepath.Join(dir, e.Name())

		ns, err := loadNamespace(namespaceDir)
		if err != nil {
			return Config{}, errors.Wrapf(err, "namespace %s", namespaceID)
		}

		namespaces[namespaceID] = ns
	}

	return Config{Namespaces: namespaces}, nil
}

func loadNamespace(namespaceDir string) (NamespaceConfig, error) {
	ns := NamespaceConfig{
		HTTPRoutes: nil,
		Services:   ServicesConfig{Endpoints: make(map[string]EndpointDef)},
		Graphs:     make(map[string]GraphConfig),
	}

	routerPath := filepath.Join(namespaceDir, httpRouterName)
	if data, err := os.ReadFile(routerPath); err == nil {
		var router HTTPRouterConfig
		if err := yaml.Unmarshal(data, &router); err != nil {
			return NamespaceConfig{}, errors.Wrap(err, "load http_router.yaml")
		}
		ns.HTTPRoutes = router.Routes
	}

	servicesDirPath := filepath.Join(namespaceDir, servicesDir)
	services, err := loadServicesFromDir(servicesDirPath)
	if err != nil {
		return NamespaceConfig{}, errors.Wrap(err, "read service configs")
	}
	maps.Copy(ns.Services.Endpoints, services.Endpoints)

	graphsDirPath := filepath.Join(namespaceDir, graphsDir)
	graphs, err := loadGraphsFromDir(graphsDirPath)
	if err != nil {
		return NamespaceConfig{}, errors.Wrap(err, "read graph configs")
	}
	maps.Copy(ns.Graphs, graphs)

	return ns, nil
}

func loadServicesFromDir(servicesDirPath string) (ServicesConfig, error) {
	out := ServicesConfig{Endpoints: make(map[string]EndpointDef)}

	entries, err := os.ReadDir(servicesDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return ServicesConfig{}, errors.Wrap(err, "read services dir")
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		serviceName := strings.TrimSuffix(e.Name(), ".yaml")
		configPath := filepath.Join(servicesDirPath, e.Name())

		cfg, err := config.Load[ServiceConfig](configPath)
		if err != nil {
			return ServicesConfig{}, errors.Wrapf(err, "load service config %s", e.Name())
		}

		baseURL := cfg.Host
		if baseURL != "" && !strings.Contains(baseURL, "://") {
			baseURL = "http://" + baseURL
		}
		if _, err := url.Parse(baseURL); err != nil {
			return ServicesConfig{}, errors.Wrapf(err, "parse host %s", cfg.Host)
		}

		for endpointName, endpoint := range cfg.Endpoints {
			key := serviceName + "/" + endpointName
			if _, ok := out.Endpoints[key]; ok {
				return ServicesConfig{}, errors.Errorf("duplicate service endpoint %s", key)
			}
			fullURL := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(endpoint.Path, "/")

			retries := endpoint.RetriesNum
			if retries == 0 {
				retries = 1
			}
			initialInterval := endpoint.InitialIntervalMs
			if initialInterval == 0 {
				initialInterval = 100
			}
			backoff := endpoint.BackoffMultiplier
			if backoff == 0 {
				backoff = 2.0
			}
			maxInterval := endpoint.MaxIntervalMs
			if maxInterval == 0 {
				maxInterval = 5000
			}

			out.Endpoints[key] = EndpointDef{
				URL:               fullURL,
				TimeoutMs:         endpoint.TimeoutMs,
				RetriesNum:        retries,
				InitialIntervalMs: initialInterval,
				BackoffMultiplier: backoff,
				MaxIntervalMs:     maxInterval,
			}
		}
	}

	return out, nil
}

func loadGraphsFromDir(graphsDirPath string) (map[string]GraphConfig, error) {
	out := make(map[string]GraphConfig)

	entries, err := os.ReadDir(graphsDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, errors.Wrap(err, "read graphs dir")
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		graphName := strings.TrimSuffix(e.Name(), ".yaml")
		configPath := filepath.Join(graphsDirPath, e.Name())

		cfg, err := config.Load[GraphFileConfig](configPath)
		if err != nil {
			return nil, errors.Wrapf(err, "load graph config %s", e.Name())
		}

		if _, ok := out[graphName]; ok {
			return nil, errors.Errorf("duplicate graph %s", graphName)
		}

		nodes := make(map[string]NodeConfig)
		for nodeName, nc := range cfg.Nodes {
			nodes[nodeName] = NodeConfig{
				ID:           nodeName,
				Type:         nc.Type,
				EndpointID:   nc.EndpointID,
				Dependencies: nc.Dependencies,
			}
		}

		inputNode := defaultInputNode
		if cfg.InputNode != "" {
			inputNode = cfg.InputNode
		}
		outputNode := defaultOutputNode
		if cfg.OutputNode != "" {
			outputNode = cfg.OutputNode
		}

		out[graphName] = GraphConfig{
			Nodes:           nodes,
			InputNode:       inputNode,
			OutputNode:      outputNode,
			FailureStrategy: cfg.FailureStrategy,
			CustomStrategy:  cfg.CustomStrategy,
			TimeoutMs:       cfg.TimeoutMs,
		}
	}

	return out, nil
}
