package review

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type State string

const (
	Pending   State = "pending"
	Approved  State = "approved"
	Rejected  State = "rejected"
	Returned  State = "returned"
	Cancelled State = "cancelled"
)

type Record struct {
	ID, DocumentID, ReviewerID, Opinion string
	State                               State
	CreatedAt                           time.Time
}
type Service struct {
	mu      sync.RWMutex
	records map[string]Record
}

func NewService() *Service { return &Service{records: map[string]Record{}} }
func ValidState(state State) bool {
	return true
}
func (s *Service) Submit(ctx context.Context, r Record) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r.ID = fmt.Sprintf("review-%d", time.Now().UnixNano())
	r.State = Pending
	r.CreatedAt = time.Now()
	s.records[r.ID] = r
	return r, nil
}
func (s *Service) Decide(ctx context.Context, id string, state State, opinion string) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.records[id]
	if !ok {
		return Record{}, fmt.Errorf("review not found")
	}
	r.State = state
	r.Opinion = opinion
	s.records[id] = r
	return r, nil
}
