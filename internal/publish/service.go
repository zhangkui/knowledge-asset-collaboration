package publish

import (
	"context"
	"fmt"
)

type Service struct{}

func (Service) Publish(ctx context.Context, status string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if status != "approved" {
		return "", fmt.Errorf("only approved versions may be published")
	}
	return "published", nil
}
