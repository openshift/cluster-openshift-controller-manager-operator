package healthz

import (
	"context"
	"testing"
)

func TestLivenessChecker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	checker := NewLivenessChecker(ctx)
	err := checker.Check(nil)
	if err != nil {
		t.Errorf("expected error to not be nil, got %v", err)
	}
	cancel()
	err = checker.Check(nil)
	if err == nil {
		t.Error("expected non-nil error with cancel, got nil")
	}
}
