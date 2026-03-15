package topsort

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSort(t *testing.T) {
	tests := []struct {
		name     string
		g        map[string][]string
		wantOK   bool
		validate func(t *testing.T, levels [][]string)
	}{
		{
			name:   "empty graph",
			g:      nil,
			wantOK: true,
			validate: func(t *testing.T, levels [][]string) {
				require.Nil(t, levels)
			},
		},
		{
			name:   "empty map",
			g:      map[string][]string{},
			wantOK: true,
			validate: func(t *testing.T, levels [][]string) {
				require.Nil(t, levels)
			},
		},
		{
			name:   "single node with no edges",
			g:      map[string][]string{"a": {}},
			wantOK: true,
			validate: func(t *testing.T, levels [][]string) {
				require.Len(t, levels, 1)
				require.ElementsMatch(t, []string{"a"}, levels[0])
			},
		},
		{
			name:   "chain A -> B -> C",
			g:      map[string][]string{"a": {"b"}, "b": {"c"}, "c": {}},
			wantOK: true,
			validate: func(t *testing.T, levels [][]string) {
				require.Len(t, levels, 3)
				require.ElementsMatch(t, []string{"a"}, levels[0])
				require.ElementsMatch(t, []string{"b"}, levels[1])
				require.ElementsMatch(t, []string{"c"}, levels[2])
			},
		},
		{
			name:   "node only as target (not in keys)",
			g:      map[string][]string{"a": {"b"}},
			wantOK: true,
			validate: func(t *testing.T, levels [][]string) {
				require.Len(t, levels, 2)
				require.ElementsMatch(t, []string{"a"}, levels[0])
				require.ElementsMatch(t, []string{"b"}, levels[1])
			},
		},
		{
			name:   "multiple nodes at same level",
			g:      map[string][]string{"a": {"c"}, "b": {"c"}, "c": {}},
			wantOK: true,
			validate: func(t *testing.T, levels [][]string) {
				require.Len(t, levels, 2)
				require.ElementsMatch(t, []string{"a", "b"}, levels[0])
				require.ElementsMatch(t, []string{"c"}, levels[1])
			},
		},
		{
			name:   "diamond A -> B, A -> C, B -> D, C -> D",
			g:      map[string][]string{"a": {"b", "c"}, "b": {"d"}, "c": {"d"}, "d": {}},
			wantOK: true,
			validate: func(t *testing.T, levels [][]string) {
				require.Len(t, levels, 3)
				require.ElementsMatch(t, []string{"a"}, levels[0])
				require.ElementsMatch(t, []string{"b", "c"}, levels[1])
				require.ElementsMatch(t, []string{"d"}, levels[2])
			},
		},
		{
			name:   "cycle A -> B -> A",
			g:      map[string][]string{"a": {"b"}, "b": {"a"}},
			wantOK: false,
			validate: func(t *testing.T, levels [][]string) {
				require.Nil(t, levels)
			},
		},
		{
			name:   "self-loop A -> A",
			g:      map[string][]string{"a": {"a"}},
			wantOK: false,
			validate: func(t *testing.T, levels [][]string) {
				require.Nil(t, levels)
			},
		},
		{
			name:   "cycle in the middle A -> B -> C -> B",
			g:      map[string][]string{"a": {"b"}, "b": {"c"}, "c": {"b"}},
			wantOK: false,
			validate: func(t *testing.T, levels [][]string) {
				require.Nil(t, levels)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			levels, ok := Sort(tt.g)
			require.Equal(t, tt.wantOK, ok, "expected ok: %v, got: %v", tt.wantOK, ok)
			if tt.validate != nil {
				tt.validate(t, levels)
			}
		})
	}
}
