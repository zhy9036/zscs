package httperr

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Error Error `json:"error"`
}

func Write(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Error: Error{Code: code, Message: message},
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	Write(w, http.StatusBadRequest, "invalid_request", message)
}

func Unauthorized(w http.ResponseWriter) {
	Write(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
}

func Forbidden(w http.ResponseWriter) {
	Write(w, http.StatusForbidden, "forbidden", "Access denied")
}

func NotFound(w http.ResponseWriter) {
	Write(w, http.StatusNotFound, "not_found", "Resource not found")
}

func Conflict(w http.ResponseWriter, message string) {
	Write(w, http.StatusConflict, "conflict", message)
}

func FileTooLarge(w http.ResponseWriter) {
	Write(w, http.StatusRequestEntityTooLarge, "file_too_large", "File exceeds maximum allowed size")
}

func UnsupportedFileType(w http.ResponseWriter) {
	Write(w, http.StatusUnsupportedMediaType, "unsupported_file_type", "File type is not supported")
}

func Internal(w http.ResponseWriter) {
	Write(w, http.StatusInternalServerError, "internal_error", "Internal server error")
}

func NotImplemented(w http.ResponseWriter, message string) {
	Write(w, http.StatusNotImplemented, "not_implemented", message)
}
