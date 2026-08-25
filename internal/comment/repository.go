package comment

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Comment struct {
	ID, DocumentID, AuthorID, Body, ParentID string
	Resolved                                 bool
	CreatedAt                                time.Time
}
type Repository struct {
	mu    sync.RWMutex
	items map[string]Comment
}

func NewRepository() *Repository { return &Repository{items: map[string]Comment{}} }
func (r *Repository) Add(ctx context.Context, c Comment) (Comment, error) {
	if err := ctx.Err(); err != nil {
		return Comment{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	c.ID = fmt.Sprintf("comment-%d", time.Now().UnixNano())
	c.CreatedAt = time.Now()
	r.items[c.ID] = c
	return c, nil
}
func (r *Repository) Resolve(ctx context.Context, id string, resolved bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.items[id]
	if !ok {
		return fmt.Errorf("comment not found")
	}
	c.Resolved = resolved
	r.items[id] = c
	return nil
}
