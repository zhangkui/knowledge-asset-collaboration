package regression

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func TestBug08_MissingAuthorizationReturnsUnauthorized(t *testing.T) {
    handler := shared.NewApp().Router()
    request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
    response := httptest.NewRecorder()
    defer func() {
        if recovered := recover(); recovered != nil {
            t.Fatalf("missing Authorization must return 401 instead of panic: %v", recovered)
        }
    }()
    handler.ServeHTTP(response, request)
    if response.Code != http.StatusUnauthorized {
        t.Fatalf("missing Authorization must return 401, got %d body=%s", response.Code, response.Body.String())
    }
}
