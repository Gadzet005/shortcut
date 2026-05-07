package failurerecovery

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"go.uber.org/zap"
)

var _ failure.Recovery = (*Recovery)(nil)

func New(runGraphUC rungraph.UseCase, logger *zap.Logger) *Recovery {
	return &Recovery{
		runGraphUC: runGraphUC,
		logger:     logger.Named("failure-recovery"),
	}
}

type Recovery struct {
	runGraphUC rungraph.UseCase
	logger     *zap.Logger
}

func (r *Recovery) SetRunGraphUseCase(uc rungraph.UseCase) {
	r.runGraphUC = uc
}

func (r *Recovery) Revert(_ context.Context, namespaceID, graphID, requestID string, _ []string) (bool, error) {
	r.logger.Info("revert requested",
		zap.String("namespace_id", namespaceID),
		zap.String("graph_id", graphID),
		zap.String("request_id", requestID))
	return true, nil
}

func (r *Recovery) Retry(ctx context.Context, namespaceID, graphID, method, path string, body []byte) (bool, error) {
	return r.runOriginal(ctx, namespaceID, graphID, method, path, body)
}

func (r *Recovery) Finish(ctx context.Context, namespaceID, graphID, requestID string, _ []string, method, path string, body []byte) (bool, error) {
	ok, err := r.runOriginal(ctx, namespaceID, graphID, method, path, body)
	if err != nil {
		r.logger.Warn("finish failed",
			zap.String("namespace_id", namespaceID),
			zap.String("graph_id", graphID),
			zap.String("request_id", requestID),
			zap.Error(err))
	}
	return ok, err
}

func (r *Recovery) runOriginal(ctx context.Context, namespaceID, _ string, method, path string, body []byte) (bool, error) {
	req := decodeRequest(method, path, body)
	resp, err := r.runGraphUC.RunGraph(ctx, graph.NamespaceID(namespaceID), req)
	if err != nil {
		return false, err
	}
	return resp.StatusCode < 500, nil
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
