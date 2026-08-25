package publish

import (
	"context"
	"errors"
)

// ErrNotApproved is returned when a document that has not been approved is
// attempted to be published. It is a sentinel error so callers up the stack
// can detect publish failures with errors.Is while preserving the error chain.
var ErrNotApproved = errors.New("only approved versions may be published")

type Service struct{}

func (Service) Publish(ctx context.Context, status string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if status != "approved" {
		return "", ErrNotApproved
	}
	return "published", nil
}
