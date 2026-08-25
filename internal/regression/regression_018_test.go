package regression

import (
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/recycle"
    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug18_NilRecycleContextReturnsError(t *testing.T) {
    app := shared.NewApp()
    err := app.PutRecycleItem(nil, recycle.Item{ID: "recycle-18", ObjectID: "doc-18", ObjectType: "document", DeletedBy: "owner-18"})
    if err == nil {
        t.Fatal("nil recycle context must return an error")
    }
}
