package graph

import (
	"context"
	"strings"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	InputNodeID     NodeID   = "input"
	DefaultItemName ItemName = "default"
)

var DefaultItemID = ItemID{
	NodeID: InputNodeID,
	Name:   DefaultItemName,
}

type NodeID string

func (i NodeID) String() string {
	return string(i)
}

type ItemName string

func (i ItemName) String() string {
	return string(i)
}

type ItemID struct {
	NodeID NodeID
	Name   ItemName
}

func (i ItemID) String() string {
	var builder strings.Builder
	builder.Grow(len(i.NodeID) + len(i.Name) + 1)
	builder.WriteString(i.NodeID.String())
	builder.WriteString(".")
	builder.WriteString(i.Name.String())
	return builder.String()
}

type Item struct {
	Data []byte
}

type RunNodeRequest struct {
	Client *resty.Client
	Items  map[ItemID]Item
}

type RunNodeResponse struct {
	Items map[ItemName]Item
}

type Node interface {
	Run(
		ctx context.Context,
		logger *zap.Logger,
		req RunNodeRequest,
	) (RunNodeResponse, error)
	ID() NodeID
	Dependencies() []ItemID
}
