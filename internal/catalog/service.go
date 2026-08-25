package catalog

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zhangkui/knowledge-asset-collaboration/internal/permission"
	"github.com/zhangkui/knowledge-asset-collaboration/internal/review"
)

type WorkspaceVisibility string

const (
	VisibilityPrivate      WorkspaceVisibility = "private"
	VisibilityOrganization WorkspaceVisibility = "organization"
	VisibilityPublic       WorkspaceVisibility = "public"
)

type Workspace struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	OwnerID     string              `json:"owner_id"`
	Visibility  WorkspaceVisibility `json:"visibility"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type Folder struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	ParentID    string    `json:"parent_id,omitempty"`
	Name        string    `json:"name"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
}

type DocumentStatus string

const (
	DocumentDraft     DocumentStatus = "draft"
	DocumentReviewing DocumentStatus = "reviewing"
	DocumentPublished DocumentStatus = "published"
	DocumentArchived  DocumentStatus = "archived"
	DocumentDeleted   DocumentStatus = "deleted"
)

type Document struct {
	ID               string         `json:"id"`
	WorkspaceID      string         `json:"workspace_id"`
	FolderID         string         `json:"folder_id,omitempty"`
	Title            string         `json:"title"`
	Summary          string         `json:"summary"`
	Body             string         `json:"body"`
	AuthorID         string         `json:"author_id"`
	Status           DocumentStatus `json:"status"`
	Version          int64          `json:"version"`
	PublishedVersion int64          `json:"published_version"`
	Favorite         bool           `json:"favorite"`
	Pinned           bool           `json:"pinned"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type Version struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Number     int64     `json:"number"`
	Body       string    `json:"body"`
	Summary    string    `json:"summary"`
	AuthorID   string    `json:"author_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Comment struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	AuthorID   string    `json:"author_id"`
	ParentID   string    `json:"parent_id,omitempty"`
	Body       string    `json:"body"`
	Resolved   bool      `json:"resolved"`
	CreatedAt  time.Time `json:"created_at"`
}

type Review struct {
	ID         string     `json:"id"`
	DocumentID string     `json:"document_id"`
	Version    int64      `json:"version"`
	ReviewerID string     `json:"reviewer_id"`
	State      string     `json:"state"`
	Opinion    string     `json:"opinion"`
	CreatedAt  time.Time  `json:"created_at"`
	DecidedAt  *time.Time `json:"decided_at,omitempty"`
}

type PermissionGrant struct {
	SubjectID    string `json:"subject_id"`
	ResourceID   string `json:"resource_id"`
	Permission   string `json:"permission"`
	ExplicitDeny bool   `json:"explicit_deny"`
}

type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
type Share struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Token      string    `json:"token"`
	Permission string    `json:"permission"`
	ExpiresAt  time.Time `json:"expires_at"`
}
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Kind      string    `json:"kind"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}
type AuditLog struct {
	ID         string            `json:"id"`
	ActorID    string            `json:"actor_id"`
	Action     string            `json:"action"`
	ObjectType string            `json:"object_type"`
	ObjectID   string            `json:"object_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}
type SearchResult struct {
	Document  Document `json:"document"`
	Highlight string   `json:"highlight"`
}

type RecycleItem struct {
	ID, ObjectType, ObjectID, Name, DeletedBy string
	DeletedAt                                 time.Time
}
type Annotation struct {
	ID, DocumentID, AuthorID, Quote, Comment string
	Start, End                               int
	CreatedAt                                time.Time
}
type Attachment struct {
	ID, DocumentID, Name, ContentType string
	Size                              int64
	Completed                         bool
	CreatedAt                         time.Time
}
type Report struct {
	WorkspaceID                                                            string
	Documents, Published, Reads, Edits, Comments, Attachments, ActiveUsers int
}

type Service struct {
	mu             sync.RWMutex
	workspaces     map[string]Workspace
	folders        map[string]Folder
	documents      map[string]Document
	versions       map[string]Version
	comments       map[string]Comment
	reviews        map[string]Review
	grants         []PermissionGrant
	tags           map[string]Tag
	documentTags   map[string]map[string]bool
	shares         map[string]Share
	notifications  map[string]Notification
	audit          []AuditLog
	recent         map[string][]string
	recycle        map[string]RecycleItem
	annotations    map[string]Annotation
	attachments    map[string]Attachment
	reads          map[string]int
	permissionGate *permission.Service
	next           uint64
}

