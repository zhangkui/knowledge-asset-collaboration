package regression

import (
    "context"
    "testing"
    "time"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/task"
)

func TestBug25_CancelledScheduleDoesNotEnqueueJobs(t *testing.T) {
    scheduler := task.NewScheduler()
    queue := &task.Queue{}
    ctx, cancel := context.WithCancel(context.Background())
    if _, err := scheduler.Register(context.Background(), task.Schedule{ID: "schedule-25", Kind: "publish", Interval: time.Second}); err != nil { t.Fatalf("register schedule: %v", err) }
    cancel()
    if _, err := scheduler.EnqueueDue(ctx, queue, time.Now().Add(time.Second)); err == nil { t.Fatal("cancelled scheduling must return context cancellation") }
    if _, ok := queue.Next(context.Background()); ok { t.Fatal("cancelled scheduling enqueued a job") }
}
