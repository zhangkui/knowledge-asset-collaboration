package audit

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	ActorID, Action, ObjectType, ObjectID string
	At                                    time.Time
}
type Logger struct {
	mu     sync.RWMutex
	events []Event
}

func (l *Logger) Record(ctx context.Context, e Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e.At = time.Now()
	l.events = append(l.events, e)
	return nil
}
func (l *Logger) List(ctx context.Context) []Event {
	if err := ctx.Err(); err != nil {
		return nil
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}
