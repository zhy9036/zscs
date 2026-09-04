package message

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zscaler/migration-platform/backend/internal/auth"
	"github.com/zscaler/migration-platform/backend/internal/httperr"
	"github.com/zscaler/migration-platform/backend/internal/project"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type messageResponse struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type listResponse struct {
	Items []messageResponse `json:"items"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")

	msgs, err := h.service.List(r.Context(), userID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrNotFound) {
			httperr.NotFound(w)
			return
		}
		httperr.Internal(w)
		return
	}

	items := make([]messageResponse, 0, len(msgs))
	for _, m := range msgs {
		items = append(items, toMessageResponse(m))
	}
	writeJSON(w, http.StatusOK, listResponse{Items: items})
}

type sendRequest struct {
	Content string `json:"content"`
}

type sendResponse struct {
	UserMessage      messageResponse `json:"user_message"`
	AssistantMessage *messageResponse `json:"assistant_message"`
	Status           string           `json:"status"`
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	projectID := chi.URLParam(r, "projectID")

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.BadRequest(w, "Invalid request body")
		return
	}

	result, err := h.service.Send(r.Context(), userID, projectID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, project.ErrNotFound):
			httperr.NotFound(w)
		case strings.Contains(err.Error(), "must not be empty"):
			httperr.BadRequest(w, err.Error())
		default:
			httperr.Internal(w)
		}
		return
	}

	resp := sendResponse{
		UserMessage: toMessageResponse(result.UserMessage),
		Status:      result.Status,
	}
	if result.AssistantMessage != nil {
		asst := toMessageResponse(*result.AssistantMessage)
		resp.AssistantMessage = &asst
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	httperr.NotImplemented(w, "Streaming is not implemented yet")
}

func toMessageResponse(m Message) messageResponse {
	return messageResponse{
		ID:        m.ID,
		Role:      m.Role,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
