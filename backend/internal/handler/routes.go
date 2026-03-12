package handler

import (
	"github.com/go-chi/chi/v5"

	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// MountRoutes wires all HTTP routes onto the given router.
// This is the single place that maps URLs → handlers.
func MountRoutes(r chi.Router, svc *service.Services, auth *middleware.Auth) {
	r.Route("/api", func(r chi.Router) {
		// --- Health ---
		r.Get("/health", Health(svc))
		r.Get("/health/ready", HealthReady(svc))

		// --- Auth (public, no auth middleware) ---
		r.Route("/auth", func(r chi.Router) {
			authH := NewAuthHandler(svc.Auth, svc.IsProduction)
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)
			r.Post("/logout", authH.Logout)
			r.Post("/verify-email", authH.VerifyEmail)
			r.Post("/forgot-password", authH.ForgotPassword)
			r.Post("/reset-password", authH.ResetPassword)
		})

		// --- Authenticated routes ---
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuthenticated)

			// Auth - me
			r.Get("/auth/me", NewAuthHandler(svc.Auth, svc.IsProduction).Me)

			// TODO (Sprint 1): Company routes (authenticated write ops)
			// TODO (Sprint 2): List routes, User settings, SSE
			// TODO (Sprint 3): Payment routes, Resume routes
			// TODO (Sprint 4): ATS routes, Curated list routes

			// --- Premium routes ---
			r.Group(func(r chi.Router) {
				r.Use(auth.RequirePremium)
				// TODO (Sprint 3): Resume upload, ATS checks
				// TODO (Sprint 4): Curated list generation
			})

			// --- Admin routes ---
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireAdmin)
				// TODO (Sprint 5): Admin endpoints
			})
		})

		// --- Public routes (no auth) ---
		// TODO (Sprint 1): Company directory (public read)

		// --- Webhooks ---
		// TODO (Sprint 3): Razorpay webhook
	})
}
