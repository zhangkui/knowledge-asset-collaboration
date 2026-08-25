package postgres

import (
	"context"
	"errors"
	"sync"
)

type Executor interface {
	ExecContext(context.Context, string, ...any) (Result, error)
	QueryContext(context.Context, string, ...any) (Rows, error)
}
type Result interface{ RowsAffected() (int64, error) }
type Rows interface {
	Next() bool
	Scan(...any) error
	Close() error
}
type Store struct {
	mu       sync.RWMutex
	executor Executor
	ready    bool
}

func New(executor Executor) *Store { return &Store{executor: executor} }
func (s *Store) Ping(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	s.ready = true
	s.mu.Unlock()
	return nil
}
func (s *Store) Ready() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.ready }
func (s *Store) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if fn == nil {
		return errors.New("transaction callback is required")
	}
	return fn(ctx)
}
