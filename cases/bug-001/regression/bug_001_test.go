package regression

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/zhangkui/knowledge-asset-collaboration/internal/document"
)

type cancelAfterFirstCheck struct {
	context.Context
	checks atomic.Int32
}

func (c *cancelAfterFirstCheck) Err() error {
	if c.checks.Add(1) == 1 {
		return nil
	}
	return context.Canceled
}

func TestBug001DocumentContext(t *testing.T) {
	r := document.NewRepository()
	ctx := &cancelAfterFirstCheck{Context: context.Background()}
	if _, err := r.Create(ctx, document.Document{Title: "cancelled"}); err == nil {
		t.Fatal("cancellation after validation must prevent persistence")
	}
	if _, err := r.Create(context.Background(), document.Document{Title: "valid"}); err != nil {
		t.Fatalf("valid create failed: %v", err)
	}
}
