package graphnodes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNextInterval(t *testing.T) {
	tests := []struct {
		name       string
		current    time.Duration
		multiplier float64
		max        time.Duration
		want       time.Duration
	}{
		{
			name:       "doubles without cap",
			current:    100 * time.Millisecond,
			multiplier: 2.0,
			max:        0,
			want:       200 * time.Millisecond,
		},
		{
			name:       "capped at max",
			current:    400 * time.Millisecond,
			multiplier: 2.0,
			max:        500 * time.Millisecond,
			want:       500 * time.Millisecond,
		},
		{
			name:       "already at max — stays capped",
			current:    500 * time.Millisecond,
			multiplier: 2.0,
			max:        500 * time.Millisecond,
			want:       500 * time.Millisecond,
		},
		{
			name:       "fractional multiplier",
			current:    1 * time.Second,
			multiplier: 1.5,
			max:        0,
			want:       1500 * time.Millisecond,
		},
		{
			name:       "zero current",
			current:    0,
			multiplier: 2.0,
			max:        0,
			want:       0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nextInterval(tc.current, tc.multiplier, tc.max)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestHTTPStatusToErrorCode(t *testing.T) {
	tests := []struct {
		status int
		want   graph.ErrorCode
	}{
		{400, graph.ErrCodeBadRequest},
		{401, graph.ErrCodeUnauthorized},
		{403, graph.ErrCodeForbidden},
		{404, graph.ErrCodeNotFound},
		{500, graph.ErrCodeInternal},
		{503, graph.ErrCodeInternal},
		{422, graph.ErrCodeInternal},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tc.want, httpStatusToErrorCode(tc.status))
		})
	}
}

func TestWithRetry(t *testing.T) {
	logger := zap.NewNop()

	t.Run("success on first attempt — retry_count is 0", func(t *testing.T) {
		calls := 0
		resp, err := withRetry(context.Background(), logger, Endpoint{RetriesNum: 3}, func(_ context.Context) (graph.NodeExecutorResponse, error) {
			calls++
			return graph.NodeExecutorResponse{}, nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, calls)
		require.Equal(t, 0, resp.Meta["retry_count"])
	})

	t.Run("retries on internal error, succeeds on second attempt", func(t *testing.T) {
		calls := 0
		resp, err := withRetry(context.Background(), logger, Endpoint{RetriesNum: 3, InitialInterval: time.Millisecond}, func(_ context.Context) (graph.NodeExecutorResponse, error) {
			calls++
			if calls < 2 {
				return graph.NodeExecutorResponse{}, &graph.NodeError{Code: graph.ErrCodeInternal}
			}
			return graph.NodeExecutorResponse{}, nil
		})
		require.NoError(t, err)
		require.Equal(t, 2, calls)
		require.Equal(t, 1, resp.Meta["retry_count"])
	})

	t.Run("no retry on non-internal NodeError", func(t *testing.T) {
		calls := 0
		_, err := withRetry(context.Background(), logger, Endpoint{RetriesNum: 3, InitialInterval: time.Millisecond}, func(_ context.Context) (graph.NodeExecutorResponse, error) {
			calls++
			return graph.NodeExecutorResponse{}, &graph.NodeError{Code: graph.ErrCodeBadRequest}
		})
		require.Error(t, err)
		require.Equal(t, 1, calls)

		var nodeErr *graph.NodeError
		require.ErrorAs(t, err, &nodeErr)
		require.Equal(t, graph.ErrCodeBadRequest, nodeErr.Code)
	})

	t.Run("exhausts all retries — returns last error", func(t *testing.T) {
		calls := 0
		sentinel := errors.New("persistent failure")
		_, err := withRetry(context.Background(), logger, Endpoint{RetriesNum: 2, InitialInterval: time.Millisecond}, func(_ context.Context) (graph.NodeExecutorResponse, error) {
			calls++
			return graph.NodeExecutorResponse{}, sentinel
		})
		require.ErrorIs(t, err, sentinel)
		require.Equal(t, 3, calls) // 1 initial + 2 retries
	})

	t.Run("context cancelled during backoff", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		_, err := withRetry(ctx, logger, Endpoint{RetriesNum: 5, InitialInterval: 10 * time.Second}, func(_ context.Context) (graph.NodeExecutorResponse, error) {
			calls++
			cancel()
			return graph.NodeExecutorResponse{}, &graph.NodeError{Code: graph.ErrCodeInternal}
		})
		require.Error(t, err)
		require.Equal(t, 1, calls)
	})

	t.Run("zero retries — calls fn exactly once", func(t *testing.T) {
		calls := 0
		sentinel := errors.New("fail")
		_, err := withRetry(context.Background(), logger, Endpoint{RetriesNum: 0}, func(_ context.Context) (graph.NodeExecutorResponse, error) {
			calls++
			return graph.NodeExecutorResponse{}, sentinel
		})
		require.ErrorIs(t, err, sentinel)
		require.Equal(t, 1, calls)
	})
}
