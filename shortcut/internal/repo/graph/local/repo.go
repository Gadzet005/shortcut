package graphrepolocal

import (
	"github.com/Gadzet005/shortcut/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/shortcut/pkg/shortcut"
)

func NewLocalRepo(cfg map[graph.ID]graph.Graph) *LocalRepo {
	return &LocalRepo{
		graphs: cfg,
	}
}

type LocalRepo struct {
	graphs map[graph.ID]graph.Graph
}

func (s *LocalRepo) GetGraph(id graph.ID) (graph.Graph, error) {
	curGraph, ok := s.graphs[id]
	if !ok {
		return graph.Graph{}, shortcut.ErrItemNotFound
	}

	return curGraph, nil
}
