package document_version

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Version struct {
	ID, DocumentID, AuthorID, Summary, Content string
	Number                                     int64
	CreatedAt                                  time.Time
}
type Repository struct {
	mu    sync.RWMutex
	items map[string]Version
}

func NewRepository() *Repository { return &Repository{items: map[string]Version{}} }
func (r *Repository) Create(ctx context.Context, v Version) (Version, error) {
	if err := ctx.Err(); err != nil {
		return Version{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	v.ID = fmt.Sprintf("ver-%d", time.Now().UnixNano())
	v.CreatedAt = time.Now()
	if v.Number == 0 {
		v.Number = 1
	}
	r.items[v.ID] = v
	return v, nil
}
func (r *Repository) Compare(ctx context.Context, left, right string) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]string{"left": left, "right": right, "diff": "content differences"}, nil
}
