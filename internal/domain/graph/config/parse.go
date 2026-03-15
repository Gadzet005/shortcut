package graphconfig

import (
	"net/url"
	"strings"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	configutils "github.com/Gadzet005/shortcut/pkg/app/config"
	"github.com/Gadzet005/shortcut/pkg/containers/sets"
	"github.com/Gadzet005/shortcut/pkg/containers/slices"
	errorsutils "github.com/Gadzet005/shortcut/pkg/errors"
)

func ParseConfig(namespaceConfigs map[string]NamespaceConfig, warnUser func(s string)) (map[graph.ID]graph.Graph, error) {
	serviceMap := make(map[graph.ID]graph.Graph)

	for namespace, namespaceConfig := range namespaceConfigs {
		servicesConfig, existingDependencies, err := readServiceConfigs(namespaceConfig.Services)
		if err != nil {
			return nil, errorsutils.Wrapf(err, "read service configs in namespace %s", namespace)
		}

		graphConfig, err := readGraphConfigs(namespaceConfig.Graphs)
		if err != nil {
			return nil, errorsutils.Wrapf(err, "read graph configs in namespace %s", namespace)
		}

		for graphName, gc := range graphConfig {
			id := graph.ID(strings.Join([]string{namespace, graphName}, "::"))

			nodes := make(map[graph.NodeID]graph.Node)

			for _, nodeConfig := range gc.Nodes {
				node, err := readNodeConfig(nodeConfig, servicesConfig, existingDependencies, namespace)
				if err != nil {
					return nil, errorsutils.Wrapf(err, "read node %s config in namespace %s", nodeConfig.EndpointID, namespace)
				}

				nodes[graph.NodeID(nodeConfig.EndpointID)] = node
			}

			parsedFailureStrategy, ok := graph.ParseFailureStrategy(gc.FailureStrategy)
			if !ok {
				warnUser("Failure strategy not specified for graph " + graphName + ". Ignore strategy will be used by default.")
			}

			retGraph := graph.Graph{
				ID:              graph.ID(id),
				Nodes:           nodes,
				FailureStrategy: parsedFailureStrategy,
			}

			numRetNodes := 0

			for _, g := range retGraph.Nodes {
				if slices.Contains(g.ReturnIDs(), graph.DefaultItemID) {
					numRetNodes += 1
				}
			}

			if numRetNodes > 1 {
				return nil, errorsutils.Wrapf(err, "found more than one ret node in graph %s, namespace %s", retGraph.ID, namespace)
			}

			_, err := graph.TopSort(retGraph)
			if err != nil {
				return nil, errorsutils.Wrapf(err, "top sort error %s", namespace)
			}

			serviceMap[id] = retGraph
		}
	}

	return serviceMap, nil
}

func readServiceConfigs(configsPath []string) (map[string]graph.Node, *sets.Set[graph.Dependency], error) {
	serviceConfigs := make(map[string]graph.Node)
	existingDependencies := sets.New[graph.Dependency]()
	existingDependencies.Add(graph.Dependency{
		NodeID: graph.InputNodeID,
		ItemID: graph.DefaultItemID,
	})

	for _, configPath := range configsPath {
		serviceName := getFilenameFromPath(configPath)

		_, ok := serviceConfigs[serviceName]
		if ok {
			return nil, nil, errorsutils.Errorf("Found duplicate service with name %s", serviceName)
		}

		config, err := configutils.Load[ServiceConfig](configPath)
		if err != nil {
			return nil, nil, errorsutils.WrapFail(err, "load service config")
		}

		for endpointName, endpoint := range config.Endpoints {
			nodeID := graph.NodeID(strings.Join([]string{serviceName, endpointName}, "/"))
			for _, dep := range endpoint.ResultIDs {
				existingDependencies.Add(graph.Dependency{
					NodeID: nodeID,
					ItemID: graph.ItemID(dep),
				})
			}

			baseURL, err := url.Parse(config.Host)
			if err != nil {
				return nil, nil, errorsutils.Wrapf(err, "parse url: %s", config.Host)
			}

			serviceConfigs[strings.Join([]string{serviceName, endpointName}, "/")] = graph.NewEndpoint(
				nodeID,
				[]graph.Dependency{},
				graph.Backend{
					BaseURL: baseURL,
				},
				endpoint.Path,
				slices.Map(endpoint.ResultIDs, func(s string) graph.ItemID {
					return graph.ItemID(s)
				}),
			)
		}
	}

	return serviceConfigs, existingDependencies, nil
}

func readGraphConfigs(configsPath []string) (map[string]GraphConfig, error) {
	graphConfigs := make(map[string]GraphConfig)

	for _, configPath := range configsPath {
		graphName := getFilenameFromPath(configPath)

		_, ok := graphConfigs[graphName]
		if ok {
			return nil, errorsutils.Errorf("Found duplicate graph with name %s", graphName)
		}

		config, err := configutils.Load[GraphConfig](configPath)
		if err != nil {
			return nil, errorsutils.WrapFail(err, "load graph config")
		}

		graphConfigs[graphName] = config
	}

	return graphConfigs, nil
}

func readNodeConfig(nodeConfig NodeConfig, servicesConfig map[string]graph.Node, existingDependencies *sets.Set[graph.Dependency], namespace string) (graph.Node, error) {
	node, ok := servicesConfig[nodeConfig.EndpointID]
	if !ok {
		return nil, errorsutils.Errorf("node with id %s not found in namespace %s", nodeConfig.EndpointID, namespace)
	}

	var dependencies []graph.Dependency
	for node, deps := range nodeConfig.Dependencies {
		newDependencies := slices.Map(deps, func(dependencyName string) graph.Dependency {
			return graph.Dependency{
				NodeID: graph.NodeID(node),
				ItemID: graph.ItemID(dependencyName),
			}
		})

		for _, dep := range newDependencies {
			if !existingDependencies.Contains(dep) {
				return nil, errorsutils.Errorf("dependency %v not found in namespace %s", dep, namespace)
			}
		}

		dependencies = append(dependencies, newDependencies...)
	}

	return node.WithDependencies(dependencies), nil
}

func getFilenameFromPath(path string) string {
	parts := strings.Split(path, "/")
	serviceName := strings.Split(parts[len(parts)-1], ".")[0]
	return serviceName
}
