package graphconfig

import (
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	graphnodes "github.com/Gadzet005/shortcut/internal/domain/graph/nodes"
	"github.com/Gadzet005/shortcut/pkg/containers/slices"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/go-resty/resty/v2"
)

type warnUserFunc func(s string)

func Convert(cfg Config, warnUser warnUserFunc, client *resty.Client) (map[graph.NamespaceID]graph.Namespace, error) {
	namespaces := make(map[graph.NamespaceID]graph.Namespace)

	for namespaceIDStr, nsCfg := range cfg.Namespaces {
		namespaceID := graph.NamespaceID(namespaceIDStr)
		ns, err := convertNamespace(nsCfg, namespaceID, warnUser, client)
		if err != nil {
			return nil, errors.Wrapf(err, "convert namespace %s", namespaceIDStr)
		}
		namespaces[namespaceID] = ns
	}

	return namespaces, nil
}

func convertNamespace(
	ns NamespaceConfig,
	namespaceID graph.NamespaceID,
	warnUser warnUserFunc,
	client *resty.Client,
) (graph.Namespace, error) {
	nsOut := graph.Namespace{
		ID:         namespaceID,
		Graphs:     make(map[graph.ID]graph.Graph),
		HTTPRoutes: make(map[string]graph.HTTPRoute),
	}

	for routeName, r := range ns.HTTPRoutes {
		nsOut.HTTPRoutes[routeName] = graph.HTTPRoute{
			Path:    r.Path,
			Method:  r.Method,
			GraphID: graph.ID(r.Graph),
		}
	}

	for graphName, gCfg := range ns.Graphs {
		_, ok := graph.ParseFailureStrategy(gCfg.FailureStrategy)
		if !ok {
			warnUser("Failure strategy not specified for graph " + graphName + ". Ignore strategy will be used by default.")
		}

		nodesMap, err := convertGraphNodes(gCfg, ns.Services, namespaceID, client)
		if err != nil {
			return graph.Namespace{}, errors.Wrapf(err, "graph %s", graphName)
		}

		g, err := graph.NewGraph(nodesMap, graph.NodeID(gCfg.InputNode), graph.NodeID(gCfg.OutputNode), time.Duration(gCfg.TimeoutMs)*time.Millisecond)
		if err != nil {
			return graph.Namespace{}, errors.Wrapf(err, "build graph %s", graphName)
		}

		nsOut.Graphs[graph.ID(graphName)] = g
	}

	return nsOut, nil
}

func convertGraphNodes(
	gCfg GraphConfig,
	services ServicesConfig,
	namespaceID graph.NamespaceID,
	client *resty.Client,
) (map[graph.NodeID]graph.Node, error) {
	nodesMap := make(map[graph.NodeID]graph.Node)

	nodesMap[graph.NodeID(gCfg.InputNode)] = graph.Node{
		ID:           graph.NodeID(gCfg.InputNode),
		Dependencies: nil,
		Executor:     graphnodes.NewTransparentNodeExecutor(),
	}

	for nodeName, nCfg := range gCfg.Nodes {
		node, err := convertNode(nCfg, services, namespaceID, client)
		if err != nil {
			return nil, errors.Wrapf(err, "node %s", nodeName)
		}
		node.ID = graph.NodeID(nCfg.ID)
		nodesMap[graph.NodeID(nodeName)] = node
	}

	return nodesMap, nil
}

func convertNode(
	nCfg NodeConfig,
	services ServicesConfig,
	namespaceID graph.NamespaceID,
	client *resty.Client,
) (graph.Node, error) {
	deps := slices.Map(nCfg.Dependencies, func(d DependencyConfig) graph.Dependency {
		return graph.Dependency{
			NodeID:         graph.NodeID(d.NodeID),
			ItemID:         graph.ItemID(d.ItemID),
			OverrideItemID: graph.ItemID(d.OverridenItemID),
		}
	})

	switch nCfg.Type {
	case NodeTypeTransparent:
		return graph.Node{
			ID:           graph.NodeID(nCfg.ID),
			Dependencies: deps,
			Executor:     graphnodes.NewTransparentNodeExecutor(),
		}, nil
	case NodeTypeDefault, NodeType(""):
		// endpoint node
	default:
		return graph.Node{}, errors.Errorf("unknown node type %q in namespace %s", nCfg.Type, namespaceID)
	}

	ep, ok := services.Endpoints[nCfg.EndpointID]
	if !ok {
		return graph.Node{}, errors.Errorf("endpoint %s not found in namespace %s", nCfg.EndpointID, namespaceID)
	}

	executor := graphnodes.NewDefaultNodeExecutor(client, graphnodes.Endpoint{
		URL:               ep.URL,
		Timeout:           time.Duration(ep.TimeoutMs) * time.Millisecond,
		RetriesNum:        ep.RetriesNum,
		InitialInterval:   time.Duration(ep.InitialIntervalMs) * time.Millisecond,
		BackoffMultiplier: ep.BackoffMultiplier,
		MaxInterval:       time.Duration(ep.MaxIntervalMs) * time.Millisecond,
	})
	return graph.Node{
		ID:           graph.NodeID(nCfg.ID),
		Dependencies: deps,
		Executor:     executor,
	}, nil
}