func NewService() *Service {
	return &Service{workspaces: map[string]Workspace{}, folders: map[string]Folder{}, documents: map[string]Document{}, versions: map[string]Version{}, comments: map[string]Comment{}, reviews: map[string]Review{}, tags: map[string]Tag{}, documentTags: map[string]map[string]bool{}, shares: map[string]Share{}, notifications: map[string]Notification{}, recent: map[string][]string{}, recycle: map[string]RecycleItem{}, annotations: map[string]Annotation{}, attachments: map[string]Attachment{}, reads: map[string]int{}}
}
func (s *Service) id(prefix string) string { s.next++; return fmt.Sprintf("%s-%d", prefix, s.next) }
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	return ctx.Err()
}

func (s *Service) CreateWorkspace(ctx context.Context, ownerID, name, description string, visibility WorkspaceVisibility) (Workspace, error) {
	if err := checkContext(ctx); err != nil {
		return Workspace{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Workspace{}, errors.New("workspace name is required")
	}
	if visibility == "" {
		visibility = VisibilityOrganization
	}
	if visibility != VisibilityPrivate && visibility != VisibilityOrganization && visibility != VisibilityPublic {
		return Workspace{}, errors.New("invalid visibility")
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	w := Workspace{ID: s.id("ws"), OwnerID: ownerID, Name: name, Description: description, Visibility: visibility, Status: "active", CreatedAt: now, UpdatedAt: now}
	s.workspaces[w.ID] = w
	s.recordLocked(ownerID, "workspace.created", "workspace", w.ID, nil)
	return w, nil
}
func (s *Service) ListWorkspaces(ctx context.Context, subjectID string) ([]Workspace, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Workspace, 0)
	for _, w := range s.workspaces {
		if w.OwnerID == subjectID || w.Visibility == VisibilityPublic || w.Visibility == VisibilityOrganization {
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (s *Service) GetWorkspace(ctx context.Context, id string) (Workspace, error) {
	if err := checkContext(ctx); err != nil {
		return Workspace{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.workspaces[id]
	if !ok {
		return Workspace{}, errors.New("workspace not found")
	}
	return w, nil
}
func (s *Service) UpdateWorkspace(ctx context.Context, actorID, id, name, description string, visibility WorkspaceVisibility) (Workspace, error) {
	if err := checkContext(ctx); err != nil {
		return Workspace{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.workspaces[id]
	if !ok {
		return Workspace{}, errors.New("workspace not found")
	}
	if w.OwnerID != actorID && !s.allowedLocked(actorID, id, "admin") {
		return Workspace{}, errors.New("workspace admin permission required")
	}
	if strings.TrimSpace(name) != "" {
		w.Name = strings.TrimSpace(name)
	}
	w.Description = description
	if visibility != "" {
		w.Visibility = visibility
	}
	w.UpdatedAt = time.Now()
	s.workspaces[id] = w
	s.recordLocked(actorID, "workspace.updated", "workspace", id, nil)
	return w, nil
}

func (s *Service) CreateFolder(ctx context.Context, actorID, workspaceID, parentID, name string) (Folder, error) {
	if err := checkContext(ctx); err != nil {
		return Folder{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Folder{}, errors.New("folder name is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workspaces[workspaceID]; !ok {
		return Folder{}, errors.New("workspace not found")
	}
	if parentID != "" {
		p, ok := s.folders[parentID]
		if !ok || p.WorkspaceID != workspaceID {
			return Folder{}, errors.New("parent folder not found")
		}
	}
	f := Folder{ID: s.id("folder"), WorkspaceID: workspaceID, ParentID: parentID, Name: name, Position: s.nextPositionLocked(workspaceID, parentID), CreatedAt: time.Now()}
	s.folders[f.ID] = f
	s.recordLocked(actorID, "folder.created", "folder", f.ID, nil)
	return f, nil
}
func (s *Service) ListFolders(ctx context.Context, workspaceID, parentID string) ([]Folder, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Folder{}
	for _, f := range s.folders {
		if f.WorkspaceID == workspaceID && f.ParentID == parentID {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}
func (s *Service) MoveFolder(ctx context.Context, actorID, id, targetParent string) (Folder, error) {
	if err := checkContext(ctx); err != nil {
		return Folder{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.folders[id]
	if !ok {
		return Folder{}, errors.New("folder not found")
	}
	if targetParent == id {
		return Folder{}, errors.New("folder cannot move into itself")
	}
	if targetParent != "" {
		p, ok := s.folders[targetParent]
		if !ok || p.WorkspaceID != f.WorkspaceID {
			return Folder{}, errors.New("target folder not found")
		}
		if s.isDescendantLocked(targetParent, id) {
			return Folder{}, errors.New("folder cannot move into descendant")
		}
	}
	f.ParentID = targetParent
	f.Position = s.nextPositionLocked(f.WorkspaceID, targetParent)
	s.folders[id] = f
	s.recordLocked(actorID, "folder.moved", "folder", id, map[string]string{"parent_id": targetParent})
	return f, nil
}
func (s *Service) isDescendantLocked(candidate, parent string) bool {
	for candidate != "" {
		if candidate == parent {
			return true
		}
		f, ok := s.folders[candidate]
		if !ok {
			return false
		}
		candidate = f.ParentID
	}
	return false
}
func (s *Service) nextPositionLocked(workspaceID, parentID string) int {
	n := 0
	for _, f := range s.folders {
		if f.WorkspaceID == workspaceID && f.ParentID == parentID && f.Position >= n {
			n = f.Position + 1
		}
	}
	return n
}

func (s *Service) CreateDocument(ctx context.Context, actorID, workspaceID, folderID, title, summary, body string) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return Document{}, errors.New("document title is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workspaces[workspaceID]; !ok {
		return Document{}, errors.New("workspace not found")
	}
	if folderID != "" {
		f, ok := s.folders[folderID]
		if !ok || f.WorkspaceID != workspaceID {
			return Document{}, errors.New("folder not found")
		}
	}
	now := time.Now()
	d := Document{ID: s.id("doc"), WorkspaceID: workspaceID, FolderID: folderID, Title: title, Summary: summary, Body: body, AuthorID: actorID, Status: DocumentDraft, Version: 1, CreatedAt: now, UpdatedAt: now}
	s.documents[d.ID] = d
	s.createVersionLocked(d, actorID, "initial draft")
	s.recordLocked(actorID, "document.created", "document", d.ID, nil)
	return d, nil
}
func (s *Service) ListDocuments(ctx context.Context, subjectID, workspaceID, query, status string) ([]Document, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	query = strings.ToLower(strings.TrimSpace(query))
	out := []Document{}
	for _, d := range s.documents {
		if workspaceID != "" && d.WorkspaceID != workspaceID {
			continue
		}
		if status != "" && string(d.Status) != status {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(d.Title+" "+d.Summary+" "+d.Body), query) {
			continue
		}
		if d.Status == DocumentDeleted {
			continue
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (s *Service) GetDocument(ctx context.Context, id string) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.documents[id]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	return d, nil
}
func (s *Service) SetPermissionGate(gate *permission.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.permissionGate = gate
}

func (s *Service) SaveDraft(ctx context.Context, actorID, id, body string, expectedVersion int64) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[id]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	editable := s.canEditLocked(actorID, d)
	if !editable && s.permissionGate != nil {
		editable = s.permissionGate.Allowed(ctx, actorID, d.ID, "edit") || s.permissionGate.Allowed(ctx, actorID, d.WorkspaceID, "edit")
	}
	if !editable {
		return Document{}, errors.New("document edit permission required")
	}
	if expectedVersion != d.Version {
		return Document{}, fmt.Errorf("version conflict: expected %d current %d", expectedVersion, d.Version)
	}
	d.Body = body
	d.Version++
	d.UpdatedAt = time.Now()
	if d.Status == DocumentPublished {
		d.Status = DocumentDraft
	}
	s.documents[id] = d
	s.createVersionLocked(d, actorID, "autosave")
	s.recordLocked(actorID, "document.saved", "document", id, nil)
	return d, nil
}
func (s *Service) SubmitReview(ctx context.Context, actorID, id, reviewerID string) (Review, error) {
	if err := checkContext(ctx); err != nil {
		return Review{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[id]
	if !ok {
		return Review{}, errors.New("document not found")
	}
	if d.AuthorID != actorID && !s.allowedLocked(actorID, id, "review") {
		return Review{}, errors.New("review permission required")
	}
	if reviewerID == "" {
		return Review{}, errors.New("reviewer is required")
	}
	d.Status = DocumentReviewing
	s.documents[id] = d
	r := Review{ID: s.id("review"), DocumentID: id, Version: d.Version, ReviewerID: reviewerID, State: "pending", CreatedAt: time.Now()}
	s.reviews[r.ID] = r
	s.notifyLocked(reviewerID, "review.requested", fmt.Sprintf("Document %s is waiting for your review", d.Title))
	s.recordLocked(actorID, "review.submitted", "document", id, map[string]string{"review_id": r.ID})
	return r, nil
}
func (s *Service) DecideReview(ctx context.Context, actorID, id, state, opinion string) (Review, error) {
	if err := checkContext(ctx); err != nil {
		return Review{}, err
	}
	if !review.ValidState(review.State(state)) {
		return Review{}, errors.New("invalid review state")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.reviews[id]
	if !ok {
		return Review{}, errors.New("review not found")
	}
	if r.ReviewerID != actorID {
		return Review{}, errors.New("only assigned reviewer may decide")
	}
	if r.State != "pending" {
		return Review{}, errors.New("review already decided")
	}
	r.State = rState(state)
	r.Opinion = opinion
	now := time.Now()
	r.DecidedAt = &now
	s.reviews[id] = r
	d := s.documents[r.DocumentID]
	if state == "approved" {
		d.Status = DocumentPublished
		d.PublishedVersion = r.Version
		d.UpdatedAt = now
		s.documents[d.ID] = d
		s.notifyLocked(d.AuthorID, "document.published", fmt.Sprintf("Document %s was published", d.Title))
	} else if state == "returned" {
		d.Status = DocumentDraft
		s.documents[d.ID] = d
		s.notifyLocked(d.AuthorID, "review.returned", fmt.Sprintf("Document %s needs changes", d.Title))
	} else {
		d.Status = DocumentDraft
		s.documents[d.ID] = d
		s.notifyLocked(d.AuthorID, "review.rejected", fmt.Sprintf("Document %s was rejected", d.Title))
	}
	s.recordLocked(actorID, "review.decided", "review", id, map[string]string{"state": state})
	return r, nil
}
func rState(v string) string { return v }

func (s *Service) ListVersions(ctx context.Context, documentID string) ([]Version, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Version{}
	for _, v := range s.versions {
		if v.DocumentID == documentID {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number > out[j].Number })
	return out, nil
}
func (s *Service) RestoreVersion(ctx context.Context, actorID, documentID, versionID string) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	v, ok := s.versions[versionID]
	if !ok || v.DocumentID != documentID {
		return Document{}, errors.New("version not found")
	}
	if !s.canEditLocked(actorID, d) {
		return Document{}, errors.New("document edit permission required")
	}
	d.Body = v.Body
	d.Version++
	d.Status = DocumentDraft
	d.UpdatedAt = time.Now()
	s.documents[documentID] = d
	s.createVersionLocked(d, actorID, fmt.Sprintf("restored version %d", v.Number))
	s.recordLocked(actorID, "document.version_restored", "document", documentID, map[string]string{"version_id": versionID})
	return d, nil
}
func (s *Service) createVersionLocked(d Document, authorID, summary string) {
	v := Version{ID: s.id("version"), DocumentID: d.ID, Number: d.Version, Body: d.Body, Summary: summary, AuthorID: authorID, CreatedAt: time.Now()}
	s.versions[v.ID] = v
}

func (s *Service) AddComment(ctx context.Context, actorID, documentID, parentID, body string) (Comment, error) {
	if err := checkContext(ctx); err != nil {
		return Comment{}, err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return Comment{}, errors.New("comment body is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.documents[documentID]; !ok {
		return Comment{}, errors.New("document not found")
	}
	if parentID != "" {
		p, ok := s.comments[parentID]
		if !ok || p.DocumentID != documentID {
			return Comment{}, errors.New("parent comment not found")
		}
	}
	c := Comment{ID: s.id("comment"), DocumentID: documentID, AuthorID: actorID, ParentID: parentID, Body: body, CreatedAt: time.Now()}
	s.comments[c.ID] = c
	d := s.documents[documentID]
	s.notifyLocked(d.AuthorID, "comment.created", fmt.Sprintf("New comment on %s", d.Title))
	s.recordLocked(actorID, "comment.created", "document", documentID, nil)
	return c, nil
}
func (s *Service) ListComments(ctx context.Context, documentID string) ([]Comment, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Comment{}
	for _, c := range s.comments {
		if c.DocumentID == documentID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (s *Service) ResolveComment(ctx context.Context, actorID, id string, resolved bool) (Comment, error) {
	if err := checkContext(ctx); err != nil {
		return Comment{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.comments[id]
	if !ok {
		return Comment{}, errors.New("comment not found")
	}
	if c.AuthorID != actorID && !s.allowedLocked(actorID, c.DocumentID, "admin") {
		return Comment{}, errors.New("comment permission required")
	}
	c.Resolved = resolved
	s.comments[id] = c
	return c, nil
}

func (s *Service) GrantPermission(ctx context.Context, actorID string, g PermissionGrant) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.allowedLocked(actorID, g.ResourceID, "admin") && actorID != s.ownerLocked(g.ResourceID) {
		return errors.New("admin permission required")
	}
	if g.SubjectID == "" || g.Permission == "" {
		return errors.New("subject and permission are required")
	}
	s.grants = append(s.grants, g)
	s.recordLocked(actorID, "permission.granted", "resource", g.ResourceID, nil)
	return nil
}
func (s *Service) Can(ctx context.Context, subject, resource, permission string) bool {
	if checkContext(ctx) != nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.allowedLocked(subject, resource, permission)
}
func (s *Service) allowedLocked(subject, resource, permission string) bool {
	allowed := false
	for _, g := range s.grants {
		if g.SubjectID == subject && (g.ResourceID == resource || g.ResourceID == "*") && g.Permission == permission {
			if g.ExplicitDeny {
				return false
			}
			allowed = true
		}
	}
	return allowed
}
func (s *Service) canEditLocked(actorID string, d Document) bool {
	return d.AuthorID == actorID || s.allowedLocked(actorID, d.ID, "edit") || s.allowedLocked(actorID, d.WorkspaceID, "edit")
}
func (s *Service) ownerLocked(resource string) string {
	if w, ok := s.workspaces[resource]; ok {
		return w.OwnerID
	}
	if d, ok := s.documents[resource]; ok {
		return d.AuthorID
	}
	return ""
}

func (s *Service) CreateTag(ctx context.Context, actorID, name, color string) (Tag, error) {
	if err := checkContext(ctx); err != nil {
		return Tag{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Tag{}, errors.New("tag name is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Tag{ID: s.id("tag"), Name: name, Color: color}
	s.tags[t.ID] = t
	return t, nil
}
func (s *Service) AttachTag(ctx context.Context, actorID, documentID, tagID string) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.documents[documentID]; !ok {
		return errors.New("document not found")
	}
	if _, ok := s.tags[tagID]; !ok {
		return errors.New("tag not found")
	}
	if s.documentTags[documentID] == nil {
		s.documentTags[documentID] = map[string]bool{}
	}
	s.documentTags[documentID][tagID] = true
	return nil
}
func (s *Service) ListTags(ctx context.Context, documentID string) ([]Tag, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Tag{}
	for id := range s.documentTags[documentID] {
		if t, ok := s.tags[id]; ok {
			out = append(out, t)
		}
	}
	return out, nil
}

func (s *Service) CreateShare(ctx context.Context, actorID, documentID, permission string, ttl time.Duration) (Share, error) {
	if err := checkContext(ctx); err != nil {
		return Share{}, err
	}
	if ttl <= 0 || ttl > 30*24*time.Hour {
		return Share{}, errors.New("share ttl must be between one second and thirty days")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return Share{}, errors.New("document not found")
	}
	if !s.canEditLocked(actorID, d) && !s.allowedLocked(actorID, documentID, "share") {
		return Share{}, errors.New("share permission required")
	}
	sh := Share{ID: s.id("share"), DocumentID: documentID, Token: s.id("token"), Permission: permission, ExpiresAt: time.Now().Add(ttl)}
	s.shares[sh.Token] = sh
	s.recordLocked(actorID, "share.created", "document", documentID, nil)
	return sh, nil
}
func (s *Service) ResolveShare(ctx context.Context, token string) (Share, Document, error) {
	if err := checkContext(ctx); err != nil {
		return Share{}, Document{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	sh, ok := s.shares[token]
	if !ok || time.Now().After(sh.ExpiresAt) {
		return Share{}, Document{}, errors.New("share is invalid or expired")
	}
	d, ok := s.documents[sh.DocumentID]
	if !ok || d.Status == DocumentDeleted {
		return Share{}, Document{}, errors.New("document not available")
	}
	return sh, d, nil
}

func (s *Service) Notify(ctx context.Context, userID, kind, message string) (Notification, error) {
	if err := checkContext(ctx); err != nil {
		return Notification{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.notifyLocked(userID, kind, message), nil
}
func (s *Service) notifyLocked(userID, kind, message string) Notification {
	n := Notification{ID: s.id("notification"), UserID: userID, Kind: kind, Message: message, CreatedAt: time.Now()}
	s.notifications[n.ID] = n
	return n
}
func (s *Service) ListNotifications(ctx context.Context, userID string, unreadOnly bool) ([]Notification, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Notification{}
	for _, n := range s.notifications {
		if n.UserID == userID && (!unreadOnly || !n.Read) {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (s *Service) MarkNotificationsRead(ctx context.Context, userID string) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, n := range s.notifications {
		if n.UserID == userID {
			n.Read = true
			s.notifications[id] = n
		}
	}
	return nil
}

func (s *Service) Search(ctx context.Context, subjectID, query string) ([]SearchResult, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return []SearchResult{}, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	q := strings.ToLower(query)
	out := []SearchResult{}
	for _, d := range s.documents {
		if d.Status == DocumentDeleted {
			continue
		}
		hay := strings.ToLower(d.Title + " " + d.Summary + " " + d.Body)
		idx := strings.Index(hay, q)
		if idx < 0 {
			continue
		}
		start := idx - 40
		if start < 0 {
			start = 0
		}
		end := idx + len(query) + 80
		if end > len(d.Body) {
			end = len(d.Body)
		}
		out = append(out, SearchResult{Document: d, Highlight: d.Body[start:end]})
	}
	return out, nil
}
func (s *Service) RecordRead(ctx context.Context, userID, documentID string) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	recent := s.recent[userID]
	for _, id := range recent {
		if id == documentID {
			return nil
		}
	}
	s.recent[userID] = append([]string{documentID}, recent...)
	if len(s.recent[userID]) > 20 {
		s.recent[userID] = s.recent[userID][:20]
	}
	return nil
}
func (s *Service) RecentDocuments(ctx context.Context, userID string) ([]Document, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Document{}
	for _, id := range s.recent[userID] {
		if d, ok := s.documents[id]; ok {
			out = append(out, d)
		}
	}
	return out, nil
}

func (s *Service) AuditLogs(ctx context.Context, actorID, action string) ([]AuditLog, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []AuditLog{}
	for _, l := range s.audit {
		if actorID != "" && l.ActorID != actorID {
			continue
		}
		if action != "" && l.Action != action {
			continue
		}
		out = append(out, l)
	}
	return out, nil
}
func (s *Service) recordLocked(actor, action, objectType, objectID string, metadata map[string]string) {
	s.audit = append(s.audit, AuditLog{ID: s.id("audit"), ActorID: actor, Action: action, ObjectType: objectType, ObjectID: objectID, Metadata: metadata, CreatedAt: time.Now()})
}
func (s *Service) nextPositionLockedUnused() int { return len(s.documents) }
