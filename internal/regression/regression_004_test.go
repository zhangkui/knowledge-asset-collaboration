package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/editor"
)

func TestBug04_OnlinePresenceIncludesEveryCollaborator(t *testing.T) {
    hub := editor.NewHub()
    ctx := context.Background()
    if err := hub.Join(ctx, "doc-4", editor.Presence{UserID: "editor-a", DocumentID: "doc-4"}); err != nil {
        t.Fatalf("join editor-a: %v", err)
    }
    if err := hub.Join(ctx, "doc-4", editor.Presence{UserID: "editor-b", DocumentID: "doc-4"}); err != nil {
        t.Fatalf("join editor-b: %v", err)
    }
    online := hub.Online("doc-4")
    if len(online) != 2 {
        t.Fatalf("online collaborators must include both users, got %d", len(online))
    }
    seen := map[string]bool{}
    for _, presence := range online {
        seen[presence.UserID] = true
    }
    if !seen["editor-a"] || !seen["editor-b"] {
        t.Fatalf("online collaborator list lost a user: %#v", seen)
    }
}
