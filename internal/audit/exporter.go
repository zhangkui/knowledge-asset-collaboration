package audit

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

type Filter struct {
	ActorID, Action, ObjectType, ObjectID, Keyword string
	From, To                                       time.Time
	Page, PageSize                                 int
}
type Page struct {
	Items                 []Event
	Total, Page, PageSize int
}
type Exporter struct {
	mu     sync.RWMutex
	events []Event
}

func NewExporter() *Exporter { return &Exporter{events: []Event{}} }
func (e *Exporter) Append(ctx context.Context, event Event) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	event.At = time.Now()
	e.events = append(e.events, event)
	return nil
}
func (e *Exporter) Query(ctx context.Context, f Filter) (Page, error) {
	if ctx == nil {
		return Page{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Page{}, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 500 {
		f.PageSize = 50
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	matches := make([]Event, 0)
	for _, event := range e.events {
		if f.ActorID != "" && event.ActorID != f.ActorID {
			continue
		}
		if f.Action != "" && event.Action != f.Action {
			continue
		}
		if f.ObjectType != "" && event.ObjectType != f.ObjectType {
			continue
		}
		if f.ObjectID != "" && event.ObjectID != f.ObjectID {
			continue
		}
		if !f.From.IsZero() && event.At.Before(f.From) {
			continue
		}
		if !f.To.IsZero() && event.At.After(f.To) {
			continue
		}
		if f.Keyword != "" && !strings.Contains(strings.ToLower(event.Action+" "+event.ObjectID), strings.ToLower(f.Keyword)) {
			continue
		}
		matches = append(matches, event)
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].At.After(matches[j].At) })
	total := len(matches)
	start := (f.Page - 1) * f.PageSize
	if start > total {
		start = total
	}
	end := start + f.PageSize
	if end > total {
		end = total
	}
	return Page{Items: append([]Event(nil), matches[start:end]...), Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}
func (e *Exporter) ExportCSV(ctx context.Context, w io.Writer, f Filter) error {
	page, err := e.Query(ctx, f)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"actor_id", "action", "object_type", "object_id", "at"}); err != nil {
		return err
	}
	for _, event := range page.Items {
		if err := writer.Write([]string{event.ActorID, event.Action, event.ObjectType, event.ObjectID, event.At.Format(time.RFC3339)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
func (e *Exporter) CountByAction(ctx context.Context, from, to time.Time) (map[string]int, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := map[string]int{}
	for _, event := range e.events {
		if !from.IsZero() && event.At.Before(from) {
			continue
		}
		if !to.IsZero() && event.At.After(to) {
			continue
		}
		out[event.Action]++
	}
	return out, nil
}
func (e *Exporter) Purge(ctx context.Context, before time.Time) int {
	if ctx == nil || ctx.Err() != nil {
		return 0
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	kept := e.events[:0]
	removed := 0
	for _, event := range e.events {
		if !before.IsZero() && event.At.Before(before) {
			removed++
		} else {
			kept = append(kept, event)
		}
	}
	e.events = kept
	return removed
}
