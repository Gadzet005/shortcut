package graphrepostub

import (
	"net/url"
	"os"

	"github.com/Gadzet005/shortcut/shortcut/internal/domain/graph"
)

var (
	echoPath         = "/echo"
	echoDependencies = []graph.Dependency{
		{
			NodeID:          graph.InputNodeID,
			ItemID:          graph.DefaultItemID,
			OverridenItemID: "req",
		},
	}

	echo1Node       = graph.NodeID("echo1")
	echo2Node       = graph.NodeID("echo2")
	sumNode         = graph.NodeID("sum")
	sumDependencies = []graph.Dependency{
		{
			NodeID:          echo1Node,
			ItemID:          "resp",
			OverridenItemID: "a",
		},
		{
			NodeID:          echo2Node,
			ItemID:          "resp",
			OverridenItemID: "b",
		},
	}
	sumPath = "/sum"
)

func NewStubRepo() *StubRepo {
	return &StubRepo{
		echo1Backend: getBackendURL("MOCK_SERVICE_8080", "http://localhost:8081"),
		echo2Backend: getBackendURL("MOCK_SERVICE_8081", "http://localhost:8082"),
		sumBackend:   getBackendURL("MOCK_SERVICE_8082", "http://localhost:8083"),
	}
}

func getBackendURL(envKey, defaultURL string) url.URL {
	urlStr := os.Getenv(envKey)
	if urlStr == "" {
		urlStr = defaultURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		parsedURL, err = url.Parse(defaultURL)
		if err != nil {
			return url.URL{}
		}
	}

	return *parsedURL
}

type StubRepo struct {
	echo1Backend url.URL
	echo2Backend url.URL
	sumBackend   url.URL
}

func (s *StubRepo) GetGraph(id graph.ID) (graph.Graph, error) {
	return graph.Graph{
		ID: id,
		Nodes: map[graph.NodeID]graph.Node{
			echo1Node: graph.NewEndpoint(
				echo1Node,
				echoDependencies,
				graph.Backend{
					BaseURL: &s.echo1Backend,
				},
				echoPath,
				[]graph.ItemID{"resp"},
			),
			echo2Node: graph.NewEndpoint(
				echo2Node,
				echoDependencies,
				graph.Backend{
					BaseURL: &s.echo2Backend,
				},
				echoPath,
				[]graph.ItemID{"resp"},
			),
			sumNode: graph.NewEndpoint(
				sumNode,
				sumDependencies,
				graph.Backend{
					BaseURL: &s.sumBackend,
				},
				sumPath,
				[]graph.ItemID{"resp"},
			),
		},
	}, nil
}
