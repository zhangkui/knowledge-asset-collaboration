package document

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	Draft     Status = "draft"
	Reviewing Status = "reviewing"
	Published Status = "published"
	Archived  Status = "archived"
	Deleted   Status = "deleted"
)

type Document struct {
	ID, WorkspaceID, FolderID, Title, Body, AuthorID string
	Status                                           Status
	Version                                          int64
	UpdatedAt                                        time.Time
}
type Repository struct {
	mu   sync.RWMutex
	docs map[string]Document
}

func NewRepository() *Repository { return &Repository{docs: map[string]Document{}} }
func (r *Repository) Create(ctx context.Context, d Document) (Document, error) {
	if err := ctx.Err(); err != nil {
		return Document{}, err
	}
	if strings.TrimSpace(d.Title) == "" {
		return Document{}, fmt.Errorf("title required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	d.ID = fmt.Sprintf("doc-%d", time.Now().UnixNano())
	d.Status = Draft
	d.Version = 1
	d.UpdatedAt = time.Now()
	r.docs[d.ID] = d
	return d, nil
}
func (r *Repository) Save(ctx context.Context, id, body string, expected int64) (Document, error) {
	if err := ctx.Err(); err != nil {
		return Document{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.docs[id]
	if !ok {
		return Document{}, fmt.Errorf("document not found")
	}
	if expected != d.Version {
		return Document{}, fmt.Errorf("optimistic lock conflict")
	}
	d.Body = body
	d.Version++
	d.UpdatedAt = time.Now()
	r.docs[id] = d
	return d, nil
}
func (r *Repository) Publish(ctx context.Context, id string) (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.docs[id]
	if !ok {
		return Document{}, fmt.Errorf("document not found")
	}
	d.Status = Published
	d.UpdatedAt = time.Now()
	r.docs[id] = d
	return d, nil
}
