package shared

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func requestJSON(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}
func TestDocumentLifecycleThroughHTTP(t *testing.T) {
	app := NewApp()
	h := app.Router()
	login := requestJSON(t, h, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"email": "owner@example.com", "password": "secret"})
	if login.Code != 200 {
		t.Fatalf("login status=%d body=%s", login.Code, login.Body)
	}
	var tokens struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &tokens); err != nil {
		t.Fatal(err)
	}
	workspace := requestJSON(t, h, http.MethodPost, "/api/v1/workspaces", tokens.AccessToken, map[string]string{"name": "Engineering", "description": "Technical knowledge", "visibility": "organization"})
	if workspace.Code != 201 {
		t.Fatalf("workspace status=%d body=%s", workspace.Code, workspace.Body)
	}
	var ws struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(workspace.Body.Bytes(), &ws); err != nil {
		t.Fatal(err)
	}
	doc := requestJSON(t, h, http.MethodPost, "/api/v1/documents", tokens.AccessToken, map[string]any{"workspace_id": ws.ID, "title": "Runbook", "summary": "Deploy service", "body": "step one"})
	if doc.Code != 201 {
		t.Fatalf("document status=%d body=%s", doc.Code, doc.Body)
	}
	var d struct {
		ID      string `json:"id"`
		Version int64  `json:"version"`
	}
	if err := json.Unmarshal(doc.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	saved := requestJSON(t, h, http.MethodPatch, "/api/v1/documents/"+d.ID+"/draft", tokens.AccessToken, map[string]any{"body": "step one\nstep two", "expected_version": d.Version})
	if saved.Code != 200 {
		t.Fatalf("save status=%d body=%s", saved.Code, saved.Body)
	}
	found := requestJSON(t, h, http.MethodGet, "/api/v1/search?q=step", tokens.AccessToken, nil)
	if found.Code != 200 {
		t.Fatalf("search status=%d body=%s", found.Code, found.Body)
	}
}
func TestUnauthorizedRequest(t *testing.T) {
	app := NewApp()
	w := httptest.NewRecorder()
	app.Router().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil))
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}
