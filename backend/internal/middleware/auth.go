// Package middleware provides HTTP middleware for the CareerDock API server.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/service"
)

// authCtxKey is used to store auth info in request context.
type authCtxKey string

const (
	ctxUserID authCtxKey = "auth_user_id"
	ctxRole   authCtxKey = "auth_role"
)

// Auth middleware validates JWT tokens and enforces role-based access.
type Auth struct {
	authService *service.AuthService
}

// NewAuth creates a new Auth middleware.
func NewAuth(authService *service.AuthService) *Auth {
	return &Auth{authService: authService}
}

// RequireAuthenticated ensures the request has a valid access token.
func (a *Auth) RequireAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil || cookie.Value == "" {
			respondUnauthorized(w, "missing access token")
			return
		}

		userID, role, err := a.authService.ValidateAccessToken(cookie.Value)
		if err != nil {
			respondUnauthorized(w, "invalid or expired access token")
			return
		}

		// Store user info in context for downstream handlers
		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		ctx = context.WithValue(ctx, ctxRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePremium ensures the user has premium access.
// Must be used after RequireAuthenticated.
func (a *Auth) RequirePremium(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == uuid.Nil {
			respondUnauthorized(w, "authentication required")
			return
		}

		// Look up user to check premium status
		user, err := a.authService.GetUserByID(r.Context(), userID)
		if err != nil {
			respondUnauthorized(w, "user not found")
			return
		}

		if !user.IsPremium() {
			respondForbidden(w, "premium subscription required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole ensures the user has one of the allowed roles.
// Must be used after RequireAuthenticated.
func (a *Auth) RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFromContext(r.Context())
			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			respondForbidden(w, "insufficient permissions")
		})
	}
}

// RequireModerator ensures the user is a moderator or admin.
func (a *Auth) RequireModerator(next http.Handler) http.Handler {
	return a.RequireRole(domain.RoleModerator, domain.RoleAdmin)(next)
}

// RequireAdmin ensures the user is an admin.
func (a *Auth) RequireAdmin(next http.Handler) http.Handler {
	return a.RequireRole(domain.RoleAdmin)(next)
}

// --- Context helpers (exported for use by handlers) ---

// UserIDFromContext extracts the authenticated user's ID from context.
func UserIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ctxUserID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// RoleFromContext extracts the authenticated user's role from context.
func RoleFromContext(ctx context.Context) domain.Role {
	if role, ok := ctx.Value(ctxRole).(domain.Role); ok {
		return role
	}
	return ""
}

// --- Internal JSON error helpers ---

func respondUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"` + msg + `"}}`))
}

func respondForbidden(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"error":{"code":"FORBIDDEN","message":"` + msg + `"}}`))
}
