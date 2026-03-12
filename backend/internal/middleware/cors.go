package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
// Configuration follows SECURITY.md §3 — no wildcard origins, credentials
// required for httpOnly cookie auth.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,  // Required for httpOnly cookie auth
		MaxAge:           86400, // Preflight cache: 24 hours
	})
}
