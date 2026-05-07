package failure

import (
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"github.com/stretchr/testify/assert"
)

func TestFailure_VisitedNodes(t *testing.T) {
	f := Failure{
		NodeTraces: []trace.NodeTrace{
			{NodeID: "a", Error: ""},
			{NodeID: "b", Error: "boom"},
			{NodeID: "c", Error: ""},
		},
	}
	assert.Equal(t, []string{"a", "c"}, f.VisitedNodes())
}

func TestFailure_VisitedNodes_Empty(t *testing.T) {
	f := Failure{}
	assert.Empty(t, f.VisitedNodes())
}

func TestFailure_VisitedNodes_AllFailed(t *testing.T) {
	f := Failure{
		NodeTraces: []trace.NodeTrace{
			{NodeID: "a", Error: "boom"},
			{NodeID: "b", Error: "boom"},
		},
	}
	assert.Empty(t, f.VisitedNodes())
}
