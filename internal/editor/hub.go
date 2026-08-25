package editor

import (
	"context"
	"sync"
	"time"
)

type Presence struct {
	UserID, DocumentID string
	JoinedAt           time.Time
}
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]Presence
}

func NewHub() *Hub { return &Hub{rooms: map[string]map[string]Presence{}} }
func (h *Hub) Join(ctx context.Context, documentID string, p Presence) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[documentID] == nil {
		h.rooms[documentID] = map[string]Presence{}
	}
	p.JoinedAt = time.Now()
	h.rooms[documentID][p.UserID] = p
	return nil
}
func (h *Hub) Leave(documentID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[documentID]; room != nil {
		delete(room, userID)
		if len(room) == 0 {
			delete(h.rooms, documentID)
		}
	}
}
func (h *Hub) LeaveSession(session Session) {
	h.Leave(session.DocumentID, session.UserID)
}

func (h *Hub) Online(documentID string) []Presence {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := []Presence{}
	for _, p := range h.rooms[documentID] {
		out = append(out, p)
	}
	return out
}
