package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/catalog"
)

func TestBug21_InvalidReviewStateDoesNotMutateRecord(t *testing.T) {
    service := catalog.NewService()
    ctx := context.Background()
    workspace, err := service.CreateWorkspace(ctx, "owner-21", "Engineering", "Knowledge", catalog.VisibilityOrganization)
    if err != nil {
        t.Fatalf("create workspace: %v", err)
    }
    document, err := service.CreateDocument(ctx, "owner-21", workspace.ID, "", "Review policy", "Approval", "draft")
    if err != nil {
        t.Fatalf("create document: %v", err)
    }
    review, err := service.SubmitReview(ctx, "owner-21", document.ID, "reviewer-21")
    if err != nil {
        t.Fatalf("submit review: %v", err)
    }
    if _, err := service.DecideReview(ctx, "reviewer-21", review.ID, "unknown", "invalid"); err == nil {
        t.Fatal("invalid review state must return an error")
    }
    current, err := service.GetDocument(ctx, document.ID)
    if err != nil {
        t.Fatalf("get document: %v", err)
    }
    if current.Status != catalog.DocumentReviewing {
        t.Fatalf("invalid review decision changed document status to %q", current.Status)
    }
}
