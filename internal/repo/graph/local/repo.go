package graphlocalrepo

import (
	"github.com/Gadzet005/shortcut/internal/domain/graph"
)

var _ graph.NamespaceRepo = &localRepo{}

func NewLocalRepo(namespaces map[graph.NamespaceID]graph.Namespace) *localRepo {
	return &localRepo{
		namespaces: namespaces,
	}
}

type localRepo struct {
	namespaces map[graph.NamespaceID]graph.Namespace
}

func (r *localRepo) GetNamespace(id graph.NamespaceID) (graph.Namespace, error) {
	namespace, ok := r.namespaces[id]
	if !ok {
		return graph.Namespace{}, graph.ErrNotFound
	}
	return namespace, nil
}
