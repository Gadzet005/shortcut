package graph

import (
	"context"
	"net/url"

	errors "github.com/Gadzet005/shortcut/pkg/errors"
	multipartutils "github.com/Gadzet005/shortcut/pkg/utils/multipart"
	"go.uber.org/zap"
)

func NewEndpoint(
	id NodeID,
	dependencies []Dependency,
	backend Backend,
	path string,
	returnIDs 	 []ItemID,
) Endpoint {
	return Endpoint{
		id:           id,
		dependencies: dependencies,
		backend:      backend,
		path:         path,
		returnIDs:    returnIDs,
	}
}

type Endpoint struct {
	id           NodeID
	dependencies []Dependency
	backend      Backend
	path         string
	returnIDs 	 []ItemID
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

func (e Endpoint) WithDependencies(newDependencies []Dependency) Node {
	e.dependencies = newDependencies
	return e
}

func (e Endpoint) ReturnIDs() []ItemID {
	return e.returnIDs
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
		return RunNodeResponse{}, errors.Wrap(err, "make request")
	}
	if !resp.IsSuccess() {
		return RunNodeResponse{}, errors.Errorf(
			"request failed with status code %d and body %s",
			resp.StatusCode(), resp.String(),
		)
	}

	body := resp.RawResponse.Body
	defer body.Close()

	data, err := multipartutils.ReadMultipartData(resp.Header(), body)
	if err != nil {
		return RunNodeResponse{}, errors.Wrap(err, "read multipart data")
	}

	items := make(map[ItemID]Item, len(data))
	for id, item := range data {
		items[ItemID(id)] = Item{Data: item}
	}

	return RunNodeResponse{Items: items}, nil
}
