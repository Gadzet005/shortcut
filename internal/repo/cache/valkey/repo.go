package cachevalkey

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/valkey-io/valkey-go"
)

func NewRepo(client valkey.Client) *Repo {
	return &Repo{client: client}
}

type Repo struct {
	client valkey.Client
}

type cachedResponse struct {
	Items map[string][]byte `json:"items"`
	Meta  map[string]any    `json:"meta,omitempty"`
}

func (r *Repo) Get(ctx context.Context, key string) (graph.NodeExecutorResponse, bool, error) {
	result := r.client.Do(ctx, r.client.B().Get().Key(key).Build())
	if valkey.IsValkeyNil(result.Error()) {
		return graph.NodeExecutorResponse{}, false, nil
	}

	data, err := result.AsBytes()
	if err != nil {
		return graph.NodeExecutorResponse{}, false, err
	}

	var cached cachedResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		return graph.NodeExecutorResponse{}, false, err
	}

	resp := graph.NodeExecutorResponse{
		Items: make(map[graph.ItemID]graph.Item, len(cached.Items)),
		Meta:  cached.Meta,
	}
	for k, v := range cached.Items {
		resp.Items[graph.ItemID(k)] = graph.Item{Data: v}
	}
	return resp, true, nil
}

func (r *Repo) Set(ctx context.Context, key string, resp graph.NodeExecutorResponse, ttl time.Duration) error {
	cached := cachedResponse{
		Items: make(map[string][]byte, len(resp.Items)),
		Meta:  resp.Meta,
	}
	for k, v := range resp.Items {
		cached.Items[string(k)] = v.Data
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	var cmd valkey.Completed
	if ttl > 0 {
		cmd = r.client.B().Set().Key(key).Value(string(data)).Ex(ttl).Build()
	} else {
		cmd = r.client.B().Set().Key(key).Value(string(data)).Build()
	}
	return r.client.Do(ctx, cmd).Error()
}
