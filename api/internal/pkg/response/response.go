// Package response provides standardized JSON response helpers for HTTP handlers.
package response

import (
	"encoding/json"
	"net/http"
)

// envelope wraps all API responses in a consistent shape.
type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *apiError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type apiError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta carries pagination or other response metadata.
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func write(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// JSON sends a 200 OK with data payload.
func JSON(w http.ResponseWriter, data interface{}) {
	write(w, http.StatusOK, envelope{Success: true, Data: data})
}

// Created sends a 201 Created with data payload.
func Created(w http.ResponseWriter, data interface{}) {
	write(w, http.StatusCreated, envelope{Success: true, Data: data})
}

// NoContent sends a 204 No Content.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WithMeta sends a 200 OK with data and pagination metadata.
func WithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	write(w, http.StatusOK, envelope{Success: true, Data: data, Meta: meta})
}

// Error sends an error response with the given status code.
func Error(w http.ResponseWriter, status int, code, message string, details ...interface{}) {
	e := &apiError{Code: code, Message: message}
	if len(details) > 0 {
		e.Details = details[0]
	}
	write(w, status, envelope{Success: false, Error: e})
}

// BadRequest sends a 400 Bad Request.
func BadRequest(w http.ResponseWriter, message string, details ...interface{}) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message, details...)
}

// Unauthorized sends a 401 Unauthorized.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 Not Found.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 Conflict.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// UnprocessableEntity sends a 422 with validation error details.
func UnprocessableEntity(w http.ResponseWriter, details interface{}) {
	Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "validation failed", details)
}

// TooManyRequests sends a 429 Too Many Requests.
func TooManyRequests(w http.ResponseWriter) {
	Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later")
}

// InternalError sends a 500 Internal Server Error.
// Details are intentionally omitted to avoid leaking internals.
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
}

// ServiceUnavailable sends a 503.
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}
