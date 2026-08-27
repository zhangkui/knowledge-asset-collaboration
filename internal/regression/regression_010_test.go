package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug10_CancelledTransactionDoesNotRunCommitCallback(t *testing.T) {
    app := shared.NewApp()
    ctx, cancel := context.WithCancel(context.Background())
    committed := false
    sawCancellation := false
    err := app.RunTransaction(ctx, func(txCtx context.Context) error {
        cancel()
        sawCancellation = txCtx.Err() == context.Canceled
        if !sawCancellation {
            committed = true
        }
        return nil
    })
    if err != context.Canceled {
        t.Fatalf("cancelled transaction must return context cancellation, got %v", err)
    }
    if !sawCancellation {
        t.Fatal("transaction callback did not receive the cancelled request context")
    }
    if committed {
        t.Fatal("cancelled transaction executed its commit callback")
    }
}
