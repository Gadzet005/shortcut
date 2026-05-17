package failurerecovery

import (
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/pkg/errors"
)

func NewGraphInfoResolver(namespaceRepo graph.NamespaceRepo) *graphInfoResolver {
	return &graphInfoResolver{namespaceRepo: namespaceRepo}
}

type graphInfoResolver struct {
	namespaceRepo graph.NamespaceRepo
}

func (r *graphInfoResolver) GetFailureSteps(namespaceID graph.NamespaceID, graphID graph.ID) ([]graph.FailureStep, graph.FailureStrategy, error) {
	ns, err := r.namespaceRepo.GetNamespace(namespaceID)
	if err != nil {
		return nil, graph.AbsentFailureStrategy, errors.Wrap(err, "get namespace")
	}
	info, ok := ns.GraphInfo[graphID]
	if !ok {
		return nil, graph.AbsentFailureStrategy, errors.Errorf("graph %s not found in namespace %s", graphID, namespaceID)
	}
	return info.FailureSteps, info.FailureStrategy, nil
}

func (r *graphInfoResolver) GetFailureStrategy(namespaceID graph.NamespaceID, graphID graph.ID) (graph.FailureStrategy, error) {
	ns, err := r.namespaceRepo.GetNamespace(namespaceID)
	if err != nil {
		return "", errors.Wrap(err, "get namespace")
	}
	info, ok := ns.GraphInfo[graphID]
	if !ok {
		return "", errors.Errorf("graph %s not found in namespace %s", graphID, namespaceID)
	}
	return info.FailureStrategy, nil
}
