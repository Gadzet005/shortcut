package trace

import "context"

type Repo interface {
	Save(ctx context.Context, t Trace) error
	GetByRequestID(ctx context.Context, requestID RequestID) (Trace, error)
}
