package search

import (
	"context"
	"strings"
	"time"
)

type Result struct {
	DocumentID, Title, Snippet string
	Score                      float64
	UpdatedAt                  time.Time
}
type Index struct{ docs []Result }

func NewIndex() *Index        { return &Index{} }
func (i *Index) Add(r Result) { i.docs = append(i.docs, r) }
func (i *Index) Query(ctx context.Context, q string) []Result {
	if err := ctx.Err(); err != nil {
		return nil
	}
	q = strings.ToLower(strings.TrimSpace(q))
	out := []Result{}
	for _, d := range i.docs {
		if strings.Contains(strings.ToLower(d.Title+" "+d.Snippet), q) {
			out = append(out, d)
		}
	}
	return out
}
