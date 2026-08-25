package regression

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/zhangkui/knowledge-asset-collaboration/internal/shared"
)

func bug06Request(t *testing.T, handler http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
    t.Helper()
    payload, err := json.Marshal(body)
    if err != nil {
        t.Fatalf("marshal request: %v", err)
    }
    req := httptest.NewRequest(method, path, bytes.NewReader(payload))
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    response := httptest.NewRecorder()
    handler.ServeHTTP(response, req)
    return response
}

func TestBug06_VersionConflictKeepsHTTPConflictStatus(t *testing.T) {
    handler := shared.NewApp().Router()
    login := bug06Request(t, handler, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"email": "owner@example.com", "password": "secret"})
    if login.Code != http.StatusOK {
        t.Fatalf("login status=%d body=%s", login.Code, login.Body.String())
    }
    var auth struct { AccessToken string `json:"access_token"` }
    if err := json.Unmarshal(login.Body.Bytes(), &auth); err != nil {
        t.Fatalf("decode login: %v", err)
    }
    workspace := bug06Request(t, handler, http.MethodPost, "/api/v1/workspaces", auth.AccessToken, map[string]string{"name": "Engineering", "description": "Knowledge", "visibility": "organization"})
    if workspace.Code != http.StatusCreated {
        t.Fatalf("workspace status=%d body=%s", workspace.Code, workspace.Body.String())
    }
    var created struct { ID string `json:"id"` }
    if err := json.Unmarshal(workspace.Body.Bytes(), &created); err != nil {
        t.Fatalf("decode workspace: %v", err)
    }
    document := bug06Request(t, handler, http.MethodPost, "/api/v1/documents", auth.AccessToken, map[string]any{"workspace_id": created.ID, "title": "Runbook", "summary": "Release", "body": "initial"})
    if document.Code != http.StatusCreated {
        t.Fatalf("document status=%d body=%s", document.Code, document.Body.String())
    }
    var doc struct { ID string `json:"id"`; Version int64 `json:"version"` }
    if err := json.Unmarshal(document.Body.Bytes(), &doc); err != nil {
        t.Fatalf("decode document: %v", err)
    }
    first := bug06Request(t, handler, http.MethodPatch, "/api/v1/documents/"+doc.ID+"/draft", auth.AccessToken, map[string]any{"body": "new", "expected_version": doc.Version})
    if first.Code != http.StatusOK {
        t.Fatalf("first save status=%d body=%s", first.Code, first.Body.String())
    }
    stale := bug06Request(t, handler, http.MethodPatch, "/api/v1/documents/"+doc.ID+"/draft", auth.AccessToken, map[string]any{"body": "stale", "expected_version": doc.Version})
    if stale.Code != http.StatusConflict {
        t.Fatalf("version conflict must map to HTTP 409, got %d body=%s", stale.Code, stale.Body.String())
    }
}
