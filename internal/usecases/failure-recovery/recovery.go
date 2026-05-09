package failurerecovery

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
	"github.com/Gadzet005/shortcut/pkg/errors"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"go.uber.org/zap"
)

var _ failure.Recovery = (*Recovery)(nil)

func New(runGraphUC rungraph.UseCase, namespaceRepo graph.NamespaceRepo, logger *zap.Logger) *Recovery {
	return &Recovery{
		runGraphUC:    runGraphUC,
		namespaceRepo: namespaceRepo,
		logger:        logger.Named("failure-recovery"),
	}
}

type Recovery struct {
	runGraphUC    rungraph.UseCase
	namespaceRepo graph.NamespaceRepo
	logger        *zap.Logger
}

func (r *Recovery) SetRunGraphUseCase(uc rungraph.UseCase) {
	r.runGraphUC = uc
}

func (r *Recovery) SetNamespaceRepo(repo graph.NamespaceRepo) {
	r.namespaceRepo = repo
}

func (r *Recovery) Revert(ctx context.Context, namespaceID, graphID, requestID string, visitedNodes []string) (bool, error) {
	logger := r.logger.With(
		zap.String("namespace_id", namespaceID),
		zap.String("graph_id", graphID),
		zap.String("request_id", requestID),
	)

	g, err := r.lookupGraph(graph.NamespaceID(namespaceID), graph.ID(graphID))
	if err != nil {
		logger.Error("revert lookup graph failed", zap.Error(err))
		return false, err
	}

	nodeIDs := toNodeIDs(visitedNodes)
	logger.Info("revert started", zap.Int("visited_nodes", len(nodeIDs)))

	ok, err := g.TryRevert(ctx, logger, requestID, nodeIDs)
	if err != nil {
		logger.Warn("revert traversal returned error", zap.Error(err))
	}
	return ok, err
}

func (r *Recovery) Retry(ctx context.Context, namespaceID, graphID, method, path string, body []byte) (bool, error) {
	return r.runOriginal(ctx, namespaceID, graphID, method, path, body)
}

func (r *Recovery) Finish(ctx context.Context, namespaceID, graphID, requestID string, visitedNodes []string, method, path string, body []byte) (bool, error) {
	logger := r.logger.With(
		zap.String("namespace_id", namespaceID),
		zap.String("graph_id", graphID),
		zap.String("request_id", requestID),
	)

	g, err := r.lookupGraph(graph.NamespaceID(namespaceID), graph.ID(graphID))
	if err != nil {
		logger.Warn("finish lookup graph failed, falling back to full re-run", zap.Error(err))
		return r.runOriginal(ctx, namespaceID, graphID, method, path, body)
	}

	req := decodeRequest(method, path, body)
	rawHTTPRequest, marshalErr := json.Marshal(req)
	if marshalErr != nil {
		logger.Error("finish marshal http request failed", zap.Error(marshalErr))
		return false, errors.Wrap(marshalErr, "marshal http request")
	}

	items := map[graph.ItemID]graph.Item{
		graph.ItemID("http_request"): {Data: rawHTTPRequest},
	}

	logger.Info("finish started", zap.Int("visited_nodes", len(visitedNodes)))

	if _, err := g.TryFinish(ctx, logger, items, nil, toNodeIDs(visitedNodes)); err != nil {
		logger.Warn("finish traversal failed", zap.Error(err))
		return false, err
	}
	return true, nil
}

func (r *Recovery) lookupGraph(namespaceID graph.NamespaceID, graphID graph.ID) (graph.Graph, error) {
	if r.namespaceRepo == nil {
		return nil, errors.Error("namespace repo not configured")
	}
	ns, err := r.namespaceRepo.GetNamespace(namespaceID)
	if err != nil {
		return nil, errors.Wrap(err, "get namespace")
	}
	g, ok := ns.Graphs[graphID]
	if !ok {
		return nil, errors.Wrapf(graph.ErrNotFound, "graph %s in namespace %s", graphID, namespaceID)
	}
	return g, nil
}

func (r *Recovery) runOriginal(ctx context.Context, namespaceID, _ string, method, path string, body []byte) (bool, error) {
	if r.runGraphUC == nil {
		return false, errors.Error("run graph usecase not configured")
	}
	req := decodeRequest(method, path, body)
	resp, err := r.runGraphUC.RunGraph(ctx, graph.NamespaceID(namespaceID), req)
	if err != nil {
		return false, err
	}
	return resp.StatusCode < 500, nil
}

func toNodeIDs(visitedNodes []string) []graph.NodeID {
	if len(visitedNodes) == 0 {
		return nil
	}
	ids := make([]graph.NodeID, len(visitedNodes))
	for i, id := range visitedNodes {
		ids[i] = graph.NodeID(id)
	}
	return ids
}

func decodeRequest(method, path string, body []byte) shortcutapi.HttpRequest {
	if len(body) == 0 {
		return shortcutapi.HttpRequest{Method: method, Path: path, Query: url.Values{}}
	}
	var req shortcutapi.HttpRequest
	if err := json.Unmarshal(body, &req); err == nil && req.Method != "" {
		return req
	}
	return shortcutapi.HttpRequest{
		Method: method,
		Path:   path,
		Body:   body,
		Query:  url.Values{},
	}
}
