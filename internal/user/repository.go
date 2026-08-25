package user

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type User struct {
	ID, Email, Name, DepartmentID string
	Enabled                       bool
	CreatedAt                     time.Time
}
type Repository struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewRepository() *Repository { return &Repository{users: map[string]User{}} }
func (r *Repository) Create(ctx context.Context, u User) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if u.ID == "" {
		u.ID = fmt.Sprintf("usr-%d", time.Now().UnixNano())
	}
	u.CreatedAt = time.Now()
	u.Enabled = true
	r.users[u.ID] = u
	return u, nil
}
func (r *Repository) Get(ctx context.Context, id string) (User, bool) {
	if err := ctx.Err(); err != nil {
		return User{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	return u, ok
}
func (r *Repository) Disable(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return fmt.Errorf("user %s not found", id)
	}
	u.Enabled = false
	r.users[id] = u
	return nil
}
