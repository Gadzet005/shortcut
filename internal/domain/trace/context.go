package trace

import "context"

type collectorKeyType struct{}

var collectorKey = collectorKeyType{}

func WithCollector(ctx context.Context, c *Collector) context.Context {
	return context.WithValue(ctx, collectorKey, c)
}

func GetCollector(ctx context.Context) (*Collector, bool) {
	c, ok := ctx.Value(collectorKey).(*Collector)
	return c, ok
}
