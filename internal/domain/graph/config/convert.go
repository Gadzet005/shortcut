package graphconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	graphnodes "github.com/Gadzet005/shortcut/internal/domain/graph/nodes"
	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"github.com/Gadzet005/shortcut/pkg/containers/slices"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/go-resty/resty/v2"
)

type warnUserFunc func(s string)

func Convert(
	cfg Config,
	warnUser warnUserFunc,
	client *resty.Client,
	cacheRepo graphnodes.CacheRepo,
	nodeMetrics graphnodes.NodeMetrics,
	cacheMetrics graphnodes.CacheMetrics,
) (map[graph.NamespaceID]graph.Namespace, error) {
	namespaces := make(map[graph.NamespaceID]graph.Namespace)

	for namespaceIDStr, nsCfg := range cfg.Namespaces {
		namespaceID := graph.NamespaceID(namespaceIDStr)
		ns, err := convertNamespace(nsCfg, namespaceID, warnUser, client, cacheRepo, nodeMetrics, cacheMetrics)
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
	cacheRepo graphnodes.CacheRepo,
	nodeMetrics graphnodes.NodeMetrics,
	cacheMetrics graphnodes.CacheMetrics,
) (graph.Namespace, error) {
	nsOut := graph.Namespace{
		ID:         namespaceID,
		Graphs:     make(map[graph.ID]graph.Graph),
		HTTPRoutes: make(map[string]graph.HTTPRoute),
		GraphInfo:  make(map[graph.ID]graph.GraphInfo),
	}

	for routeName, r := range ns.HTTPRoutes {
		nsOut.HTTPRoutes[routeName] = graph.HTTPRoute{
			Path:    r.Path,
			Method:  r.Method,
			GraphID: graph.ID(r.Graph),
		}
	}

	for graphName, gCfg := range ns.Graphs {
		strategy, ok := graph.ParseFailureStrategy(gCfg.FailureStrategy)
		if !ok {
			warnUser("Failure strategy not specified for graph " + graphName + ". Absent strategy will be used by default.")
		}

		steps, err := convertFailureSteps(gCfg.FailureSteps)
		if err != nil {
			return graph.Namespace{}, errors.Wrapf(err, "graph %s failure steps", graphName)
		}

		graphHash, err := computeGraphHash(gCfg, ns.Services)
		if err != nil {
			return graph.Namespace{}, errors.Wrapf(err, "graph %s", graphName)
		}

		nodesMap, err := convertGraphNodes(gCfg, ns.Services, namespaceID, graph.ID(graphName), client, graphHash, cacheRepo, nodeMetrics, cacheMetrics)
		if err != nil {
			return graph.Namespace{}, errors.Wrapf(err, "graph %s", graphName)
		}

		g, err := graph.NewGraph(nodesMap, graph.NodeID(gCfg.InputNode), graph.NodeID(gCfg.OutputNode), time.Duration(gCfg.TimeoutMs)*time.Millisecond)
		if err != nil {
			return graph.Namespace{}, errors.Wrapf(err, "build graph %s", graphName)
		}

		nsOut.Graphs[graph.ID(graphName)] = g
		nsOut.GraphInfo[graph.ID(graphName)] = graph.GraphInfo{
			FailureStrategy: strategy,
			FailureSteps:    steps,
		}
	}

	return nsOut, nil
}

func convertFailureSteps(cfg []FailureStepConfig) ([]graph.FailureStep, error) {
	if len(cfg) == 0 {
		return nil, nil
	}
	steps := make([]graph.FailureStep, len(cfg))
	for i, s := range cfg {
		action, ok := graph.ParseStrategyAction(s.Action)
		if !ok {
			return nil, errors.Errorf("unknown strategy action %q at step %d", s.Action, i)
		}
		steps[i] = graph.FailureStep{
			Action:             action,
			Condition:          graph.StrategyCondition(s.Condition),
			WaitBefore:         time.Duration(s.WaitBeforeMs) * time.Millisecond,
			WaitBetweenRetries: time.Duration(s.WaitBetweenRetriesMs) * time.Millisecond,
			NumAttempts:        s.NumAttempts,
		}
	}
	return steps, nil
}

func computeGraphHash(gCfg GraphConfig, services ServicesConfig) (string, error) {
	type hashInput struct {
		Graph    GraphConfig    `json:"graph"`
		Services ServicesConfig `json:"services"`
	}
	b, err := json.Marshal(hashInput{Graph: gCfg, Services: services})
	if err != nil {
		return "", errors.Wrap(err, "marshal hash input")
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func convertGraphNodes(
	gCfg GraphConfig,
	services ServicesConfig,
	namespaceID graph.NamespaceID,
	graphName graph.ID,
	client *resty.Client,
	graphHash string,
	cacheRepo graphnodes.CacheRepo,
	nodeMetrics graphnodes.NodeMetrics,
	cacheMetrics graphnodes.CacheMetrics,
) (map[graph.NodeID]graph.Node, error) {
	nodesMap := make(map[graph.NodeID]graph.Node)

	inputNodeID := graph.NodeID(gCfg.InputNode)
	inputExec := graph.NodeExecutor(graphnodes.NewTransparentNodeExecutor())
	inputExec = trace.NewTracingExecutor(inputExec, inputNodeID, string(NodeTypeTransparent), nil)
	inputExec = graphnodes.NewMetricsExecutor(inputExec, inputNodeID, namespaceID, graphName, string(NodeTypeTransparent), nodeMetrics)
	nodesMap[inputNodeID] = graph.Node{
		ID:           inputNodeID,
		Dependencies: nil,
		Executor:     inputExec,
	}

	for nodeName, nCfg := range gCfg.Nodes {
		node, err := convertNode(nCfg, services, namespaceID, client)
		if err != nil {
			return nil, errors.Wrapf(err, "node %s", nodeName)
		}
		node.ID = graph.NodeID(nCfg.ID)
		nodeID := graph.NodeID(nodeName)

		traceDeps := make([]trace.NodeDependency, len(nCfg.Dependencies))
		for i, d := range nCfg.Dependencies {
			traceDeps[i] = trace.NodeDependency{NodeID: d.NodeID}
		}

		nodeType := string(nCfg.Type)
		if nodeType == "" {
			nodeType = string(NodeTypeDefault)
		}

		if nCfg.Cache != nil && nCfg.Cache.Enabled && cacheRepo != nil {
			node.Executor = graphnodes.NewCachingExecutor(node.Executor, nodeID, graphHash, nCfg.Cache.TTL, cacheRepo, cacheMetrics)
		}
		node.Executor = trace.NewTracingExecutor(node.Executor, nodeID, nodeType, traceDeps)
		node.Executor = graphnodes.NewMetricsExecutor(node.Executor, nodeID, namespaceID, graphName, nodeType, nodeMetrics)
		nodesMap[nodeID] = node
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
	case NodeTypeDefault, NodeType(""), NodeTypeHTTPAdapter:
		// endpoint node — falls through to endpoint lookup below
	default:
		return graph.Node{}, errors.Errorf("unknown node type %q in namespace %s", nCfg.Type, namespaceID)
	}

	ep, ok := services.Endpoints[nCfg.EndpointID]
	if !ok {
		return graph.Node{}, errors.Errorf("endpoint %s not found in namespace %s", nCfg.EndpointID, namespaceID)
	}

	endpoint := graphnodes.Endpoint{
		URL:               ep.URL,
		RevertURL:         ep.RevertURL,
		Timeout:           time.Duration(ep.TimeoutMs) * time.Millisecond,
		RetriesNum:        ep.RetriesNum,
		InitialInterval:   time.Duration(ep.InitialIntervalMs) * time.Millisecond,
		BackoffMultiplier: ep.BackoffMultiplier,
		MaxInterval:       time.Duration(ep.MaxIntervalMs) * time.Millisecond,
	}

	var executor graph.NodeExecutor
	if nCfg.Type == NodeTypeHTTPAdapter {
		executor = graphnodes.NewHTTPAdapterNodeExecutor(client, endpoint)
	} else {
		executor = graphnodes.NewDefaultNodeExecutor(client, endpoint)
	}

	return graph.Node{
		ID:           graph.NodeID(nCfg.ID),
		Dependencies: deps,
		Executor:     executor,
	}, nil
}
