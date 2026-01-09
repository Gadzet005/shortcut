package graph

import (
	"context"
	"net/url"

	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
	multipartutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/multipart"
	"go.uber.org/zap"
)

func NewEndpoint(
	id NodeID,
	dependencies []Dependency,
	backend Backend,
	path string,
) Endpoint {
	return Endpoint{
		id:           id,
		dependencies: dependencies,
		backend:      backend,
		path:         path,
	}
}

type Endpoint struct {
	id           NodeID
	dependencies []Dependency
	backend      Backend
	path         string
}

type Backend struct {
	BaseURL *url.URL
}

func (e Endpoint) URL() string {
	return e.backend.BaseURL.JoinPath(e.path).String()
}

func (e Endpoint) ID() NodeID {
	return e.id
}

func (e Endpoint) Dependencies() []Dependency {
	return e.dependencies
}

func (e Endpoint) Run(
	ctx context.Context,
	logger *zap.Logger,
	req RunNodeRequest,
) (RunNodeResponse, error) {
	formData := make(map[string]string, len(req.Items))
	for id, item := range req.Items {
		formData[id.String()] = string(item.Data)
	}

	resp, err := req.Client.R().
		SetContext(ctx).
		SetFormData(formData).
		SetDoNotParseResponse(true).
		Post(e.URL())
	if err != nil {
		return RunNodeResponse{}, errorsutils.WrapFail(err, "make request")
	}
	if !resp.IsSuccess() {
		return RunNodeResponse{}, errorsutils.Errorf(
			"request failed with status code %d and body %s",
			resp.StatusCode(), resp.String(),
		)
	}

	body := resp.RawResponse.Body
	defer body.Close()

	data, err := multipartutils.ReadMultipartData(resp.Header(), body)
	if err != nil {
		return RunNodeResponse{}, errorsutils.WrapFail(err, "read multipart data")
	}

	items := make(map[ItemID]Item, len(data))
	for id, item := range data {
		items[ItemID(id)] = Item{Data: item}
	}

	return RunNodeResponse{Items: items}, nil
}
