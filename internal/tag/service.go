package tag

import (
	"context"
	"fmt"
	"sync"
)

type Tag struct{ ID, Name, Color string }
type Service struct {
	mu   sync.RWMutex
	tags map[string]Tag
}

func NewService() *Service { return &Service{tags: map[string]Tag{}} }
func (s *Service) Create(ctx context.Context, name, color string) (Tag, error) {
	if err := ctx.Err(); err != nil {
		return Tag{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Tag{ID: fmt.Sprintf("tag-%d", len(s.tags)+1), Name: name, Color: color}
	s.tags[t.ID] = t
	return t, nil
}
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tags, id)
	return nil
}
