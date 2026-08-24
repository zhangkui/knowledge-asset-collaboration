package regression

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/zhangkui/knowledge-asset-collaboration/internal/organization"
)

func TestBug004ConcurrentTeamMemberUpdates(t *testing.T) {
	service := organization.NewService()
	if _, err := service.CreateTeam(context.Background(), "team-1", "Editors", "owner"); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for n := 0; n < 100; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if _, err := service.AddToTeam(context.Background(), "team-1", fmt.Sprintf("user-%03d", n)); err != nil {
				t.Errorf("add member %d: %v", n, err)
			}
		}(n)
	}
	wg.Wait()
	members, err := service.TeamMembers(context.Background(), "team-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 101 {
		t.Fatalf("members=%d, want 101", len(members))
	}
}
