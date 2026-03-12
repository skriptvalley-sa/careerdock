package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

type reqIDKey struct{}

// RequestID is a middleware that reads or generates a request correlation ID.
// The ID is added to the response headers and stored in the request context
// for use by the logger and downstream handlers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(requestIDHeader)
		if reqID == "" {
			reqID = uuid.Must(uuid.NewV7()).String()
		}

		// Set on response header
		w.Header().Set(requestIDHeader, reqID)

		// Store in context
		ctx := context.WithValue(r.Context(), reqIDKey{}, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext extracts the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(reqIDKey{}).(string); ok {
		return id
	}
	return ""
}
