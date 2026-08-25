package attachment

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Chunk struct {
	Number       int
	Offset, Size int64
	UploadedAt   time.Time
	Checksum     string
}
type Upload struct {
	ID, DocumentID, Name, ContentType string
	Size, Uploaded                    int64
	TotalChunks                       int
	Completed                         bool
	CreatedAt, CompletedAt            *time.Time
	Chunks                            map[int]Chunk
}
type Manager struct {
	mu      sync.RWMutex
	uploads map[string]Upload
}

func NewManager() *Manager { return &Manager{uploads: map[string]Upload{}} }
func valid(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	return ctx.Err()
}
func (m *Manager) Start(ctx context.Context, doc, name, contentType string, size int64, total int) (Upload, error) {
	if err := valid(ctx); err != nil {
		return Upload{}, err
	}
	if doc == "" || name == "" || size < 0 || total <= 0 {
		return Upload{}, errors.New("invalid upload")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	u := Upload{ID: "upload-" + now.Format("20060102150405.000000000"), DocumentID: doc, Name: name, ContentType: contentType, Size: size, TotalChunks: total, CreatedAt: &now, Chunks: map[int]Chunk{}}
	m.uploads[u.ID] = u
	return u, nil
}
func (m *Manager) PutChunk(ctx context.Context, id string, c Chunk) (Upload, error) {
	if err := valid(ctx); err != nil {
		return Upload{}, err
	}
	if c.Number < 0 || c.Size < 0 {
		return Upload{}, errors.New("invalid chunk")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.uploads[id]
	if !ok {
		return Upload{}, errors.New("upload not found")
	}
	if c.Number >= u.TotalChunks {
		return Upload{}, errors.New("chunk out of range")
	}
	if _, exists := u.Chunks[c.Number]; !exists {
		u.Uploaded += c.Size
	}
	c.UploadedAt = time.Now()
	u.Chunks[c.Number] = c
	m.uploads[id] = u
	return u, nil
}
func (m *Manager) Complete(ctx context.Context, id string) (Upload, error) {
	if err := valid(ctx); err != nil {
		return Upload{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.uploads[id]
	if !ok {
		return Upload{}, errors.New("upload not found")
	}
	if len(u.Chunks) != u.TotalChunks {
		return Upload{}, errors.New("upload incomplete")
	}
	if u.Uploaded != u.Size {
		return Upload{}, errors.New("uploaded size mismatch")
	}
	now := time.Now()
	u.Completed = true
	u.CompletedAt = &now
	m.uploads[id] = u
	return u, nil
}
func (m *Manager) Get(ctx context.Context, id string) (Upload, error) {
	if err := valid(ctx); err != nil {
		return Upload{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.uploads[id]
	if !ok {
		return Upload{}, errors.New("upload not found")
	}
	u.Chunks = u.Chunks
	return u, nil
}
func copyChunks(src map[int]Chunk) map[int]Chunk {
	out := map[int]Chunk{}
	for k, v := range src {
		out[k] = v
	}
	return out
}
func (m *Manager) MissingChunks(ctx context.Context, id string) ([]int, error) {
	u, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	out := []int{}
	for n := 0; n < u.TotalChunks; n++ {
		if _, ok := u.Chunks[n]; !ok {
			out = append(out, n)
		}
	}
	sort.Ints(out)
	return out, nil
}
func (m *Manager) Remove(ctx context.Context, id string) error {
	if err := valid(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.uploads[id]; !ok {
		return errors.New("upload not found")
	}
	delete(m.uploads, id)
	return nil
}
