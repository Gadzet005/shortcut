package rungraph

import (
	"context"
	"encoding/json"
	"maps"
	"net/url"
	"strings"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"github.com/Gadzet005/shortcut/pkg/errors"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	httpRequestItemID  = graph.ItemID("http_request")
	httpResponseItemID = graph.ItemID("http_response")

	contentTypeKey  = "Content-Type"
	contentTypeJSON = "application/json"
)

type UseCase interface {
	RunGraph(
		ctx context.Context,
		namespaceID graph.NamespaceID,
		input shortcutapi.HttpRequest,
	) (shortcutapi.HttpResponse, error)
}

var _ UseCase = useCase{}

func NewUseCase(
	client *resty.Client,
	logger *zap.Logger,
	namespaceRepo graph.NamespaceRepo,
	traceRepo trace.Repo,
) useCase {
	return useCase{
		client:        client,
		namespaceRepo: namespaceRepo,
		logger:        logger,
		traceRepo:     traceRepo,
	}
}

type useCase struct {
	client        *resty.Client
	logger        *zap.Logger
	namespaceRepo graph.NamespaceRepo
	traceRepo     trace.Repo
}

func (u useCase) RunGraph(
	ctx context.Context,
	namespaceID graph.NamespaceID,
	input shortcutapi.HttpRequest,
) (shortcutapi.HttpResponse, error) {
	namespace, err := u.namespaceRepo.GetNamespace(namespaceID)
	if err != nil {
		return shortcutapi.HttpResponse{}, errors.Wrap(err, "get graph")
	}

	graphID, err := getGraphID(namespace, input.Path, input.Method)
	if err != nil {
		return shortcutapi.HttpResponse{}, errors.Wrap(err, "get graph id")
	}

	g, ok := namespace.Graphs[graphID]
	if !ok {
		return shortcutapi.HttpResponse{}, errors.Wrap(graph.ErrNotFound, "graph not found")
	}

	overrides, parseErr := parseNodeOverrides(input.Query["node-rwr"])
	if parseErr != nil {
		return nodeErrorToHTTPResponse(&graph.NodeError{
			Code:    graph.ErrCodeBadRequest,
			Payload: map[string]any{"error": parseErr.Error()},
		})
	}

	cleanedInput := input
	if len(input.Query["node-rwr"]) > 0 {
		cleanedQuery := make(url.Values, len(input.Query))
		maps.Copy(cleanedQuery, input.Query)
		delete(cleanedQuery, "node-rwr")
		cleanedInput.Query = cleanedQuery
	}

	rawHTTPRequest, err := json.Marshal(cleanedInput)
	if err != nil {
		return shortcutapi.HttpResponse{}, errors.Wrap(err, "marshal http request")
	}

	items := map[graph.ItemID]graph.Item{
		httpRequestItemID: {Data: rawHTTPRequest},
	}

	start := time.Now()
	resp, runErr := g.Run(ctx, u.logger, items, overrides)
	finished := time.Now()

	u.saveTrace(ctx, start, finished, namespaceID, graphID, input, runErr)

	if runErr != nil {
		var nodeErr *graph.NodeError
		if errors.As(runErr, &nodeErr) {
			return nodeErrorToHTTPResponse(nodeErr)
		}
		return shortcutapi.HttpResponse{}, errors.Wrap(runErr, "run graph")
	}

	item, ok := resp[httpResponseItemID]
	if !ok {
		return shortcutapi.HttpResponse{}, errors.Error("http response item not found")
	}

	parsedHTTPResponse := shortcutapi.HttpResponse{}
	if err := json.Unmarshal(item.Data, &parsedHTTPResponse); err != nil {
		return shortcutapi.HttpResponse{}, errors.Wrap(err, "unmarshal http response")
	}

	return parsedHTTPResponse, nil
}

func (u useCase) saveTrace(
	ctx context.Context,
	start, finished time.Time,
	namespaceID graph.NamespaceID,
	graphID graph.ID,
	input shortcutapi.HttpRequest,
	runErr error,
) {
	collector, ok := trace.GetCollector(ctx)
	if !ok || u.traceRepo == nil {
		return
	}

	t := trace.Trace{
		RequestID:   collector.RequestID(),
		NamespaceID: namespaceID.String(),
		GraphID:     graphID.String(),
		Method:      input.Method,
		Path:        input.Path,
		StartedAt:   start,
		FinishedAt:  finished,
		DurationMs:  finished.Sub(start).Milliseconds(),
		Status:      trace.TraceStatusOK,
		NodeTraces:  collector.NodeTraces(),
	}
	if runErr != nil {
		t.Status = trace.TraceStatusError
		t.Error = runErr.Error()
	}

	if saveErr := u.traceRepo.Save(ctx, t); saveErr != nil {
		u.logger.Error("failed to save trace", zap.Error(saveErr))
	}
}

func nodeErrorToHTTPResponse(e *graph.NodeError) (shortcutapi.HttpResponse, error) {
	payloadRaw, err := json.Marshal(e.Payload)
	if err != nil {
		return shortcutapi.HttpResponse{}, errors.Wrap(err, "marshal payload")
	}
	return shortcutapi.HttpResponse{
		StatusCode: errorCodeToHTTPStatus(e.Code),
		Headers:    map[string][]string{contentTypeKey: {contentTypeJSON}},
		Body:       payloadRaw,
	}, nil
}

func errorCodeToHTTPStatus(code graph.ErrorCode) int {
	switch code {
	case graph.ErrCodeBadRequest:
		return 400
	case graph.ErrCodeUnauthorized:
		return 401
	case graph.ErrCodeForbidden:
		return 403
	case graph.ErrCodeNotFound:
		return 404
	default:
		return 500
	}
}

// parseNodeOverrides parses values of the "node-rwr" query param (format: "NODE_NAME:host:port").
func parseNodeOverrides(values []string) (map[graph.NodeID]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make(map[graph.NodeID]string, len(values))
	for _, v := range values {
		parts := strings.SplitN(v, ":", 3)
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return nil, errors.Errorf("invalid node-rwr value %q, expected NODE_NAME:host:port", v)
		}
		result[graph.NodeID(parts[0])] = parts[1] + ":" + parts[2]
	}
	return result, nil
}

func getGraphID(namespace graph.Namespace, path string, method string) (graph.ID, error) {
	for _, route := range namespace.HTTPRoutes {
		if route.Path == path && route.Method == method {
			return route.GraphID, nil
		}
	}
	return "", graph.ErrNotFound
}
