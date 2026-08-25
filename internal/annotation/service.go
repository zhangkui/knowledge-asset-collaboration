package annotation

import (
	"context"
	"fmt"
	"time"
)

type Annotation struct {
	ID, DocumentID, AuthorID, Quote, Note string
	StartOffset, EndOffset                int
	CreatedAt                             time.Time
}

func Create(ctx context.Context, a Annotation) (Annotation, error) {
	if err := ctx.Err(); err != nil {
		return Annotation{}, err
	}
	if a.StartOffset < 0 || a.EndOffset < a.StartOffset {
		return Annotation{}, fmt.Errorf("invalid range")
	}
	a.ID = fmt.Sprintf("annotation-%d", time.Now().UnixNano())
	a.CreatedAt = time.Now()
	return a, nil
}
