package rungraph

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
)

type Request struct {
	GraphID graph.ID
	Data    []byte
}

type Response struct {
	Data []byte
}

type UseCase interface {
	RunGraph(ctx context.Context, input Request) (Response, error)
}
