package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/skriptvalley/careerdock/internal/middleware"
)

// SSEHandler handles Server-Sent Events for real-time notifications.
// This is a skeleton — no events are emitted yet (Sprint 3+ will push
// job completion events for resume parsing, ATS scoring, etc.).
type SSEHandler struct{}

// NewSSEHandler creates a new SSEHandler.
func NewSSEHandler() *SSEHandler {
	return &SSEHandler{}
}

// Events handles GET /api/events.
// Opens a persistent SSE connection. Currently just keeps it open
// with periodic heartbeats until the client disconnects.
func (h *SSEHandler) Events(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial connection event
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"user_id\":\"%s\"}\n\n", userID)
	flusher.Flush()

	slog.Info("SSE client connected", "user_id", userID)

	// Keep connection open until client disconnects
	<-r.Context().Done()

	slog.Info("SSE client disconnected", "user_id", userID)
}
