package rungraph

import (
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/stretchr/testify/require"
)

func TestParseNodeOverrides(t *testing.T) {
	tests := []struct {
		name       string
		input      []string
		wantResult map[graph.NodeID]string
		wantErr    bool
	}{
		{
			name:       "nil input returns nil",
			input:      nil,
			wantResult: nil,
		},
		{
			name:       "empty slice returns nil",
			input:      []string{},
			wantResult: nil,
		},
		{
			name:  "single valid override",
			input: []string{"my-node:localhost:9090"},
			wantResult: map[graph.NodeID]string{
				"my-node": "localhost:9090",
			},
		},
		{
			name:  "multiple valid overrides",
			input: []string{"node-a:host1:8080", "node-b:host2:9090"},
			wantResult: map[graph.NodeID]string{
				"node-a": "host1:8080",
				"node-b": "host2:9090",
			},
		},
		{
			name:    "only two parts — missing port",
			input:   []string{"my-node:localhost"},
			wantErr: true,
		},
		{
			name:    "only one part",
			input:   []string{"my-node"},
			wantErr: true,
		},
		{
			name:    "empty node name",
			input:   []string{":localhost:9090"},
			wantErr: true,
		},
		{
			name:    "empty host",
			input:   []string{"my-node::9090"},
			wantErr: true,
		},
		{
			name:    "empty port",
			input:   []string{"my-node:localhost:"},
			wantErr: true,
		},
		{
			name:    "completely empty string",
			input:   []string{""},
			wantErr: true,
		},
		{
			name:    "invalid entry among valid ones",
			input:   []string{"good-node:host:9000", "bad"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseNodeOverrides(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantResult, result)
		})
	}
}
