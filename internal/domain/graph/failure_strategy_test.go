package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFailureStrategy(t *testing.T) {
	cases := []struct {
		input string
		want  FailureStrategy
		ok    bool
	}{
		{"ignore", IgnoreFailureStrategy, true},
		{"revert", RevertFailureStrategy, true},
		{"save", SaveFailureStrategy, true},
		{"finish", FinishFailureStrategy, true},
		{"custom", CustomFailureStrategy, true},
		{"absent", AbsentFailureStrategy, true},
		{"", AbsentFailureStrategy, false},
		{"unknown", AbsentFailureStrategy, false},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := ParseFailureStrategy(tc.input)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseStrategyAction(t *testing.T) {
	cases := []struct {
		input string
		want  StrategyAction
		ok    bool
	}{
		{"skip", SkipStrategyAction, true},
		{"retry", RetryStrategyAction, true},
		{"revert", RevertStrategyAction, true},
		{"finish", FinishStrategyAction, true},
		{"", "", false},
		{"unknown", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := ParseStrategyAction(tc.input)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}
