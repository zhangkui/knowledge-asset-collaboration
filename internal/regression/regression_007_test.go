package regression

import (
    "context"
    "fmt"
    "sync"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/notification"
    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug07_ConcurrentNotificationPublishingKeepsEveryNotice(t *testing.T) {
    app := shared.NewApp()
    const total = 64
    start := make(chan struct{})
    errorsCh := make(chan error, total)
    var group sync.WaitGroup
    for index := 0; index < total; index++ {
        group.Add(1)
        go func(index int) {
            defer group.Done()
            <-start
            errorsCh <- app.PublishNotification(context.Background(), notification.Notification{
                ID: fmt.Sprintf("notice-%02d", index),
                UserID: "reader-7",
                Kind: "document.published",
                Message: fmt.Sprintf("document %02d published", index),
            })
        }(index)
    }
    close(start)
    group.Wait()
    close(errorsCh)
    for err := range errorsCh {
        if err != nil {
            t.Fatalf("publish notification: %v", err)
        }
    }
    notices := app.NotificationCenter.Unread("reader-7")
    if len(notices) != total {
        t.Fatalf("concurrent notification publishing lost notices: got %d want %d", len(notices), total)
    }
    seen := make(map[string]bool, total)
    for _, notice := range notices {
        if seen[notice.ID] {
            t.Fatalf("concurrent notification publishing duplicated notice %q", notice.ID)
        }
        seen[notice.ID] = true
    }
}
