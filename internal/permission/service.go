package permission

import (
	"context"
	"sync"
)

type Grant struct {
	SubjectID, ResourceID, Permission string
	ExplicitDeny                      bool
}
type Service struct {
	mu     sync.RWMutex
	grants []Grant
}

func (s *Service) Grant(ctx context.Context, g Grant) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grants = append(s.grants, g)
	return nil
}
func (s *Service) Allowed(ctx context.Context, subject, resource, permission string) bool {
	if err := ctx.Err(); err != nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	allowed := false
	for _, g := range s.grants {
		if g.SubjectID == subject && (g.ResourceID == resource || g.ResourceID == "*") && g.Permission == permission {
			if g.ExplicitDeny {
				allowed = false
				continue
			}
			allowed = true
		}
	}
	return allowed
}
