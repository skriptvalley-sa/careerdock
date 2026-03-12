package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	auth         *service.AuthService
	secureCookie bool // true in production (Secure flag on cookies)
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(auth *service.AuthService, secureCookie bool) *AuthHandler {
	return &AuthHandler{
		auth:         auth,
		secureCookie: secureCookie,
	}
}

// --- Request/Response DTOs ---

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// userResponse is the public user representation (no password hash).
type userResponse struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Name                string     `json:"name"`
	Role                string     `json:"role"`
	EmailVerified       bool       `json:"email_verified"`
	PremiumSince        *time.Time `json:"premium_since,omitempty"`
	CurrentTitle        *string    `json:"current_title,omitempty"`
	ExperienceLevel     *string    `json:"experience_level,omitempty"`
	PreferredTechStacks []string   `json:"preferred_tech_stacks"`
	TargetDomains       []string   `json:"target_domains"`
	TargetLocations     []string   `json:"target_locations"`
	DefaultResumeID     *string    `json:"default_resume_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func toUserResponse(u *domain.User) userResponse {
	resp := userResponse{
		ID:                  u.ID.String(),
		Email:               u.Email,
		Name:                u.Name,
		Role:                string(u.Role),
		EmailVerified:       u.EmailVerified,
		PremiumSince:        u.PremiumSince,
		CurrentTitle:        u.CurrentTitle,
		PreferredTechStacks: u.PreferredTechStacks,
		TargetDomains:       u.TargetDomains,
		TargetLocations:     u.TargetLocations,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	}

	if u.ExperienceLevel != nil {
		s := string(*u.ExperienceLevel)
		resp.ExperienceLevel = &s
	}

	if u.DefaultResumeID != nil {
		s := u.DefaultResumeID.String()
		resp.DefaultResumeID = &s
	}

	// Ensure slices are never null in JSON
	if resp.PreferredTechStacks == nil {
		resp.PreferredTechStacks = []string{}
	}
	if resp.TargetDomains == nil {
		resp.TargetDomains = []string{}
	}
	if resp.TargetLocations == nil {
		resp.TargetLocations = []string{}
	}

	return resp
}

// --- Handlers ---

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	// Basic input validation
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || !isValidEmail(req.Email) {
		respondError(w, r, domain.ValidationError("invalid email address", map[string]any{
			"field": "email",
		}))
		return
	}
	if req.Name == "" || len(req.Name) > 255 {
		respondError(w, r, domain.ValidationError("name must be 1-255 characters", map[string]any{
			"field": "name",
		}))
		return
	}

	user, tokens, err := h.auth.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	// Set auth cookies
	h.setTokenCookies(w, tokens)

	respondJSON(w, http.StatusCreated, DataResponse{Data: toUserResponse(user)})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, tokens, err := h.auth.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	h.setTokenCookies(w, tokens)

	respondJSON(w, http.StatusOK, DataResponse{Data: toUserResponse(user)})
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		respondError(w, r, domain.Unauthorized("missing refresh token"))
		return
	}

	_, tokens, err := h.auth.Refresh(r.Context(), cookie.Value)
	if err != nil {
		respondError(w, r, err)
		return
	}

	h.setTokenCookies(w, tokens)

	respondMessage(w, http.StatusOK, "Tokens refreshed")
}

// Logout handles POST /api/auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		_ = h.auth.Logout(r.Context(), cookie.Value)
	}

	// Clear cookies
	h.clearTokenCookies(w)

	w.WriteHeader(http.StatusNoContent)
}

// VerifyEmail handles POST /api/auth/verify-email.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		respondError(w, r, domain.ValidationError("token is required", map[string]any{
			"field": "token",
		}))
		return
	}

	if err := h.auth.VerifyEmail(r.Context(), req.Token); err != nil {
		respondError(w, r, err)
		return
	}

	respondMessage(w, http.StatusOK, "Email verified successfully")
}

// ForgotPassword handles POST /api/auth/forgot-password.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Always returns success to prevent user enumeration
	_ = h.auth.ForgotPassword(r.Context(), req.Email)

	respondMessage(w, http.StatusOK, "If an account with that email exists, a reset link has been sent")
}

// ResetPassword handles POST /api/auth/reset-password.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	req.Token = strings.TrimSpace(req.Token)

	if err := h.auth.ResetPassword(r.Context(), service.ResetPasswordInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		respondError(w, r, err)
		return
	}

	respondMessage(w, http.StatusOK, "Password reset successfully")
}

// Me handles GET /api/auth/me.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	user, err := h.auth.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: toUserResponse(user)})
}

// --- Cookie helpers ---

func (h *AuthHandler) setTokenCookies(w http.ResponseWriter, tokens *service.TokenPair) {
	sameSite := http.SameSiteStrictMode

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/api",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: sameSite,
		Expires:  tokens.AccessExp,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: sameSite,
		Expires:  tokens.RefreshExp,
	})
}

func (h *AuthHandler) clearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/api",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// --- Validation helpers ---

// isValidEmail does a basic check for a valid email format.
// Full RFC 5322 validation is overkill for registration.
func isValidEmail(email string) bool {
	if len(email) > 255 {
		return false
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 || at >= len(email)-1 {
		return false
	}
	domain := email[at+1:]
	if len(domain) < 3 || !strings.Contains(domain, ".") {
		return false
	}
	return true
}
