package handler

import (
	"net/http"
	"time"

	"github.com/skriptvalley/careerdock/internal/service"
)

// Health returns a liveness probe handler.
// GET /api/health — always returns 200 if the process is running.
func Health(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, DataResponse{
			Data: map[string]any{
				"status":    "ok",
				"version":   svc.Version,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}
}

// HealthReady returns a readiness probe handler.
// GET /api/health/ready — checks database and Redis connectivity.
func HealthReady(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		checks := make(map[string]string)

		// Check database
		if err := svc.PingDB(ctx); err != nil {
			checks["database"] = "error"
		} else {
			checks["database"] = "ok"
		}

		// Check Redis
		if err := svc.PingRedis(ctx); err != nil {
			checks["redis"] = "error"
		} else {
			checks["redis"] = "ok"
		}

		// Determine overall status
		status := http.StatusOK
		statusText := "ready"
		for _, v := range checks {
			if v != "ok" {
				status = http.StatusServiceUnavailable
				statusText = "degraded"
				break
			}
		}

		respondJSON(w, status, DataResponse{
			Data: map[string]any{
				"status": statusText,
				"checks": checks,
			},
		})
	}
}
