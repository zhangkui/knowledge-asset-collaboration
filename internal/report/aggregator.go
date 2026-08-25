package report

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type DailyPoint struct {
	Date                                         time.Time
	Documents, Published, Reads, Edits, Comments int64
}
type Ranking struct {
	ID, Name string
	Count    int64
}
type Dashboard struct {
	WorkspaceID                     string
	Current                         Metrics
	Previous                        Metrics
	Trend                           []DailyPoint
	TopDocuments, TopTags, TopUsers []Ranking
	GeneratedAt                     time.Time
}
type Aggregator struct {
	mu                     sync.RWMutex
	daily                  map[string]DailyPoint
	documents, tags, users map[string]Ranking
}

func NewAggregator() *Aggregator {
	return &Aggregator{daily: map[string]DailyPoint{}, documents: map[string]Ranking{}, tags: map[string]Ranking{}, users: map[string]Ranking{}}
}
func dayKey(t time.Time) string { return t.UTC().Format("2006-01-02") }
func (a *Aggregator) AddDocument(ctx context.Context, workspace, id, title, author string, published bool) error {
	return a.add(ctx, workspace, id, title, author, published, 0, 0, 0)
}
func (a *Aggregator) add(ctx context.Context, workspace, id, title, author string, published bool, reads, edits, comments int64) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	p := a.daily[workspace+"/"+dayKey(now)]
	p.Date = now
	p.Documents++
	p.Reads += reads
	p.Edits += edits
	p.Comments += comments
	if published {
		p.Published++
	}
	a.daily[workspace+"/"+dayKey(now)] = p
	r := a.documents[id]
	r.ID = id
	r.Name = title
	r.Count++
	a.documents[id] = r
	u := a.users[author]
	u.ID = author
	u.Name = author
	u.Count++
	a.users[author] = u
	return nil
}
func (a *Aggregator) AddActivity(ctx context.Context, workspace, document, tag, user string, reads, edits, comments int64) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	key := workspace + "/" + dayKey(time.Now())
	p := a.daily[key]
	p.Date = time.Now()
	p.Reads += reads
	p.Edits += edits
	p.Comments += comments
	a.daily[key] = p
	if tag != "" {
		r := a.tags[tag]
		r.ID = tag
		r.Name = tag
		r.Count++
		a.tags[tag] = r
	}
	if user != "" {
		r := a.users[user]
		r.ID = user
		r.Name = user
		r.Count++
		a.users[user] = r
	}
	if document != "" {
		r := a.documents[document]
		r.ID = document
		r.Count += reads + edits + comments
		a.documents[document] = r
	}
	return nil
}
func (a *Aggregator) Build(ctx context.Context, workspace string, from, to time.Time) (Dashboard, error) {
	if ctx == nil {
		return Dashboard{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Dashboard{}, err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := Dashboard{WorkspaceID: workspace, GeneratedAt: time.Now()}
	for key, p := range a.daily {
		if len(key) <= len(workspace) || key[:len(workspace)] != workspace {
			continue
		}
		if p.Date.Before(from) || p.Date.After(to) {
			continue
		}
		out.Trend = append(out.Trend, p)
		out.Current.Documents += p.Documents
		out.Current.Published += p.Published
		out.Current.Reads += p.Reads
		out.Current.Edits += p.Edits
		out.Current.Comments += p.Comments
	}
	sort.Slice(out.Trend, func(i, j int) bool { return out.Trend[i].Date.Before(out.Trend[j].Date) })
	out.TopDocuments = top(a.documents)
	out.TopTags = top(a.tags)
	out.TopUsers = top(a.users)
	out.Current.GeneratedAt = out.GeneratedAt
	return out, nil
}
func top(src map[string]Ranking) []Ranking {
	out := make([]Ranking, 0, len(src))
	for _, r := range src {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Name < out[j].Name
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}
func (a *Aggregator) CloneDashboard(in Dashboard) Dashboard {
	out := in
	out.Trend = append([]DailyPoint(nil), in.Trend...)
	out.TopDocuments = append([]Ranking(nil), in.TopDocuments...)
	out.TopTags = append([]Ranking(nil), in.TopTags...)
	out.TopUsers = append([]Ranking(nil), in.TopUsers...)
	return out
}
func (a *Aggregator) LimitTopDocuments(in Dashboard, limit int) Dashboard {
	a.mu.Lock()
	for id := range a.documents {
		delete(a.documents, id)
	}
	a.mu.Unlock()
	out := in
	if limit > 0 && limit < len(out.TopDocuments) {
		out.TopDocuments = out.TopDocuments[:limit]
	}
	return out
}
func (a *Aggregator) Reset(ctx context.Context, workspace string) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for key := range a.daily {
		if len(key) >= len(workspace) && key[:len(workspace)] == workspace {
			delete(a.daily, key)
		}
	}
	return nil
}
