package graph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_topSortDag(t *testing.T) {
	tests := []struct {
		name    string
		dag     dag
		want    [][]NodeID
		wantErr bool
	}{
		{
			name: "simple linear graph",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{"B"}},
				"B": dagNode{id: "B", next: []NodeID{"C"}},
				"C": dagNode{id: "C", next: []NodeID{}},
			},
			want: [][]NodeID{
				{"A"},
				{"B"},
				{"C"},
			},
			wantErr: false,
		},
		{
			name: "graph with parallel branches",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{"B", "C"}},
				"B": dagNode{id: "B", next: []NodeID{"D"}},
				"C": dagNode{id: "C", next: []NodeID{"D"}},
				"D": dagNode{id: "D", next: []NodeID{}},
			},
			want: [][]NodeID{
				{"A"},
				{"B", "C"},
				{"D"},
			},
			wantErr: false,
		},
		{
			name: "graph with multiple start nodes",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{"C"}},
				"B": dagNode{id: "B", next: []NodeID{"C"}},
				"C": dagNode{id: "C", next: []NodeID{}},
			},
			want: [][]NodeID{
				{"A", "B"},
				{"C"},
			},
			wantErr: false,
		},
		{
			name: "complex graph with multiple parallelism levels",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{"B", "C"}},
				"B": dagNode{id: "B", next: []NodeID{"D", "E"}},
				"C": dagNode{id: "C", next: []NodeID{"E", "F"}},
				"D": dagNode{id: "D", next: []NodeID{"G"}},
				"E": dagNode{id: "E", next: []NodeID{"G"}},
				"F": dagNode{id: "F", next: []NodeID{"G"}},
				"G": dagNode{id: "G", next: []NodeID{}},
			},
			want: [][]NodeID{
				{"A"},
				{"B", "C"},
				{"D", "E", "F"},
				{"G"},
			},
			wantErr: false,
		},
		{
			name: "graph with cycle",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{"B"}},
				"B": dagNode{id: "B", next: []NodeID{"C"}},
				"C": dagNode{id: "C", next: []NodeID{"A"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "graph with self-cycle",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{"A"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "single node",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{}},
			},
			want: [][]NodeID{
				{"A"},
			},
			wantErr: false,
		},
		{
			name:    "empty graph",
			dag:     dag{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "graph with InputNodeID",
			dag: dag{
				InputNodeID: dagNode{id: InputNodeID, next: []NodeID{"A", "B"}},
				"A":         dagNode{id: "A", next: []NodeID{"C"}},
				"B":         dagNode{id: "B", next: []NodeID{"C"}},
				"C":         dagNode{id: "C", next: []NodeID{}},
			},
			want: [][]NodeID{
				{InputNodeID},
				{"A", "B"},
				{"C"},
			},
			wantErr: false,
		},
		{
			name: "diamond graph",
			dag: dag{
				"A": dagNode{id: "A", next: []NodeID{"B", "C"}},
				"B": dagNode{id: "B", next: []NodeID{"D"}},
				"C": dagNode{id: "C", next: []NodeID{"D"}},
				"D": dagNode{id: "D", next: []NodeID{"E"}},
				"E": dagNode{id: "E", next: []NodeID{}},
			},
			want: [][]NodeID{
				{"A"},
				{"B", "C"},
				{"D"},
				{"E"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := topSortDag(tt.dag)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.want), len(got), "number of levels should match")
			for i := range tt.want {
				require.ElementsMatch(t, tt.want[i], got[i])
			}
		})
	}
}
