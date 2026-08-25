package regression

import (
    "context"
    "testing"
    "time"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug20_CancelledRefreshDoesNotTouchSession(t *testing.T) {
    app := shared.NewApp()
    ctx := context.Background()
    session, err := app.SessionStore.Create(ctx, "user-20", "refresh-20", "127.0.0.1", "test", time.Hour)
    if err != nil {
        t.Fatalf("create session: %v", err)
    }
    before, err := app.SessionStore.Lookup(ctx, session.ID)
    if err != nil {
        t.Fatalf("lookup session: %v", err)
    }
    cancelled, cancel := context.WithCancel(ctx)
    cancel()
    if _, err := app.RefreshSession(cancelled, session.ID); err == nil {
        t.Fatal("cancelled refresh must return context cancellation")
    }
    after, err := app.SessionStore.Lookup(ctx, session.ID)
    if err != nil {
        t.Fatalf("lookup session after cancelled refresh: %v", err)
    }
    if !after.LastSeen.Equal(before.LastSeen) {
        t.Fatalf("cancelled refresh extended session activity: before=%v after=%v", before.LastSeen, after.LastSeen)
    }
}
