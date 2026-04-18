package graph

import (
	"context"

	"go.uber.org/zap"
)

type Graph interface {
	Run(
		ctx context.Context,
		logger *zap.Logger,
		items map[ItemID]Item,
	) (map[ItemID]Item, error)
	TryRevert(
		ctx context.Context,
		logger *zap.Logger,
		requestID string,
	) (bool, error)
}

type ID string

func (i ID) String() string {
	return string(i)
}

type NamespaceID string

func (n NamespaceID) String() string {
	return string(n)
}

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
	NodeID         NodeID
	ItemID         ItemID
	OverrideItemID ItemID
}

type Node struct {
	ID           NodeID
	Dependencies []Dependency
	Executor     NodeExecutor
}

type NodeExecutorRequest struct {
	Items map[ItemID]Item
}

type NodeExecutorResponse struct {
	Items map[ItemID]Item
	Meta  map[string]any
}

type NodeExecutor interface {
	Run(
		ctx context.Context,
		logger *zap.Logger,
		req NodeExecutorRequest,
	) (NodeExecutorResponse, error)
}

type Namespace struct {
	ID         NamespaceID
	Graphs     map[ID]Graph
	HTTPRoutes map[string]HTTPRoute
}

type HTTPRoute struct {
	Path    string
	Method  string
	GraphID ID
}

type NamespaceRepo interface {
	GetNamespace(id NamespaceID) (Namespace, error)
}
