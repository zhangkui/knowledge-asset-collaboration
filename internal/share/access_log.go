package share

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Access struct {
	ShareID, Token, VisitorID, IP, UserAgent string
	Allowed                                  bool
	Reason                                   string
	At                                       time.Time
}
type Registry struct {
	mu       sync.RWMutex
	links    map[string]Link
	accesses []Access
	revoked  map[string]time.Time
}

func NewRegistry() *Registry {
	return &Registry{links: map[string]Link{}, revoked: map[string]time.Time{}}
}
func (r *Registry) Add(ctx context.Context, link Link) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if link.Token == "" || link.DocumentID == "" {
		return errors.New("invalid share link")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.Token] = link
	return nil
}
func (r *Registry) Revoke(ctx context.Context, token string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.links[token]; !ok {
		return errors.New("share not found")
	}
	r.revoked[token] = time.Now()
	return nil
}
func (r *Registry) Lookup(ctx context.Context, token string) (*Link, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	link, ok := r.links[token]
	if !ok {
		return nil, nil
	}
	copy := link
	return &copy, nil
}
func (r *Registry) Open(ctx context.Context, token, visitor, ip, agent string) (Link, Access, error) {
	if ctx == nil {
		return Link{}, Access{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Link{}, Access{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	link, ok := r.links[token]
	access := Access{ShareID: link.ID, Token: token, VisitorID: visitor, IP: ip, UserAgent: agent, At: time.Now()}
	if !ok {
		access.Reason = "share not found"
		r.accesses = append(r.accesses, access)
		return Link{}, access, errors.New("share not found")
	}
	if _, revoked := r.revoked[token]; revoked {
		access.Reason = "share revoked"
		r.accesses = append(r.accesses, access)
		return Link{}, access, errors.New("share revoked")
	}
	if !Valid(link, time.Now()) {
		access.Reason = "share expired"
		r.accesses = append(r.accesses, access)
		return Link{}, access, errors.New("share expired")
	}
	access.Allowed = true
	r.accesses = append(r.accesses, access)
	return link, access, nil
}
func (r *Registry) AccessLog(ctx context.Context, token string) ([]Access, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Access{}
	for _, a := range r.accesses {
		if token == "" || a.Token == token {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}
func (r *Registry) Purge(ctx context.Context, now time.Time) int {
	if ctx == nil || ctx.Err() != nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	removed := 0
	for token, link := range r.links {
		if !now.Before(link.ExpiresAt) {
			delete(r.links, token)
			removed++
		}
	}
	return removed
}
