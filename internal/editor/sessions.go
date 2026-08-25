package editor

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Lock struct {
	DocumentID, UserID    string
	AcquiredAt, ExpiresAt time.Time
}
type Change struct {
	ID, DocumentID, UserID, Operation string
	BaseVersion, ResultVersion        int64
	At                                time.Time
}
type Session struct {
	ID, DocumentID, UserID string
	ConnectedAt, LastSeen  time.Time
	Active                 bool
}
type SessionManager struct {
	mu       sync.RWMutex
	locks    map[string]Lock
	sessions map[string]Session
	changes  map[string][]Change
}

func NewSessionManager() *SessionManager {
	return &SessionManager{locks: map[string]Lock{}, sessions: map[string]Session{}, changes: map[string][]Change{}}
}
func (s *SessionManager) Acquire(ctx context.Context, doc, user string, ttl time.Duration) (Lock, error) {
	if ctx == nil {
		return Lock{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Lock{}, err
	}
	if ttl <= 0 {
		return Lock{}, errors.New("lock ttl required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if l, ok := s.locks[doc]; ok && now.Before(l.ExpiresAt) && l.UserID != user {
		return Lock{}, errors.New("document locked")
	}
	l := Lock{DocumentID: doc, UserID: user, AcquiredAt: now, ExpiresAt: now.Add(ttl)}
	s.locks[doc] = l
	return l, nil
}
func (s *SessionManager) Release(ctx context.Context, doc, user string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.locks[doc]
	if !ok {
		return nil
	}
	if l.UserID != user {
		return errors.New("lock owner required")
	}
	delete(s.locks, doc)
	return nil
}
func (s *SessionManager) Connect(ctx context.Context, id, doc, user string) (Session, error) {
	if ctx == nil {
		return Session{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	session := Session{ID: id, DocumentID: doc, UserID: user, ConnectedAt: now, LastSeen: now, Active: true}
	s.sessions[id] = session
	return session, nil
}
func (s *SessionManager) Heartbeat(ctx context.Context, id string) (Session, error) {
	if ctx == nil {
		return Session{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return Session{}, errors.New("session not found")
	}
	if !session.Active {
		return Session{}, errors.New("session disconnected")
	}
	session.LastSeen = time.Now()
	session.Active = true
	s.sessions[id] = session
	return session, nil
}
func (s *SessionManager) Disconnect(ctx context.Context, id string) error {
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
	session.Active = false
	s.sessions[id] = session
	return nil
}
func (s *SessionManager) DisconnectFromHub(ctx context.Context, hub *Hub, id string) error {
	if hub == nil {
		return errors.New("editor hub is required")
	}
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	session, ok := s.sessions[id]
	if !ok {
		s.mu.Unlock()
		return errors.New("session not found")
	}
	session.Active = false
	s.sessions[id] = session
	s.mu.Unlock()
	hub.LeaveSession(session)
	return nil
}

func (s *SessionManager) Presence(ctx context.Context, doc string) ([]Session, error) {
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
		if session.DocumentID == doc && session.Active {
			out = append(out, session)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ConnectedAt.Before(out[j].ConnectedAt) })
	return out, nil
}
func (s *SessionManager) Record(ctx context.Context, c Change) (Change, error) {
	if ctx == nil {
		return Change{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Change{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = "change-" + time.Now().Format("20060102150405.000000000")
	c.At = time.Now()
	s.changes[c.DocumentID] = append(s.changes[c.DocumentID], c)
	return c, nil
}
func (s *SessionManager) Changes(ctx context.Context, doc string, after int64) ([]Change, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Change{}
	for _, c := range s.changes[doc] {
		if c.ResultVersion > after {
			out = append(out, c)
		}
	}
	return out, nil
}
