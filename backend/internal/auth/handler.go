package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zscaler/migration-platform/backend/internal/httperr"
	"github.com/zscaler/migration-platform/backend/internal/user"
)

type Handler struct {
	service *Service
	users   *user.Repository
}

func NewHandler(service *Service, users *user.Repository) *Handler {
	return &Handler{service: service, users: users}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type loginResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
	User        userResponse `json:"user"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.BadRequest(w, "Invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		httperr.BadRequest(w, "Username and password are required")
		return
	}

	result, err := h.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httperr.Write(w, http.StatusUnauthorized, "unauthorized", "Invalid username or password")
			return
		}
		httperr.Internal(w)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
		User:        userResponse{ID: result.User.ID, Username: result.User.Username},
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	u, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		httperr.NotFound(w)
		return
	}
	writeJSON(w, http.StatusOK, userResponse{ID: u.ID, Username: u.Username})
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.service.devRegister {
		httperr.NotFound(w)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.BadRequest(w, "Invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		httperr.BadRequest(w, "Username and password are required")
		return
	}

	u, err := h.service.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameTaken):
			httperr.Conflict(w, "Username already taken")
		case errors.Is(err, ErrInvalidCredentials):
			httperr.BadRequest(w, err.Error())
		default:
			httperr.Internal(w)
		}
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{ID: u.ID, Username: u.Username})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
