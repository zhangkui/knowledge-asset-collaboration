package shared

import (
	"net/http"

	"github.com/zhangkui/knowledge-asset-collaboration/internal/attachment"
	"strconv"

	"github.com/zhangkui/knowledge-asset-collaboration/internal/catalog"
)

func (a *App) recycle(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method == http.MethodGet {
		items, err := a.Catalog.ListRecycle(r.Context(), user)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
		return
	}
	if r.Method == http.MethodPost && id != "" {
		item, err := a.Catalog.RestoreRecycle(r.Context(), user, id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
func (a *App) reports(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "method not allowed")
		return
	}
	report, err := a.Catalog.Report(r.Context(), r.URL.Query().Get("workspace_id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, 200, report)
}
func (a *App) attachments(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method == http.MethodGet {
		items, err := a.Catalog.ListAttachments(r.Context(), id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
		return
	}
	if r.Method == http.MethodPost {
		var in struct {
			DocumentID, Name, ContentType string
			Size                          int64
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		item, err := a.Catalog.CreateAttachment(r.Context(), user, in.DocumentID, in.Name, in.ContentType, in.Size)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, item)
		return
	}
	if r.Method == http.MethodPatch && id != "" {
		item, err := a.Catalog.CompleteAttachment(r.Context(), user, id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 200, item)
		return
	}
	writeError(w, 405, "method not allowed")
}
func (a *App) annotations(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method == http.MethodGet {
		items, err := a.Catalog.ListAnnotations(r.Context(), id)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writePage(w, items, len(items))
		return
	}
	if r.Method == http.MethodPost {
		var in struct {
			DocumentID, Quote, Comment string
			Start, End                 int
		}
		if err := decode(r, &in); err != nil {
			writeError(w, 400, "invalid request")
			return
		}
		item, err := a.Catalog.AddAnnotation(r.Context(), user, in.DocumentID, in.Start, in.End, in.Quote, in.Comment)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		writeJSON(w, 201, item)
		return
	}
	writeError(w, 405, "method not allowed")
}

var _ = strconv.IntSize
var _ catalog.DocumentStatus

func (a *App) uploadState(w http.ResponseWriter, r *http.Request, user, id string) {
	if r.Method != http.MethodGet || id == "" {
		writeError(w, http.StatusBadRequest, "upload id is required")
		return
	}
	item, err := a.Uploads.Get(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

var _ attachment.Chunk
