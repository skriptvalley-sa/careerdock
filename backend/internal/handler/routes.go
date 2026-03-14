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

			// --- Lists (Sprint 2) ---
			listH := NewListHandler(svc.List, svc.Company)
			r.Route("/lists", func(r chi.Router) {
				r.Get("/", listH.ListLists)
				r.Post("/", listH.CreateList)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", listH.GetList)
					r.Put("/", listH.UpdateList)
					r.Delete("/", listH.DeleteList)

					r.Route("/entries", func(r chi.Router) {
						r.Post("/", listH.CreateEntry)
						r.Post("/batch", listH.BatchCreateEntries)
						r.Put("/sync", listH.SyncListEntries)
						r.Delete("/by-company/{companyId}", listH.DeleteEntryByCompany)
						r.Route("/{entryId}", func(r chi.Router) {
							r.Put("/", listH.UpdateEntry)
							r.Delete("/", listH.DeleteEntry)
							r.Get("/history", listH.GetEntryHistory)

							r.Route("/rounds", func(r chi.Router) {
								r.Post("/", listH.CreateRound)
								r.Put("/{roundId}", listH.UpdateRound)
								r.Delete("/{roundId}", listH.DeleteRound)
							})
						})
					})
				})
			})

			// --- Lists by company (Session 03) ---
			r.Get("/lists/by-company/{companyId}", listH.ListsForCompany)

			// --- Company list counts (Session 04) ---
			r.Get("/lists/company-counts", listH.CompanyListCounts)

			// --- Dashboard (Sprint 2) ---
			r.Get("/dashboard", listH.GetDashboard)

			// --- Cross-list entry lookup (Feedback batch 4+5) ---
			r.Get("/entries", listH.ListEntriesAcrossLists)

			// --- User settings (Sprint 2) ---
			userH := NewUserHandler(svc.User)
			r.Route("/users/me", func(r chi.Router) {
				r.Put("/", userH.UpdateProfile)
				r.Put("/password", userH.ChangePassword)
				r.Delete("/", userH.DeleteAccount)
			})

			// --- SSE (Sprint 2 skeleton) ---
			sseH := NewSSEHandler()
			r.Get("/events", sseH.Events)

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

				// Feature flags
				ffH := NewAdminFeatureFlagHandler(svc.FeatureFlag)
				r.Route("/admin/feature-flags", func(r chi.Router) {
					r.Get("/", ffH.ListFlags)
					r.Put("/{id}", ffH.ToggleFlag)
				})

				// TODO (Sprint 5): Additional admin endpoints
			})
		})

		// --- Public routes (no auth) ---
		r.Route("/companies", func(r chi.Router) {
			companyH := NewCompanyHandler(svc.Company)
			r.Get("/", companyH.ListCompanies)
			r.Get("/{slug}", companyH.GetCompany)
		})

		// --- Webhooks ---
		// TODO (Sprint 3): Razorpay webhook
	})
}
