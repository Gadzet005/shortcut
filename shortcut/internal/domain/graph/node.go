package graph

import (
	"context"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	InputNodeID   NodeID = "input"
	DefaultItemID ItemID = "default"
)

type NodeID string

func (i NodeID) String() string {
	return string(i)
}

type ItemID string

func (i ItemID) String() string {
	return string(i)
}

type Item struct {
	Data []byte
}

type Dependency struct {
	NodeID          NodeID
	ItemID          ItemID
	OverridenItemID ItemID
}

type RunNodeRequest struct {
	Client *resty.Client
	Items  map[ItemID]Item
}

type RunNodeResponse struct {
	Items map[ItemID]Item
}

type Node interface {
	Run(
		ctx context.Context,
		logger *zap.Logger,
		req RunNodeRequest,
	) (RunNodeResponse, error)
	ID() NodeID
	Dependencies() []Dependency
}
