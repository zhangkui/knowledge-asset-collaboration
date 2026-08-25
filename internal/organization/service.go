package organization

import (
	"context"
	"fmt"
	"strings"
)

type Department struct{ ID, Name, ParentID string }
type Team struct {
	ID, Name string
	Members  []string
}

func NewDepartment(ctx context.Context, name, parent string) (Department, error) {
	if err := ctx.Err(); err != nil {
		return Department{}, err
	}
	if strings.TrimSpace(name) == "" {
		return Department{}, fmt.Errorf("department name required")
	}
	return Department{ID: "department-" + name, Name: name, ParentID: parent}, nil
}
func AddMember(t Team, user string) Team {
	for _, m := range t.Members {
		if m == user {
			return t
		}
	}
	t.Members = append(t.Members, user)
	return t
}
