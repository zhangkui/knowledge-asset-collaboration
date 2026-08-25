package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/catalog"
    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug11_FailedPublishKeepsDraftStatus(t *testing.T) {
    app := shared.NewApp()
    ctx := context.Background()
    workspace, err := app.Catalog.CreateWorkspace(ctx, "owner-11", "Engineering", "Knowledge", catalog.VisibilityOrganization)
    if err != nil {
        t.Fatalf("create workspace: %v", err)
    }
    document, err := app.Catalog.CreateDocument(ctx, "owner-11", workspace.ID, "", "Draft handbook", "Unreviewed", "draft body")
    if err != nil {
        t.Fatalf("create document: %v", err)
    }
    if _, err := app.PublishDocument(ctx, "owner-11", document.ID); err == nil {
        t.Fatal("unreviewed document publish must return an error")
    }
    current, err := app.Catalog.GetDocument(ctx, document.ID)
    if err != nil {
        t.Fatalf("get document after failed publish: %v", err)
    }
    if current.Status != catalog.DocumentDraft {
        t.Fatalf("failed publish changed document status to %q", current.Status)
    }
}
