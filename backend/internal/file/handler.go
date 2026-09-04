package file

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zscaler/migration-platform/backend/internal/httperr"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	fileID := chi.URLParam(r, "fileID")
	if projectID == "" || fileID == "" {
		httperr.BadRequest(w, "Missing project or file id")
		return
	}

	f, rc, err := h.service.Get(r.Context(), projectID, fileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httperr.NotFound(w)
			return
		}
		httperr.Internal(w)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", f.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+f.Filename+"\"")
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, rc); err != nil {
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	fileID := chi.URLParam(r, "fileID")
	if projectID == "" || fileID == "" {
		httperr.BadRequest(w, "Missing project or file id")
		return
	}

	if err := h.service.Delete(r.Context(), projectID, fileID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httperr.NotFound(w)
			return
		}
		httperr.Internal(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
