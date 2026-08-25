package attachment

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Attachment struct {
	ID, DocumentID, Name, ContentType string
	Size, Uploaded                    int64
	CreatedAt                         time.Time
}
type Service struct {
	mu    sync.RWMutex
	items map[string]Attachment
}

func NewService() *Service { return &Service{items: map[string]Attachment{}} }
func (s *Service) Start(ctx context.Context, a Attachment) (Attachment, error) {
	if err := ctx.Err(); err != nil {
		return Attachment{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	a.ID = fmt.Sprintf("att-%d", time.Now().UnixNano())
	a.CreatedAt = time.Now()
	s.items[a.ID] = a
	return a, nil
}
func (s *Service) Complete(ctx context.Context, id string) (Attachment, error) {
	if err := ctx.Err(); err != nil {
		return Attachment{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.items[id]
	if !ok {
		return Attachment{}, fmt.Errorf("attachment not found")
	}
	a.Uploaded = a.Size
	s.items[id] = a
	return a, nil
}
