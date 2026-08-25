package notification

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Preference struct {
	UserID                    string
	Kinds                     map[string]bool
	EmailEnabled, PushEnabled bool
	UpdatedAt                 time.Time
}
type Event struct {
	Kind, SubjectID, ObjectID, Message string
	Recipients                         []string
	CreatedAt                          time.Time
}
type Inbox struct {
	mu          sync.RWMutex
	items       map[string][]Notification
	preferences map[string]Preference
	events      []Event
}

func NewInbox() *Inbox {
	return &Inbox{items: map[string][]Notification{}, preferences: map[string]Preference{}}
}
func (i *Inbox) SetPreference(ctx context.Context, p Preference) (Preference, error) {
	if ctx == nil {
		return Preference{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Preference{}, err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	p.UpdatedAt = time.Now()
	if p.Kinds == nil {
		p.Kinds = map[string]bool{}
	}
	i.preferences[p.UserID] = p
	return p, nil
}
func (i *Inbox) Publish(ctx context.Context, e Event) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	e.CreatedAt = time.Now()
	i.events = append(i.events, e)
	count := 0
	for _, user := range e.Recipients {
		p := i.preferences[user]
		if allowed, configured := p.Kinds[e.Kind]; configured && !allowed {
			continue
		}
		n := Notification{ID: "notice-" + time.Now().Format("20060102150405.000000000"), UserID: user, Kind: e.Kind, Message: e.Message, CreatedAt: e.CreatedAt}
		i.items[user] = append(i.items[user], n)
		count++
	}
	return count, nil
}
func (i *Inbox) List(ctx context.Context, user string, unreadOnly bool) ([]Notification, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := []Notification{}
	for _, n := range i.items[user] {
		if !unreadOnly || !n.Read {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].CreatedAt.After(out[b].CreatedAt) })
	return out, nil
}
func (i *Inbox) Mark(ctx context.Context, user, notificationID string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for idx, n := range i.items[user] {
		if n.ID == notificationID {
			i.items[user][idx].Read = true
			return nil
		}
	}
	return errors.New("notification not found")
}
func (i *Inbox) MarkAll(ctx context.Context, user string) int {
	if ctx == nil || ctx.Err() != nil {
		return 0
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	count := 0
	for idx := range i.items[user] {
		if !i.items[user][idx].Read {
			i.items[user][idx].Read = true
			count++
		}
	}
	return count
}
func (i *Inbox) UnreadCount(ctx context.Context, user string) (int, error) {
	items, err := i.List(ctx, user, true)
	return len(items), err
}
func (i *Inbox) Events(ctx context.Context, kind string) ([]Event, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := []Event{}
	for _, e := range i.events {
		if kind == "" || e.Kind == kind {
			e.Recipients = append([]string(nil), e.Recipients...)
			out = append(out, e)
		}
	}
	return out, nil
}
