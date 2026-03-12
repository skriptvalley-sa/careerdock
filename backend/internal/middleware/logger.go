package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// WriteHeader captures the status code and delegates to the wrapped writer.
func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// Logger returns a middleware that logs each request with structured fields.
// Uses slog for structured, JSON-compatible logging.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			duration := time.Since(start)
			reqID := RequestIDFromContext(r.Context())

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Duration("duration", duration),
				slog.Int("bytes", sw.bytes),
				slog.String("remote_addr", r.RemoteAddr),
			}

			if reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}

			if userID := UserIDFromContext(r.Context()); userID.String() != "00000000-0000-0000-0000-000000000000" {
				attrs = append(attrs, slog.String("user_id", userID.String()))
			}

			level := slog.LevelInfo
			if sw.status >= 500 {
				level = slog.LevelError
			} else if sw.status >= 400 {
				level = slog.LevelWarn
			}

			logger.LogAttrs(r.Context(), level, "http request",
				attrs...,
			)
		})
	}
}
