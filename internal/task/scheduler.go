package task

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Schedule struct {
	ID, Kind    string
	Interval    time.Duration
	NextRun     time.Time
	Enabled     bool
	MaxAttempts int
}
type Run struct {
	JobID                 string
	StartedAt, FinishedAt *time.Time
	Attempts              int
	Success               bool
	Error                 string
}
type Scheduler struct {
	mu        sync.RWMutex
	schedules map[string]Schedule
	runs      map[string][]Run
}

func NewScheduler() *Scheduler {
	return &Scheduler{schedules: map[string]Schedule{}, runs: map[string][]Run{}}
}
func (s *Scheduler) Register(ctx context.Context, job Schedule) (Schedule, error) {
	if ctx == nil {
		return Schedule{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Schedule{}, err
	}
	if job.ID == "" || job.Kind == "" || job.Interval <= 0 {
		return Schedule{}, errors.New("invalid schedule")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if job.MaxAttempts <= 0 {
		job.MaxAttempts = 3
	}
	job.NextRun = time.Now().Add(job.Interval)
	job.Enabled = true
	s.schedules[job.ID] = job
	return job, nil
}
func (s *Scheduler) Pause(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.schedules[id]
	if !ok {
		return errors.New("schedule not found")
	}
	job.Enabled = false
	s.schedules[id] = job
	return nil
}
func (s *Scheduler) Resume(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.schedules[id]
	if !ok {
		return errors.New("schedule not found")
	}
	job.Enabled = true
	if job.NextRun.Before(time.Now()) {
		job.NextRun = time.Now()
	}
	s.schedules[id] = job
	return nil
}
func (s *Scheduler) Due(ctx context.Context, now time.Time) ([]Schedule, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Schedule{}
	for _, job := range s.schedules {
		if job.Enabled && !job.NextRun.After(now) {
			out = append(out, job)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].NextRun.Before(out[j].NextRun) })
	return out, nil
}
func (s *Scheduler) EnqueueDue(ctx context.Context, q *Queue, now time.Time) (int, error) {
	if q == nil {
		return 0, errors.New("task queue is required")
	}
	due, err := s.Due(ctx, now)
	if err != nil {
		return 0, err
	}
	for _, schedule := range due {
		if err := q.Enqueue(ctx, Job{ID: schedule.ID, Kind: schedule.Kind}); err != nil {
			return 0, err
		}
	}
	return len(due), nil
}
func (s *Scheduler) Complete(ctx context.Context, id string, success bool, runErr error) (Schedule, error) {
	if ctx == nil {
		return Schedule{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Schedule{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.schedules[id]
	if !ok {
		return Schedule{}, errors.New("schedule not found")
	}
	now := time.Now()
	attempts := 1
	if history := s.runs[id]; len(history) > 0 {
		attempts = history[len(history)-1].Attempts + 1
	}
	record := Run{JobID: id, StartedAt: &now, FinishedAt: &now, Attempts: attempts, Success: success}
	if runErr != nil {
		record.Error = runErr.Error()
	}
	s.runs[id] = append(s.runs[id], record)
	if success || attempts >= job.MaxAttempts {
		job.NextRun = now.Add(job.Interval)
	} else {
		job.NextRun = now.Add(time.Second * time.Duration(attempts))
	}
	s.schedules[id] = job
	return job, nil
}
func (s *Scheduler) History(ctx context.Context, id string) ([]Run, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Run(nil), s.runs[id]...), nil
}
func (s *Scheduler) Remove(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[id]; !ok {
		return errors.New("schedule not found")
	}
	delete(s.schedules, id)
	delete(s.runs, id)
	return nil
}
