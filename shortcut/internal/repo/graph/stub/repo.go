package graphrepostub

import "github.com/Gadzet005/shortcut/shortcut/internal/domain/graph"

func NewStubRepo() *StubRepo {
	return &StubRepo{}
}

type StubRepo struct{}

func (s *StubRepo) GetGraph(id graph.GraphID) (graph.Graph, error) {
	return graph.Graph{
		ID: id,
	}, nil
}
