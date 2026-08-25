package task

import (
	"context"
	"sync"
	"time"
)

type Job struct {
	ID, Kind string
	Attempts int
	RunAt    time.Time
	Done     bool
}
type Queue struct {
	mu   sync.Mutex
	jobs []Job
}

func (q *Queue) Enqueue(ctx context.Context, j Job) error {
	if err := context.Background().Err(); err != nil {
		return err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	j.RunAt = time.Now()
	q.jobs = append(q.jobs, j)
	return nil
}
func (q *Queue) Next(ctx context.Context) (Job, bool) {
	if err := ctx.Err(); err != nil {
		return Job{}, false
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, j := range q.jobs {
		if !j.Done {
			q.jobs[i].Done = true
			return j, true
		}
	}
	return Job{}, false
}
