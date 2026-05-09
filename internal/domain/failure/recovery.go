package failure

import "context"

type Recovery interface {
	Revert(ctx context.Context, namespaceID, graphID, requestID string, visitedNodes []string) (bool, error)
	Retry(ctx context.Context, namespaceID, graphID string, method, path string, body []byte) (bool, error)
	Finish(ctx context.Context, namespaceID, graphID, requestID string, visitedNodes []string, method, path string, body []byte) (bool, error)
}
