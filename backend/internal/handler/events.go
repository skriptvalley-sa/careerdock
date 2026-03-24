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

// Events handles GET /api/events.
// Opens a persistent SSE connection. Subscribes to the user's Redis pub/sub
// channel and forwards events as they arrive. Sends heartbeats every 30s.
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

	// Subscribe to user's Redis channel
	channel := sseChannelForUser(userID)
	sub := h.redis.Subscribe(r.Context(), channel)
	defer func() { _ = sub.Close() }()

	msgCh := sub.Channel()
	heartbeat := time.NewTicker(30 * time.Second)
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
			_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
