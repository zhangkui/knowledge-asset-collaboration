package shared

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zhangkui/knowledge-asset-collaboration/internal/auth"
	"github.com/zhangkui/knowledge-asset-collaboration/internal/catalog"
	"github.com/zhangkui/knowledge-asset-collaboration/internal/notification"
)

type App struct {
	Catalog   *catalog.Service
	Auth      auth.Service
	NotificationCenter *notification.Center
	StartedAt time.Time
}

func NewApp() *App {
	return &App{Catalog: catalog.NewService(), Auth: auth.NewService("development-secret-change-me"), NotificationCenter: notification.NewCenter(), StartedAt: time.Now()}
}
func (a *App) PublishNotification(ctx context.Context, n notification.Notification) error {
	if a.NotificationCenter == nil {
			a.NotificationCenter = notification.NewCenter()
	}
	return a.NotificationCenter.Push(ctx, n)
}

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.health)
	mux.HandleFunc("/api/v1/auth/login", a.login)
	mux.HandleFunc("/api/v1/auth/refresh", a.refresh)
	mux.HandleFunc("/api/v1/ws", a.websocket)
	mux.Handle("/api/v1/", a.Auth.Require(http.HandlerFunc(a.api)))
	return logging(mux)
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
	})
}
func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok", "service": "knowledge-asset-collaboration", "database": map[string]string{"driver": "postgresql", "status": "configured"}, "cache": map[string]string{"driver": "redis", "status": "configured"}, "uptime_seconds": int(time.Since(a.StartedAt).Seconds())})
}
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decode(r, &in); err != nil || strings.TrimSpace(in.Email) == "" || in.Password == "" {
		writeError(w, 400, "email and password are required")
		return
	}
	access, err := a.Auth.Issue(strings.ToLower(strings.TrimSpace(in.Email)), a.Auth.AccessTTL)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	refresh, err := a.Auth.Issue(strings.ToLower(strings.TrimSpace(in.Email)), a.Auth.RefreshTTL)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"access_token": access, "refresh_token": refresh, "token_type": "Bearer", "expires_in": int(a.Auth.AccessTTL.Seconds())})
}
func (a *App) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decode(r, &in); err != nil {
		writeError(w, 400, "invalid request")
		return
	}
	claims, err := a.Auth.Parse(in.RefreshToken)
	if err != nil {
		writeError(w, 401, err.Error())
		return
	}
	token, err := a.Auth.Issue(claims.Subject, a.Auth.AccessTTL)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": int(a.Auth.AccessTTL.Seconds())})
}
func (a *App) api(w http.ResponseWriter, r *http.Request) {
	claims, err := a.Auth.Authorize(r)
	if err != nil {
		writeError(w, 401, err.Error())
		return
	}
	parts := pathParts(r.URL.Path)
	if len(parts) < 3 {
		writeError(w, 404, "not found")
		return
	}
	resource := parts[2]
	tail := []string{}
	if len(parts) > 4 {
		tail = parts[4:]
	}
	id := ""
	if len(parts) > 3 {
		id = parts[3]
	}
	switch resource {
	case "workspaces":
		a.workspaces(w, r, claims.Subject, id, tail)
	case "folders":
		a.folders(w, r, claims.Subject, id)
	case "documents":
		a.documents(w, r, claims.Subject, id, tail)
	case "versions":
		a.versions(w, r, claims.Subject, id)
	case "comments":
		a.comments(w, r, claims.Subject, id)
	case "reviews":
		a.reviews(w, r, claims.Subject, id)
	case "search":
		a.search(w, r, claims.Subject)
	case "notifications":
		a.notifications(w, r, claims.Subject)
	case "shares":
		a.shares(w, r, claims.Subject, id)
	case "audit-logs":
		a.audit(w, r, claims.Subject)
	case "recycle-bin":
		a.recycle(w, r, claims.Subject, id)
	case "reports":
		a.reports(w, r, claims.Subject)
	case "attachments":
		a.attachments(w, r, claims.Subject, id)
	case "annotations":
		a.annotations(w, r, claims.Subject, id)
	case "tags":
		a.tags(w, r, claims.Subject, id)
	default:
		writeError(w, 404, "resource not found")
	}
}
func (a *App) workspaces(w http.ResponseWriter, r *http.Request, user, id string, tail []string) {
	switch r.Method {
	case "GET":
		if id != "" {
			v, err := a.Catalog.GetWorkspace(r.Context(), id)
			if err != nil {
				writeDomainError(w, err)
				return
			}
			writeJSON(w, 200, v)
			return
		}
		items, err := a.Catalog.ListWorkspaces(r.Context(), user)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
	case "POST":
		var in struct {
			Name        string                      `json:"name"`
			Description string                      `json:"description"`
			Visibility  catalog.WorkspaceVisibility `json:"visibility"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.CreateWorkspace(r.Context(), user, in.Name, in.Description, in.Visibility)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, v)
	case "PATCH":
		var in struct {
			Name        string                      `json:"name"`
			Description string                      `json:"description"`
			Visibility  catalog.WorkspaceVisibility `json:"visibility"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.UpdateWorkspace(r.Context(), user, id, in.Name, in.Description, in.Visibility)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	default:
		writeError(w, 405, "method not allowed")
	}
}
func (a *App) folders(w http.ResponseWriter, r *http.Request, user, id string) {
	switch r.Method {
	case "GET":
		workspaceID := r.URL.Query().Get("workspace_id")
		items, err := a.Catalog.ListFolders(r.Context(), workspaceID, r.URL.Query().Get("parent_id"))
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
	case "POST":
		var in struct {
			WorkspaceID string `json:"workspace_id"`
			ParentID    string `json:"parent_id"`
			Name        string `json:"name"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.CreateFolder(r.Context(), user, in.WorkspaceID, in.ParentID, in.Name)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, v)
	case "PATCH":
		var in struct {
			ParentID string `json:"parent_id"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.MoveFolder(r.Context(), user, id, in.ParentID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	default:
		writeError(w, 405, "method not allowed")
	}
}
func (a *App) documents(w http.ResponseWriter, r *http.Request, user, id string, tail []string) {
	action := ""
	if len(tail) > 0 {
		action = tail[0]
	}
	if id == "" && r.Method == "GET" {
		items, err := a.Catalog.ListDocuments(r.Context(), user, r.URL.Query().Get("workspace_id"), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
		return
	}
	if id == "" && r.Method == "POST" {
		var in struct {
			WorkspaceID string `json:"workspace_id"`
			FolderID    string `json:"folder_id"`
			Title       string `json:"title"`
			Summary     string `json:"summary"`
			Body        string `json:"body"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.CreateDocument(r.Context(), user, in.WorkspaceID, in.FolderID, in.Title, in.Summary, in.Body)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, v)
		return
	}
	if id == "" {
		writeError(w, 404, "document id is required")
		return
	}
	switch action {
	case "":
		if r.Method != "GET" {
			writeError(w, 405, "method not allowed")
			return
		}
		v, err := a.Catalog.GetDocument(r.Context(), id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		_ = a.Catalog.RecordRead(r.Context(), user, id)
		writeJSON(w, 200, v)
	case "draft":
		if r.Method != "PUT" && r.Method != "PATCH" {
			writeError(w, 405, "method not allowed")
			return
		}
		var in struct {
			Body            string `json:"body"`
			ExpectedVersion int64  `json:"expected_version"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.SaveDraft(r.Context(), user, id, in.Body, in.ExpectedVersion)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	case "submit-review":
		if r.Method != "POST" {
			writeError(w, 405, "method not allowed")
			return
		}
		var in struct {
			ReviewerID string `json:"reviewer_id"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.SubmitReview(r.Context(), user, id, in.ReviewerID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, v)
	case "comments":
		a.comments(w, r, user, id)
	case "flags":
		var in struct {
			Favorite *bool `json:"favorite"`
			Pinned   *bool `json:"pinned"`
		}
		if r.Method != "PATCH" {
			writeError(w, 405, "method not allowed")
			return
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.SetDocumentFlag(r.Context(), user, id, in.Favorite, in.Pinned)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	case "status":
		if r.Method != "PATCH" {
			writeError(w, 405, "method not allowed")
			return
		}
		var in struct {
			Status catalog.DocumentStatus `json:"status"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.ChangeDocumentStatus(r.Context(), user, id, in.Status)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	case "move":
		if r.Method != "POST" {
			writeError(w, 405, "method not allowed")
			return
		}
		var in struct {
			FolderID string `json:"folder_id"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.MoveDocument(r.Context(), user, id, in.FolderID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	case "copy":
		if r.Method != "POST" {
			writeError(w, 405, "method not allowed")
			return
		}
		var in struct {
			FolderID string `json:"folder_id"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.CopyDocument(r.Context(), user, id, in.FolderID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, v)
	case "delete":
		if r.Method != "DELETE" {
			writeError(w, 405, "method not allowed")
			return
		}
		v, err := a.Catalog.DeleteDocument(r.Context(), user, id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	case "versions":
		items, err := a.Catalog.ListVersions(r.Context(), id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
	default:
		writeError(w, 404, "document action not found")
	}
}
func (a *App) versions(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	items, err := a.Catalog.ListVersions(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writePage(w, items, len(items))
}
func (a *App) comments(w http.ResponseWriter, r *http.Request, user, id string) {
	switch r.Method {
	case "GET":
		items, err := a.Catalog.ListComments(r.Context(), id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
	case "POST":
		var in struct {
			Body     string `json:"body"`
			ParentID string `json:"parent_id"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.AddComment(r.Context(), user, id, in.ParentID, in.Body)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, v)
	case "PATCH":
		var in struct {
			Resolved bool `json:"resolved"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.ResolveComment(r.Context(), user, id, in.Resolved)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, v)
	default:
		writeError(w, 405, "method not allowed")
	}
}
func (a *App) reviews(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	var in struct {
		State   string `json:"state"`
		Opinion string `json:"opinion"`
	}
	if err := decode(r, &in); err != nil {
		writeError(w, 400, "invalid request")
		return
	}
	v, err := a.Catalog.DecideReview(r.Context(), user, id, in.State, in.Opinion)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, 200, v)
}
func (a *App) search(w http.ResponseWriter, r *http.Request, user string) {
	items, err := a.Catalog.Search(r.Context(), user, r.URL.Query().Get("q"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writePage(w, items, len(items))
}
func (a *App) notifications(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/read") {
		if err := a.Catalog.MarkNotificationsRead(r.Context(), user); err != nil {
			writeDomainError(w, err)
			return
		}
		w.WriteHeader(204)
		return
	}
	items, err := a.Catalog.ListNotifications(r.Context(), user, r.URL.Query().Get("unread") == "true")
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writePage(w, items, len(items))
}
func (a *App) shares(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method == "GET" && id != "" {
		share, doc, err := a.Catalog.ResolveShare(r.Context(), id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"share": share, "document": doc})
		return
	}
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	var in struct {
		DocumentID string `json:"document_id"`
		Permission string `json:"permission"`
		TTLSeconds int    `json:"ttl_seconds"`
	}
	if err := decode(r, &in); err != nil {
		writeError(w, 400, "invalid request")
		return
	}
	v, err := a.Catalog.CreateShare(r.Context(), user, in.DocumentID, in.Permission, time.Duration(in.TTLSeconds)*time.Second)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, 201, v)
}
func (a *App) tags(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method == "POST" {
		var in struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		v, err := a.Catalog.CreateTag(r.Context(), user, in.Name, in.Color)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, v)
		return
	}
	writeError(w, 405, "method not allowed")
}
func (a *App) audit(w http.ResponseWriter, r *http.Request, user string) {
	items, err := a.Catalog.AuditLogs(r.Context(), r.URL.Query().Get("actor_id"), r.URL.Query().Get("action"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writePage(w, items, len(items))
}
func (a *App) websocket(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" {
		writeError(w, 426, "websocket upgrade required")
		return
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		writeError(w, 400, "websocket key required")
		return
	}
	h, ok := w.(http.Hijacker)
	if !ok {
		writeError(w, 500, "websocket unavailable")
		return
	}
	accept := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	conn, buf, err := h.Hijack()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer conn.Close()
	fmt.Fprintf(buf, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", base64.StdEncoding.EncodeToString(accept[:]))
	_ = buf.Flush()
	for {
		if _, _, err := buf.ReadLine(); err != nil {
			return
		}
	}
}
func pathParts(path string) []string { return strings.Split(strings.Trim(path, "/"), "/") }
func decode(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
func writePage(w http.ResponseWriter, items any, total int) {
	writeJSON(w, 200, map[string]any{"items": items, "page": 1, "page_size": total, "total": total})
}
func writeDomainError(w http.ResponseWriter, err error) {
	status := 400
	if strings.Contains(err.Error(), "not found") {
		status = 404
	}
	if strings.Contains(err.Error(), "permission") {
		status = 403
	}
	if strings.Contains(err.Error(), "conflict") {
		status = 409
	}
	if errors.Is(err, contextCanceled) {
		status = 499
	}
	writeError(w, status, err.Error())
}

var contextCanceled = errors.New("context canceled")

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": http.StatusText(status), "message": message}})
}

var _ = strconv.IntSize
