package graphnodes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"go.uber.org/zap"
)

type CacheRepo interface {
	Get(ctx context.Context, key string) (graph.NodeExecutorResponse, bool, error)
	Set(ctx context.Context, key string, resp graph.NodeExecutorResponse, ttl time.Duration) error
}

type CachingExecutor struct {
	inner     graph.NodeExecutor
	nodeID    graph.NodeID
	graphHash string
	ttl       time.Duration
	repo      CacheRepo
}

func NewCachingExecutor(
	inner graph.NodeExecutor,
	nodeID graph.NodeID,
	graphHash string,
	ttl time.Duration,
	repo CacheRepo,
) *CachingExecutor {
	return &CachingExecutor{
		inner:     inner,
		nodeID:    nodeID,
		graphHash: graphHash,
		ttl:       ttl,
		repo:      repo,
	}
}

func (e *CachingExecutor) Run(
	ctx context.Context,
	logger *zap.Logger,
	req graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	if req.EndpointOverride != nil {
		return e.inner.Run(ctx, logger, req)
	}

	key := e.buildKey(req.Items)

	cached, ok, err := e.repo.Get(ctx, key)
	if err != nil {
		logger.Warn("cache get failed, falling back to executor", zap.Error(err))
	} else if ok {
		if cached.Meta == nil {
			cached.Meta = make(map[string]any)
		}
		cached.Meta["cached"] = true
		return cached, nil
	}

	resp, err := e.inner.Run(ctx, logger, req)
	if err != nil {
		return resp, err
	}

	if setErr := e.repo.Set(ctx, key, resp, e.ttl); setErr != nil {
		logger.Warn("cache set failed", zap.Error(setErr))
	}

	return resp, nil
}

func (e *CachingExecutor) buildKey(items map[graph.ItemID]graph.Item) string {
	itemsHash := hashItems(items)
	return fmt.Sprintf("shortcut:node:%s:%s:%s", e.graphHash, e.nodeID, itemsHash)
}

func hashItems(items map[graph.ItemID]graph.Item) string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)

	type entry struct {
		Key  string `json:"k"`
		Data []byte `json:"d"`
	}
	entries := make([]entry, 0, len(keys))
	for _, k := range keys {
		entries = append(entries, entry{Key: k, Data: items[graph.ItemID(k)].Data})
	}

	b, _ := json.Marshal(entries)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
