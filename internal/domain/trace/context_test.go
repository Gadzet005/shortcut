package trace

import (
	"context"
	"testing"
)

func TestContextCollector(t *testing.T) {
	c := NewCollector("ctx-test")
	ctx := WithCollector(context.Background(), c)

	got, ok := GetCollector(ctx)
	if !ok {
		t.Fatal("expected collector in context")
	}
	if got.RequestID() != "ctx-test" {
		t.Errorf("expected request ID 'ctx-test', got '%s'", got.RequestID())
	}
}

func TestContextCollector_Missing(t *testing.T) {
	ctx := context.Background()
	_, ok := GetCollector(ctx)
	if ok {
		t.Error("expected no collector in empty context")
	}
}
