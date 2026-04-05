package rungraph

import (
	"context"
	"encoding/json"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
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
) useCase {
	return useCase{
		client:        client,
		namespaceRepo: namespaceRepo,
		logger:        logger,
	}
}

type useCase struct {
	client        *resty.Client
	logger        *zap.Logger
	namespaceRepo graph.NamespaceRepo
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

	rawHTTPRequest, err := json.Marshal(input)
	if err != nil {
		return shortcutapi.HttpResponse{}, errors.Wrap(err, "marshal http request")
	}

	items := map[graph.ItemID]graph.Item{
		httpRequestItemID: {Data: rawHTTPRequest},
	}

	resp, err := g.Run(ctx, u.logger, items)
	if err != nil {
		var nodeErr *graph.NodeError
		if errors.As(err, &nodeErr) {
			return nodeErrorToHTTPResponse(nodeErr)
		}
		return shortcutapi.HttpResponse{}, errors.Wrap(err, "run graph")
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

func getGraphID(namespace graph.Namespace, path string, method string) (graph.ID, error) {
	for _, route := range namespace.HTTPRoutes {
		if route.Path == path && route.Method == method {
			return route.GraphID, nil
		}
	}
	return "", graph.ErrNotFound
}
