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

func RequestIDFromContext(ctx context.Context) string {
	c, ok := GetCollector(ctx)
	if !ok {
		return ""
	}
	return c.RequestID().String()
}
