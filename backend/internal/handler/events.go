package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/skriptvalley/careerdock/internal/middleware"
)

// SSEEvent represents a server-sent event payload published via Redis pub/sub.
type SSEEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// SSEHandler handles Server-Sent Events for real-time notifications.
// Uses Redis pub/sub: the worker publishes events to user-specific channels,
// and this handler subscribes and forwards them to connected clients.
type SSEHandler struct {
	redis *redis.Client
}

// NewSSEHandler creates a new SSEHandler backed by Redis pub/sub.
func NewSSEHandler(redisClient *redis.Client) *SSEHandler {
	return &SSEHandler{redis: redisClient}
}

// sseChannelForUser returns the Redis pub/sub channel name for a user.
func sseChannelForUser(userID uuid.UUID) string {
	return fmt.Sprintf("sse:user:%s", userID)
}

// PublishSSEEvent publishes an event to a user's SSE channel via Redis.
// Called from worker tasks or services to push real-time updates.
func PublishSSEEvent(redisClient *redis.Client, userID uuid.UUID, eventType string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal SSE event data: %w", err)
	}

	event := SSEEvent{
		Type: eventType,
		Data: payload,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal SSE event: %w", err)
	}

	channel := sseChannelForUser(userID)
	return redisClient.Publish(context.Background(), channel, eventJSON).Err()
}

// unwrapper is implemented by middleware ResponseWriter wrappers (like Chi's)
// that allow access to the underlying http.ResponseWriter.
type unwrapper interface {
	Unwrap() http.ResponseWriter
}

// getFlusher extracts http.Flusher from a ResponseWriter, unwrapping
// middleware wrappers as needed.
func getFlusher(w http.ResponseWriter) (http.Flusher, bool) {
	for {
		if f, ok := w.(http.Flusher); ok {
			return f, true
		}
		if u, ok := w.(unwrapper); ok {
			w = u.Unwrap()
		} else {
			return nil, false
		}
	}
}

// Events handles GET /api/events.
// Opens a persistent SSE connection. Subscribes to the user's Redis pub/sub
// channel and forwards events as they arrive. Sends heartbeats every 30s.
func (h *SSEHandler) Events(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	// Get flusher by unwrapping middleware wrappers
	flusher, ok := getFlusher(w)
	if !ok {
		slog.Error("SSE: ResponseWriter does not support Flusher")
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Send initial connection event
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"user_id\":\"%s\"}\n\n", userID)
	flusher.Flush()

	slog.Info("SSE client connected", "user_id", userID)

	// Subscribe to user's Redis channel
	channel := sseChannelForUser(userID)
	sub := h.redis.Subscribe(r.Context(), channel)
	defer func() { _ = sub.Close() }()

	msgCh := sub.Channel()
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			slog.Info("SSE client disconnected", "user_id", userID)
			return

		case msg, ok := <-msgCh:
			if !ok {
				return // channel closed
			}

			var event SSEEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				slog.Warn("invalid SSE event from Redis", "error", err)
				continue
			}

			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(event.Data))
			flusher.Flush()

		case <-heartbeat.C:
			// Heartbeat keeps the connection alive and resets the WriteTimeout.
			_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
