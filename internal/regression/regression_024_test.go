package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug24_MissingShareReturnsErrorInsteadOfPanicking(t *testing.T) {
    app := shared.NewApp()
    defer func() { if recovered := recover(); recovered != nil { t.Fatalf("missing share panicked: %v", recovered) } }()
    if _, err := app.OpenShare(context.Background(), "missing-share-24"); err == nil { t.Fatal("missing share must return an error") }
}
