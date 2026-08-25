package notification

import (
	"context"
	"sync"
	"time"
)

type Notification struct {
	ID, UserID, Kind, Message string
	Read                      bool
	CreatedAt                 time.Time
}
type Center struct {
	mu    sync.RWMutex
	items []Notification
}

func (c *Center) Push(ctx context.Context, n Notification) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	n.CreatedAt = time.Now()
	c.items = append(c.items, n)
	return nil
}
func (c *Center) MarkAllRead(ctx context.Context, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.items {
		if c.items[i].UserID == userID {
			c.items[i].Read = true
		}
	}
	return nil
}
func (c *Center) Unread(userID string) []Notification {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := []Notification{}
	for _, n := range c.items {
		if n.UserID == userID && !n.Read {
			out = append(out, n)
		}
	}
	return out
}
