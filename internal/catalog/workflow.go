package catalog

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (s *Service) SetDocumentFlag(ctx context.Context, actorID, documentID string, favorite, pinned *bool) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	if !s.canEditLocked(actorID, d) {
		return Document{}, errors.New("document edit permission required")
	}
	if favorite != nil {
		d.Favorite = *favorite
	}
	if pinned != nil {
		d.Pinned = *pinned
	}
	d.UpdatedAt = time.Now()
	s.documents[documentID] = d
	s.recordLocked(actorID, "document.flag.updated", "document", documentID, nil)
	return d, nil
}
func (s *Service) ChangeDocumentStatus(ctx context.Context, actorID, documentID string, status DocumentStatus) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	if !s.canEditLocked(actorID, d) {
		return Document{}, errors.New("document edit permission required")
	}
	if status != DocumentArchived && status != DocumentPublished && status != DocumentDraft {
		return Document{}, errors.New("invalid document status")
	}
	if status == DocumentPublished {
		d.PublishedVersion = d.Version
	}
	d.Status = status
	d.UpdatedAt = time.Now()
	s.documents[documentID] = d
	s.recordLocked(actorID, "document.status.updated", "document", documentID, map[string]string{"status": string(status)})
	return d, nil
}
func (s *Service) MoveDocument(ctx context.Context, actorID, documentID, folderID string) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	if !s.canEditLocked(actorID, d) {
		return Document{}, errors.New("document edit permission required")
	}
	if folderID != "" {
		f, ok := s.folders[folderID]
		if !ok || f.WorkspaceID != d.WorkspaceID {
			return Document{}, errors.New("folder not found")
		}
	}
	d.FolderID = folderID
	d.UpdatedAt = time.Now()
	s.documents[documentID] = d
	s.recordLocked(actorID, "document.moved", "document", documentID, map[string]string{"folder_id": folderID})
	return d, nil
}
func (s *Service) CopyDocument(ctx context.Context, actorID, documentID, folderID string) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	source, ok := s.documents[documentID]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	if !s.canEditLocked(actorID, source) {
		return Document{}, errors.New("document edit permission required")
	}
	if folderID != "" {
		f, ok := s.folders[folderID]
		if !ok || f.WorkspaceID != source.WorkspaceID {
			return Document{}, errors.New("folder not found")
		}
	}
	now := time.Now()
	copy := source
	copy.ID = s.id("doc")
	copy.FolderID = folderID
	copy.Title = strings.TrimSpace(source.Title) + " (copy)"
	copy.AuthorID = actorID
	copy.Status = DocumentDraft
	copy.Version = 1
	copy.PublishedVersion = 0
	copy.Favorite = false
	copy.Pinned = false
	copy.CreatedAt = now
	copy.UpdatedAt = now
	s.documents[copy.ID] = copy
	s.createVersionLocked(copy, actorID, "copied from "+source.ID)
	s.recordLocked(actorID, "document.copied", "document", copy.ID, map[string]string{"source_id": source.ID})
	return copy, nil
}
func (s *Service) DeleteDocument(ctx context.Context, actorID, documentID string) (RecycleItem, error) {
	if err := checkContext(ctx); err != nil {
		return RecycleItem{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return RecycleItem{}, errors.New("document not found")
	}
	if !s.canEditLocked(actorID, d) {
		return RecycleItem{}, errors.New("document delete permission required")
	}
	d.Status = DocumentDeleted
	d.UpdatedAt = time.Now()
	s.documents[documentID] = d
	item := RecycleItem{ID: s.id("recycle"), ObjectType: "document", ObjectID: d.ID, Name: d.Title, DeletedBy: actorID, DeletedAt: time.Now()}
	s.recycle[item.ID] = item
	s.recordLocked(actorID, "document.deleted", "document", documentID, nil)
	return item, nil
}
func (s *Service) ListRecycle(ctx context.Context, actorID string) ([]RecycleItem, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []RecycleItem{}
	for _, item := range s.recycle {
		if item.DeletedBy == actorID || s.allowedLocked(actorID, item.ObjectID, "admin") {
			out = append(out, item)
		}
	}
	return out, nil
}
func (s *Service) RestoreRecycle(ctx context.Context, actorID, recycleID string) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.recycle[recycleID]
	if !ok {
		return Document{}, errors.New("recycle item not found")
	}
	d, ok := s.documents[item.ObjectID]
	if !ok {
		return Document{}, errors.New("document not found")
	}
	if item.DeletedBy != actorID && !s.allowedLocked(actorID, d.ID, "admin") {
		return Document{}, errors.New("recycle restore permission required")
	}
	d.Status = DocumentDraft
	d.UpdatedAt = time.Now()
	s.documents[d.ID] = d
	delete(s.recycle, recycleID)
	s.recordLocked(actorID, "document.restored", "document", d.ID, nil)
	return d, nil
}
func (s *Service) AddAnnotation(ctx context.Context, actorID, documentID string, start, end int, quote, body string) (Annotation, error) {
	if err := checkContext(ctx); err != nil {
		return Annotation{}, err
	}
	if start < 0 || end < start {
		return Annotation{}, errors.New("invalid annotation range")
	}
	if strings.TrimSpace(body) == "" {
		return Annotation{}, errors.New("annotation comment is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return Annotation{}, errors.New("document not found")
	}
	if !s.allowedLocked(actorID, documentID, "comment") && d.AuthorID != actorID {
		return Annotation{}, errors.New("document comment permission required")
	}
	if end > len([]rune(d.Body)) {
		return Annotation{}, errors.New("annotation range exceeds document")
	}
	a := Annotation{ID: s.id("annotation"), DocumentID: documentID, AuthorID: actorID, Start: start, End: end, Quote: quote, Comment: body, CreatedAt: time.Now()}
	s.annotations[a.ID] = a
	s.recordLocked(actorID, "annotation.created", "annotation", a.ID, nil)
	return a, nil
}
func (s *Service) ListAnnotations(ctx context.Context, documentID string) ([]Annotation, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Annotation{}
	for _, a := range s.annotations {
		if a.DocumentID == documentID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (s *Service) CreateAttachment(ctx context.Context, actorID, documentID, name, contentType string, size int64) (Attachment, error) {
	if err := checkContext(ctx); err != nil {
		return Attachment{}, err
	}
	if strings.TrimSpace(name) == "" || size < 0 {
		return Attachment{}, errors.New("invalid attachment")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.documents[documentID]
	if !ok {
		return Attachment{}, errors.New("document not found")
	}
	if !s.allowedLocked(actorID, documentID, "edit") && d.AuthorID != actorID {
		return Attachment{}, errors.New("attachment permission required")
	}
	a := Attachment{ID: s.id("attachment"), DocumentID: documentID, Name: name, ContentType: contentType, Size: size, CreatedAt: time.Now()}
	s.attachments[a.ID] = a
	return a, nil
}
func (s *Service) CompleteAttachment(ctx context.Context, actorID, id string) (Attachment, error) {
	if err := checkContext(ctx); err != nil {
		return Attachment{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.attachments[id]
	if !ok {
		return Attachment{}, errors.New("attachment not found")
	}
	d := s.documents[a.DocumentID]
	if !s.allowedLocked(actorID, a.DocumentID, "edit") && d.AuthorID != actorID {
		return Attachment{}, errors.New("attachment permission required")
	}
	a.Completed = true
	s.attachments[id] = a
	s.recordLocked(actorID, "attachment.completed", "attachment", id, nil)
	return a, nil
}
func (s *Service) ListAttachments(ctx context.Context, documentID string) ([]Attachment, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Attachment{}
	for _, a := range s.attachments {
		if a.DocumentID == documentID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (s *Service) Report(ctx context.Context, workspaceID string) (Report, error) {
	if err := checkContext(ctx); err != nil {
		return Report{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	r := Report{WorkspaceID: workspaceID}
	users := map[string]bool{}
	for _, d := range s.documents {
		if d.WorkspaceID != workspaceID {
			continue
		}
		if d.Status != DocumentDeleted {
			r.Documents++
		}
		if d.Status == DocumentPublished {
			r.Published++
		}
		users[d.AuthorID] = true
	}
	for _, c := range s.comments {
		if d, ok := s.documents[c.DocumentID]; ok && d.WorkspaceID == workspaceID {
			r.Comments++
		}
	}
	for _, a := range s.attachments {
		if d, ok := s.documents[a.DocumentID]; ok && d.WorkspaceID == workspaceID {
			r.Attachments++
		}
	}
	for _, logs := range s.recent {
		for _, id := range logs {
			if d, ok := s.documents[id]; ok && d.WorkspaceID == workspaceID {
				r.Reads++
			}
		}
	}
	r.ActiveUsers = len(users)
	for _, l := range s.audit {
		if d, ok := s.documents[l.ObjectID]; ok && d.WorkspaceID == workspaceID && strings.Contains(l.Action, "updated") {
			r.Edits++
		}
	}
	return r, nil
}
