package regression

import (
    "context"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func openShareWithoutPanic(app *shared.App, token string) (err error, panicked any) {
    defer func() {
        panicked = recover()
    }()
    _, err = app.OpenShare(context.Background(), token)
    return err, nil
}

func TestBug13_InvalidShareTokensReturnErrorsWithoutPanic(t *testing.T) {
    app := shared.NewApp()
    for _, token := range []string{"", "missing-share-13"} {
        err, panicked := openShareWithoutPanic(app, token)
        if panicked != nil {
            t.Fatalf("invalid share token %q must not panic: %v", token, panicked)
        }
        if err == nil {
            t.Fatalf("invalid share token %q must return an error", token)
        }
    }
}
