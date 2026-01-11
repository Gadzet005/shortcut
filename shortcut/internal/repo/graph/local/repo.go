package graphrepolocal

import (
	"sync"

	"github.com/Gadzet005/shortcut/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/shortcut/pkg/shortcut"
)

func NewLocalRepo(cfg map[graph.ID]graph.Graph) *LocalRepo {
	return &LocalRepo{
		Graphs: cfg,
		Lock: sync.RWMutex{},
	}
}

type LocalRepo struct {
	Graphs map[graph.ID]graph.Graph
	Lock sync.RWMutex
}

func (s *LocalRepo) GetGraph(id graph.ID) (graph.Graph, error) {
	s.Lock.RLock()

	defer s.Lock.RUnlock()

	curGraph, ok := s.Graphs[id]
	if !ok {
		return graph.Graph{}, shortcut.ErrItemNotFound
	}

	return curGraph, nil
}
