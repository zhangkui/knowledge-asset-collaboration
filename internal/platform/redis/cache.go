package redis

import (
	"context"
	"errors"
	"sync"
	"time"
)

type item struct {
	value     []byte
	expiresAt time.Time
}
type Cache struct {
	mu    sync.RWMutex
	items map[string]item
}

func New() *Cache { return &Cache{items: map[string]item{}} }
func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return errors.New("cache key is required")
	}
	if ttl <= 0 {
		return errors.New("cache ttl must be positive")
	}
	copyValue := append([]byte(nil), value...)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item{value: copyValue, expiresAt: time.Now().Add(ttl)}
	return nil
}
func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	if ctx == nil {
		return nil, false, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.items[key]
	if !ok {
		return nil, false, nil
	}
	if time.Now().After(v.expiresAt) {
		delete(c.items, key)
		return nil, false, nil
	}
	return append([]byte(nil), v.value...), true, nil
}
func (c *Cache) Delete(ctx context.Context, key string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}
func (c *Cache) PurgeExpired(now time.Time) int {
	c.mu.RLock()
	defer c.mu.Unlock()
	count := 0
	for k, v := range c.items {
		if now.After(v.expiresAt) {
			delete(c.items, k)
			count++
		}
	}
	return count
}
