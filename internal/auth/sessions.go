package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

type Session struct {
	ID, UserID, RefreshToken, IP, UserAgent string
	CreatedAt, LastSeen, ExpiresAt          time.Time
	Revoked                                 bool
}
type SessionStore struct {
	mu        sync.RWMutex
	sessions  map[string]Session
	byRefresh map[string]string
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: map[string]Session{}, byRefresh: map[string]string{}}
}
func (s *SessionStore) Create(ctx context.Context, user, refresh, ip, agent string, ttl time.Duration) (Session, error) {
	if ctx == nil {
		return Session{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(user) == "" || refresh == "" || ttl <= 0 {
		return Session{}, errors.New("invalid session")
	}
	now := time.Now()
	session := Session{ID: "session-" + now.Format("20060102150405.000000000"), UserID: user, RefreshToken: refresh, IP: ip, UserAgent: agent, CreatedAt: now, LastSeen: now, ExpiresAt: now.Add(ttl)}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	s.byRefresh[refresh] = session.ID
	return session, nil
}
func (s *SessionStore) Lookup(ctx context.Context, id string) (Session, error) {
	if ctx == nil {
		return Session{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Session{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return Session{}, errors.New("session not found")
	}
	if session.Revoked || !time.Now().Before(session.ExpiresAt) {
		return Session{}, errors.New("session expired")
	}
	return session, nil
}
func (s *SessionStore) Touch(ctx context.Context, id string) (Session, error) {
	if ctx == nil {
		return Session{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok || session.Revoked {
		return Session{}, errors.New("session not found")
	}
	if !time.Now().Before(session.ExpiresAt) {
		return Session{}, errors.New("session expired")
	}
	session.LastSeen = time.Now()
	s.sessions[id] = session
	return session, nil
}
func (s *SessionStore) Revoke(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return errors.New("session not found")
	}
	session.Revoked = true
	s.sessions[id] = session
	delete(s.byRefresh, session.RefreshToken)
	return nil
}
func (s *SessionStore) RevokeByRefresh(ctx context.Context, refresh string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byRefresh[refresh]
	if !ok {
		return errors.New("refresh token not found")
	}
	session := s.sessions[id]
	session.Revoked = true
	s.sessions[id] = session
	delete(s.byRefresh, refresh)
	return nil
}
func (s *SessionStore) List(ctx context.Context, user string) ([]Session, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Session{}
	for _, session := range s.sessions {
		if session.UserID == user {
			out = append(out, session)
		}
	}
	return out, nil
}
func (s *SessionStore) Purge(ctx context.Context, now time.Time) int {
	if ctx == nil || ctx.Err() != nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for id, session := range s.sessions {
		if session.Revoked || !now.Before(session.ExpiresAt) {
			delete(s.sessions, id)
			delete(s.byRefresh, session.RefreshToken)
			n++
		}
	}
	return n
}
