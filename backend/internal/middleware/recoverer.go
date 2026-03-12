package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer is a middleware that recovers from panics, logs the error,
// and returns a 500 Internal Server Error response.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID := RequestIDFromContext(r.Context())
				slog.Error("panic recovered",
					"error", rec,
					"request_id", reqID,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":{"code":"INTERNAL_ERROR","message":"An unexpected error occurred"}}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
