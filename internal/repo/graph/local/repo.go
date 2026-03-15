package graphrepolocal

import (
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/pkg/shortcut"
)

func NewLocalRepo(cfg map[graph.ID]graph.Graph) *localRepo {
	return &localRepo{
		graphs: cfg,
	}
}

type localRepo struct {
	graphs map[graph.ID]graph.Graph
}

func (s *localRepo) GetGraph(id graph.ID) (graph.Graph, error) {
	curGraph, ok := s.graphs[id]
	if !ok {
		return graph.Graph{}, shortcut.ErrItemNotFound
	}

	return curGraph, nil
}
