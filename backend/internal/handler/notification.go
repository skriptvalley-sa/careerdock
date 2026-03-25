package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// NotificationHandler handles notification endpoints.
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// ListNotifications returns recent notifications for the authenticated user.
// GET /api/notifications?limit=20
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	notifications, err := h.svc.List(r.Context(), userID, limit)
	if err != nil {
		respondError(w, r, err)
		return
	}

	// Return empty array instead of null for JSON consistency
	if notifications == nil {
		respondJSON(w, http.StatusOK, DataResponse{Data: []struct{}{}})
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: notifications})
}

// MarkRead marks a notification as read.
// PUT /api/notifications/{id}/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorBody{Code: "VALIDATION_ERROR", Message: "invalid notification ID"},
		})
		return
	}

	if err := h.svc.MarkRead(r.Context(), id); err != nil {
		respondError(w, r, err)
		return
	}

	respondMessage(w, http.StatusOK, "notification marked as read")
}

// UnreadCount returns the count of unread notifications.
// GET /api/notifications/unread-count
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	count, err := h.svc.CountUnread(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{
		Data: map[string]int{"count": count},
	})
}
