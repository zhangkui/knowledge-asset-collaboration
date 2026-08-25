package regression

import (
    "context"
    "sync"
    "testing"
    "time"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug22_ConcurrentCachePurgeDoesNotExposeExpiredValue(t *testing.T) {
    app := shared.NewApp()
    ctx := context.Background()
    if err := app.Cache.Set(ctx, "cache-22", []byte("stale"), time.Millisecond); err != nil {
        t.Fatalf("set cache: %v", err)
    }
    time.Sleep(2 * time.Millisecond)
    start := make(chan struct{})
    var group sync.WaitGroup
    reads := make(chan bool, 32)
    for i := 0; i < 32; i++ {
        group.Add(1)
        go func() {
            defer group.Done()
            <-start
            _, found, err := app.ReadCache(ctx, "cache-22")
            if err != nil {
                t.Errorf("read cache: %v", err)
                return
            }
            reads <- found
        }()
    }
    group.Add(1)
    go func() {
        defer group.Done()
        <-start
        app.PurgeCache(time.Now())
    }()
    close(start)
    group.Wait()
    close(reads)
    for found := range reads {
        if found {
            t.Fatal("concurrent cache purge exposed an expired value")
        }
    }
}
