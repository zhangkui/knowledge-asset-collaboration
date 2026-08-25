package regression

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func bug19Request(t *testing.T, handler http.Handler, token, path string) *httptest.ResponseRecorder {
    t.Helper()
    request := httptest.NewRequest(http.MethodGet, path, bytes.NewReader(nil))
    request.Header.Set("Authorization", "Bearer "+token)
    response := httptest.NewRecorder()
    handler.ServeHTTP(response, request)
    return response
}

func TestBug19_ReportLimitDoesNotPolluteRankingState(t *testing.T) {
    app := shared.NewApp()
    ctx := context.Background()
    for index := 0; index < 3; index++ {
        id := "doc-19-" + string(rune('a'+index))
        if err := app.ReportAggregator.AddDocument(ctx, "", id, "Document "+string(rune('a'+index)), "author-19", false); err != nil {
            t.Fatalf("add report document: %v", err)
        }
    }
    token, err := app.Auth.Issue("reader-19", app.Auth.AccessTTL)
    if err != nil {
        t.Fatalf("issue token: %v", err)
    }
    limited := bug19Request(t, app.Router(), token, "/api/v1/reports?source=aggregator&top_limit=1")
    if limited.Code != http.StatusOK {
        t.Fatalf("limited report status=%d body=%s", limited.Code, limited.Body.String())
    }
    full := bug19Request(t, app.Router(), token, "/api/v1/reports?source=aggregator")
    if full.Code != http.StatusOK {
        t.Fatalf("full report status=%d body=%s", full.Code, full.Body.String())
    }
    var dashboard struct { TopDocuments []struct { ID string `json:"ID"` } `json:"TopDocuments"` }
    if err := json.Unmarshal(full.Body.Bytes(), &dashboard); err != nil {
        t.Fatalf("decode full report: %v", err)
    }
    if len(dashboard.TopDocuments) != 3 {
        t.Fatalf("limited report polluted internal ranking state: got %d documents", len(dashboard.TopDocuments))
    }
}
