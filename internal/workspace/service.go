package workspace

import (
	"context"
	"fmt"
	"time"
)

type Workspace struct {
	ID, Name, Visibility, OwnerID, Status string
	CreatedAt                             time.Time
}

func Create(ctx context.Context, name, owner string) (Workspace, error) {
	if err := ctx.Err(); err != nil {
		return Workspace{}, err
	}
	if name == "" {
		return Workspace{}, fmt.Errorf("workspace name required")
	}
	return Workspace{ID: fmt.Sprintf("ws-%d", time.Now().UnixNano()), Name: name, Visibility: "organization", OwnerID: owner, Status: "active", CreatedAt: time.Now()}, nil
}
