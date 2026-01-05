package lifecycle

import "context"

type Cycler interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
