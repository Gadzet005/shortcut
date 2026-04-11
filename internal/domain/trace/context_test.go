package trace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContextCollector(t *testing.T) {
	c := NewCollector("ctx-test")
	ctx := WithCollector(t.Context(), c)

	got, ok := GetCollector(ctx)
	require.True(t, ok)
	require.Equal(t, RequestID("ctx-test"), got.RequestID())
}

func TestContextCollector_Missing(t *testing.T) {
	_, ok := GetCollector(t.Context())
	require.False(t, ok)
}
