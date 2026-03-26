// Package handler provides HTTP handlers for the CareerDock API.
// Handlers are thin — validate input, call service, respond.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
)

// --- Response envelope types (matching api.md §1.4) ---

// DataResponse wraps a single resource.
type DataResponse struct {
	Data any `json:"data"`
}

// PaginatedResponse wraps a list of resources with cursor pagination.
type PaginatedResponse struct {
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination holds cursor-based pagination metadata.
type Pagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// ErrorResponse wraps an error for the client.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the structured error envelope.
type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// --- Response helpers ---

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// respondError maps an error to the appropriate HTTP status code and
// writes a structured error response.
func respondError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		status := mapErrorCodeToHTTP(appErr.Code)

		// Log internal details for server errors
		if status >= 500 {
			reqID := middleware.RequestIDFromContext(r.Context())
			slog.Error("internal error",
				"code", appErr.Code,
				"message", appErr.Message,
				"error", appErr.Err,
				"request_id", reqID,
			)
		}

		respondJSON(w, status, ErrorResponse{
			Error: ErrorBody{
				Code:    string(appErr.Code),
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}

	// Unexpected error — don't leak internals
	reqID := middleware.RequestIDFromContext(r.Context())
	slog.Error("unexpected error",
		"error", err,
		"request_id", reqID,
	)
	respondJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: ErrorBody{
			Code:    string(domain.ErrCodeInternal),
			Message: "An unexpected error occurred",
		},
	})
}

// respondMessage writes a simple success message response.
func respondMessage(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, DataResponse{
		Data: map[string]string{"message": message},
	})
}

// mapErrorCodeToHTTP converts a domain error code to an HTTP status code.
func mapErrorCodeToHTTP(code domain.ErrorCode) int {
	switch code {
	case domain.ErrCodeNotFound:
		return http.StatusNotFound
	case domain.ErrCodeConflict:
		return http.StatusConflict
	case domain.ErrCodeValidation:
		return http.StatusUnprocessableEntity
	case domain.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case domain.ErrCodeForbidden:
		return http.StatusForbidden
	case domain.ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case domain.ErrCodeInsufficientCredits:
		return http.StatusPaymentRequired
	case domain.ErrCodePaymentFailed:
		return http.StatusBadGateway
	case domain.ErrCodeAIUnavailable:
		return http.StatusServiceUnavailable
	case domain.ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
