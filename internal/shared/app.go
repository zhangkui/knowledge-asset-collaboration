package shared

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Entity struct {
	ID          string
	Name        string
	Description string
	Status      string
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int64
}
type AuditEvent struct {
	ID, ActorID, Action, ObjectType, ObjectID string
	At                                        time.Time
	Metadata                                  map[string]any
}
type Store struct {
	mu       sync.RWMutex
	entities map[string]Entity
	audit    []AuditEvent
}
type App struct {
	Store  *Store
	secret []byte
}

func NewApp() *App {
	return &App{Store: &Store{entities: map[string]Entity{}}, secret: []byte("development-secret-change-me")}
}
func (a *App) Router() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", a.health)
	m.HandleFunc("/api/v1/auth/login", a.login)
	m.HandleFunc("/api/v1/auth/refresh", a.refresh)
	m.HandleFunc("/api/v1/", a.api)
	return logging(m)
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(st))
	})
}
func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok", "database": "configured", "redis": "configured", "time": time.Now()})
}
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var q struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&q) != nil || q.Email == "" {
		writeError(w, 400, "invalid credentials")
		return
	}
	writeJSON(w, 200, map[string]any{"access_token": a.token(q.Email, 15*time.Minute), "refresh_token": a.token(q.Email, 24*time.Hour), "expires_in": 900})
}
func (a *App) refresh(w http.ResponseWriter, r *http.Request) {
	var q struct{ RefreshToken string }
	_ = json.NewDecoder(r.Body).Decode(&q)
	if q.RefreshToken == "" {
		writeError(w, 401, "refresh token required")
		return
	}
	writeJSON(w, 200, map[string]any{"access_token": a.token("refreshed-user", 15*time.Minute), "expires_in": 900})
}
func (a *App) token(subject string, ttl time.Duration) string {
	exp := time.Now().Add(ttl).Unix()
	raw := fmt.Sprintf("%s.%d", subject, exp)
	s := sha256.Sum256(append(a.secret, []byte(raw)...))
	return raw + "." + hex.EncodeToString(s[:])
}
func (a *App) api(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(p) < 4 {
		writeError(w, 404, "not found")
		return
	}
	res, id := p[3], ""
	if len(p) > 4 {
		id = p[4]
	}
	switch res {
	case "users", "organizations", "workspaces", "folders", "documents", "versions", "comments", "annotations", "tags", "shares", "permissions", "reviews", "attachments", "notifications", "audit-logs", "recycle-bin", "reports", "tasks", "search":
		a.resource(w, r, res, id)
	default:
		writeError(w, 404, "resource not found")
	}
}
func (a *App) resource(w http.ResponseWriter, r *http.Request, res, id string) {
	switch r.Method {
	case "GET":
		a.list(w, res, id)
	case "POST":
		a.create(w, r, res)
	case "PUT", "PATCH":
		a.update(w, r, id)
	case "DELETE":
		a.delete(w, id)
	default:
		writeError(w, 405, "method not allowed")
	}
}
func (a *App) list(w http.ResponseWriter, res, id string) {
	a.Store.mu.RLock()
	defer a.Store.mu.RUnlock()
	out := []Entity{}
	for _, e := range a.Store.entities {
		if strings.HasPrefix(e.Name, res+":") && (id == "" || e.ID == id) {
			out = append(out, e)
		}
	}
	writeJSON(w, 200, map[string]any{"items": out, "page": 1, "page_size": len(out), "total": len(out)})
}
func (a *App) create(w http.ResponseWriter, r *http.Request, res string) {
	var q struct{ Name, Description, OwnerID string }
	if json.NewDecoder(r.Body).Decode(&q) != nil {
		writeError(w, 400, "invalid json")
		return
	}
	now := time.Now()
	id := fmt.Sprintf("%s-%d", res, now.UnixNano())
	e := Entity{ID: id, Name: res + ":" + q.Name, Description: q.Description, OwnerID: q.OwnerID, Status: "active", CreatedAt: now, UpdatedAt: now, Version: 1}
	a.Store.mu.Lock()
	a.Store.entities[id] = e
	a.Store.audit = append(a.Store.audit, AuditEvent{ID: id + "-audit", Action: "create", ObjectType: res, ObjectID: id, At: now})
	a.Store.mu.Unlock()
	writeJSON(w, 201, e)
}
func (a *App) update(w http.ResponseWriter, r *http.Request, id string) {
	a.Store.mu.Lock()
	defer a.Store.mu.Unlock()
	e, ok := a.Store.entities[id]
	if !ok {
		writeError(w, 404, "not found")
		return
	}
	var q map[string]any
	if json.NewDecoder(r.Body).Decode(&q) != nil {
		writeError(w, 400, "invalid json")
		return
	}
	if v, ok := q["description"].(string); ok {
		e.Description = v
	}
	if v, ok := q["status"].(string); ok {
		e.Status = v
	}
	e.Version++
	e.UpdatedAt = time.Now()
	a.Store.entities[id] = e
	writeJSON(w, 200, e)
}
func (a *App) delete(w http.ResponseWriter, id string) {
	a.Store.mu.Lock()
	defer a.Store.mu.Unlock()
	if _, ok := a.Store.entities[id]; !ok {
		writeError(w, 404, "not found")
		return
	}
	delete(a.Store.entities, id)
	w.WriteHeader(204)
}
func writeJSON(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, s int, m string) {
	writeJSON(w, s, map[string]any{"error": map[string]string{"code": http.StatusText(s), "message": m}})
}
