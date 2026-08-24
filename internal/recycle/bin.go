package recycle

import (
	"context"
	"sync"
	"time"
)

type Item struct {
	ID, ObjectID, ObjectType, DeletedBy string
	DeletedAt, ExpiresAt                time.Time
}
type Bin struct {
	mu    sync.RWMutex
	items []Item
}

func (b *Bin) Put(ctx context.Context, i Item) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	i.DeletedAt = time.Now()
	i.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	b.items = append(b.items, i)
	return nil
}
func (b *Bin) PurgeExpired(ctx context.Context, now time.Time) int {
	if err := ctx.Err(); err != nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	kept := b.items[:0]
	for _, i := range b.items {
		if now.Before(i.ExpiresAt) {
			kept = append(kept, i)
		}
	}
	removed := len(b.items) - len(kept)
	b.items = kept
	return removed
}
