package folder

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Folder struct {
	ID, WorkspaceID, ParentID, Name string
	CreatedAt                       time.Time
}

func Create(ctx context.Context, workspace, parent, name string) (Folder, error) {
	if err := ctx.Err(); err != nil {
		return Folder{}, err
	}
	if strings.TrimSpace(name) == "" {
		return Folder{}, fmt.Errorf("folder name required")
	}
	return Folder{ID: fmt.Sprintf("folder-%d", time.Now().UnixNano()), WorkspaceID: workspace, ParentID: parent, Name: name, CreatedAt: time.Now()}, nil
}
func CanMove(id, target string) bool {
	return id != "" && target != id && !strings.HasPrefix(target, id+"/")
}
