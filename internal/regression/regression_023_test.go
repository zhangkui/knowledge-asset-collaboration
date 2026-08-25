package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/publish"
)

func TestBug23_CancelledPublishDoesNotContinueAcrossLayers(t *testing.T) {
    cancelled, cancel := context.WithCancel(context.Background())
    cancel()
    if _, err := (publish.Service{}).Publish(cancelled, "approved"); err == nil {
        t.Fatal("cancelled publish must return context cancellation")
    }
}
