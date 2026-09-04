package project

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zscaler/migration-platform/backend/internal/auth"
	"github.com/zscaler/migration-platform/backend/internal/file"
	"github.com/zscaler/migration-platform/backend/internal/httperr"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type projectResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Files     []fileResp `json:"files,omitempty"`
}

type fileResp struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type listResponse struct {
	Items []projectResponse `json:"items"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	projects, err := h.service.List(r.Context(), userID)
	if err != nil {
		httperr.Internal(w)
		return
	}
	items := make([]projectResponse, 0, len(projects))
	for _, p := range projects {
		items = append(items, toProjectResponse(p, nil))
	}
	writeJSON(w, http.StatusOK, listResponse{Items: items})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httperr.BadRequest(w, "Failed to parse multipart form")
		return
	}

	maxSize := h.service.MaxUploadSize()
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	f, header, err := r.FormFile("file")
	if err != nil {
		httperr.BadRequest(w, "File is required")
		return
	}
	defer f.Close()

	title := r.FormValue("title")

	p, fileRecord, err := h.service.Create(r.Context(), CreateInput{
		UserID:      userID,
		Title:       title,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      f,
	})
	if err != nil {
		translateFileError(w, err)
		return
	}

	resp := toProjectResponse(p, []file.File{fileRecord})
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "projectID")
	if id == "" {
		httperr.BadRequest(w, "Missing project id")
		return
	}

	detail, err := h.service.Get(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httperr.NotFound(w)
			return
		}
		httperr.Internal(w)
		return
	}

	files := make([]fileResp, 0, len(detail.Files))
	for _, fl := range detail.Files {
		files = append(files, toFileResp(fl))
	}

	writeJSON(w, http.StatusOK, projectResponse{
		ID:        detail.Project.ID,
		Title:     detail.Project.Title,
		CreatedAt: detail.Project.CreatedAt.Format(timeRFC),
		UpdatedAt: detail.Project.UpdatedAt.Format(timeRFC),
		Files:     files,
	})
}

type renameRequest struct {
	Title string `json:"title"`
}

func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "projectID")

	var req renameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.BadRequest(w, "Invalid request body")
		return
	}

	p, err := h.service.Rename(r.Context(), userID, id, req.Title)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httperr.NotFound(w)
			return
		}
		httperr.BadRequest(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toProjectResponse(p, nil))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "projectID")

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httperr.NotFound(w)
			return
		}
		httperr.Internal(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func translateFileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, file.ErrTooLarge):
		httperr.FileTooLarge(w)
	case errors.Is(err, file.ErrUnsupported):
		httperr.UnsupportedFileType(w)
	case errors.Is(err, file.ErrEmpty):
		httperr.BadRequest(w, "File is empty")
	default:
		httperr.Internal(w)
	}
}

func toProjectResponse(p Project, files []file.File) projectResponse {
	var fr []fileResp
	if len(files) > 0 {
		fr = make([]fileResp, 0, len(files))
		for _, f := range files {
			fr = append(fr, toFileResp(f))
		}
	}
	return projectResponse{
		ID:        p.ID,
		Title:     p.Title,
		CreatedAt: p.CreatedAt.Format(timeRFC),
		UpdatedAt: p.UpdatedAt.Format(timeRFC),
		Files:     fr,
	}
}

func toFileResp(f file.File) fileResp {
	return fileResp{
		ID:          f.ID,
		Filename:    f.Filename,
		ContentType: f.ContentType,
		Size:        f.Size,
	}
}

const timeRFC = "2006-01-02T15:04:05Z07:00"

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
