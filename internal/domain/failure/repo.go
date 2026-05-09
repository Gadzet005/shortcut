package failure

import (
	"context"
	"time"
)

type Repo interface {
	Save(ctx context.Context, f Failure) error
	GetByRequestID(ctx context.Context, requestID string) (Failure, error)
	Delete(ctx context.Context, requestID string) error
	ListByGraph(ctx context.Context, namespaceID, graphID string) ([]Failure, error)
	ClaimReadyBatch(ctx context.Context, batchSize int, visibilityTimeout time.Duration) ([]Failure, error)
	UpdateProgress(ctx context.Context, requestID string, numRetry int64, readyToRetryAt time.Time, status Status) error
}
