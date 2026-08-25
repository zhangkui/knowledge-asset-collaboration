package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
    "github.com/zhangkui/knowledge-asset-collaboration/internal/user"
)

func TestBug09_RolePermissionResultsDoNotMutateDefinitions(t *testing.T) {
    app := shared.NewApp()
    ctx := context.Background()
    account, err := app.Directory.Register(ctx, user.User{ID: "user-9", Email: "owner-9@example.com", Name: "Owner Nine"})
    if err != nil {
        t.Fatalf("register user: %v", err)
    }
    role, err := app.Directory.DefineRole(ctx, "Document reviewer", []string{"read", "review"}, false)
    if err != nil {
        t.Fatalf("define role: %v", err)
    }
    if err := app.Directory.AssignRole(ctx, account.ID, role.ID); err != nil {
        t.Fatalf("assign role: %v", err)
    }
    roles, err := app.RolesForUser(ctx, account.ID)
    if err != nil {
        t.Fatalf("list roles: %v", err)
    }
    if len(roles) != 1 || len(roles[0].Permissions) != 2 {
        t.Fatalf("unexpected assigned roles: %#v", roles)
    }
    roles[0].Permissions[0] = "admin"
    again, err := app.RolesForUser(ctx, account.ID)
    if err != nil {
        t.Fatalf("list roles after caller mutation: %v", err)
    }
    if len(again) != 1 || len(again[0].Permissions) != 2 || again[0].Permissions[0] != "read" {
        t.Fatalf("caller mutation polluted role definition: %#v", again)
    }
}
