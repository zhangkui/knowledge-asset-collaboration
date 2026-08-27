package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
    "github.com/zhangkui/knowledge-asset-collaboration/internal/search"
)

func TestBug15_CancelledSearchReturnsNoResults(t *testing.T) {
    app := shared.NewApp()
    app.SearchIndex.Add(search.Result{DocumentID: "doc-15", Title: "Release Guide", Snippet: "deployment instructions"})
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    results := app.SearchIndexed(ctx, "release")
    if results != nil {
        t.Fatalf("cancelled search must return no result collection, got %#v", results)
    }
}
