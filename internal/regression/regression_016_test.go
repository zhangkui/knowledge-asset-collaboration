package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/catalog"
    "github.com/zhangkui/knowledge-asset-collaboration/internal/permission"
)

func TestBug16_ExplicitDenyOverridesInheritedAllow(t *testing.T) {
    ctx := context.Background()
    service := catalog.NewService()
    gate := &permission.Service{}
    service.SetPermissionGate(gate)
    workspace, err := service.CreateWorkspace(ctx, "owner-16", "Engineering", "Knowledge", catalog.VisibilityOrganization)
    if err != nil {
        t.Fatalf("create workspace: %v", err)
    }
    document, err := service.CreateDocument(ctx, "owner-16", workspace.ID, "", "Policy", "Access", "original")
    if err != nil {
        t.Fatalf("create document: %v", err)
    }
    if err := gate.Grant(ctx, permission.Grant{SubjectID: "editor-16", ResourceID: "*", Permission: "edit"}); err != nil {
        t.Fatalf("grant inherited edit permission: %v", err)
    }
    if err := gate.Grant(ctx, permission.Grant{SubjectID: "editor-16", ResourceID: document.ID, Permission: "edit", ExplicitDeny: true}); err != nil {
        t.Fatalf("grant explicit edit deny: %v", err)
    }
    if _, err := service.SaveDraft(ctx, "editor-16", document.ID, "must not save", document.Version); err == nil {
        t.Fatal("explicit deny must prevent inherited edit permission from saving")
    }
    current, err := service.GetDocument(ctx, document.ID)
    if err != nil {
        t.Fatalf("get document after denied save: %v", err)
    }
    if current.Body != "original" || current.Version != document.Version {
        t.Fatalf("denied save mutated document: %#v", current)
    }
}
