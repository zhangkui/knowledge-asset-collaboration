package regression

import (
    "context"
    "sync"
    "testing"
    "time"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/task"
)

func TestBug12_ConcurrentWorkersClaimDueJobOnce(t *testing.T) {
    ctx := context.Background()
    scheduler := task.NewScheduler()
    if _, err := scheduler.Register(ctx, task.Schedule{ID: "schedule-12", Kind: "review-reminder", Interval: time.Nanosecond}); err != nil {
        t.Fatalf("register schedule: %v", err)
    }
    queue := &task.Queue{}
    count, err := scheduler.EnqueueDue(ctx, queue, time.Now().Add(time.Second))
    if err != nil {
        t.Fatalf("enqueue due schedules: %v", err)
    }
    if count != 1 {
        t.Fatalf("expected one due schedule, got %d", count)
    }
    start := make(chan struct{})
    results := make(chan bool, 2)
    jobs := make(chan task.Job, 2)
    var group sync.WaitGroup
    for i := 0; i < 2; i++ {
        group.Add(1)
        go func() {
            defer group.Done()
            <-start
            job, ok := queue.Next(ctx)
            results <- ok
            if ok {
                jobs <- job
            }
        }()
    }
    close(start)
    group.Wait()
    close(results)
    close(jobs)
    successes := 0
    for ok := range results {
        if ok {
            successes++
        }
    }
    if successes != 1 {
        t.Fatalf("concurrent workers must claim one job exactly once, successes=%d", successes)
    }
    for job := range jobs {
        if job.ID != "schedule-12" {
            t.Fatalf("claimed unexpected job: %#v", job)
        }
    }
}
