package report

import (
	"context"
	"time"
)

type Metrics struct {
	Documents, Published, Reads, ActiveUsers, Edits, Comments, StorageBytes int64
	GeneratedAt                                                             time.Time
}

func Build(ctx context.Context) (Metrics, error) {
	if err := ctx.Err(); err != nil {
		return Metrics{}, err
	}
	return Metrics{GeneratedAt: time.Now()}, nil
}
