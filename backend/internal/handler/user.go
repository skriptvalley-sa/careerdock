package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// UserHandler handles user settings HTTP endpoints.
type UserHandler struct {
	users *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// --- Request DTOs ---

type updateProfileRequest struct {
	Name                *string  `json:"name"`
	CurrentTitle        *string  `json:"current_title"`
	ExperienceLevel     *string  `json:"experience_level"`
	PreferredTechStacks []string `json:"preferred_tech_stacks"`
	TargetDomains       []string `json:"target_domains"`
	TargetLocations     []string `json:"target_locations"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type deleteAccountRequest struct {
	Password string `json:"password"`
}

// --- Handlers ---

// UpdateProfile handles PUT /api/users/me.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		req.Name = &trimmed
	}

	var expLevel *domain.ExperienceLevel
	if req.ExperienceLevel != nil {
		el := domain.ExperienceLevel(*req.ExperienceLevel)
		expLevel = &el
	}

	userID := middleware.UserIDFromContext(r.Context())
	user, err := h.users.UpdateProfile(r.Context(), service.UpdateProfileInput{
		UserID:              userID,
		Name:                req.Name,
		CurrentTitle:        req.CurrentTitle,
		ExperienceLevel:     expLevel,
		PreferredTechStacks: req.PreferredTechStacks,
		TargetDomains:       req.TargetDomains,
		TargetLocations:     req.TargetLocations,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: toUserResponse(user)})
}

// ChangePassword handles PUT /api/users/me/password.
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	if req.CurrentPassword == "" {
		respondError(w, r, domain.ValidationError("current_password is required", map[string]any{"field": "current_password"}))
		return
	}
	if req.NewPassword == "" {
		respondError(w, r, domain.ValidationError("new_password is required", map[string]any{"field": "new_password"}))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.users.ChangePassword(r.Context(), service.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		respondError(w, r, err)
		return
	}

	respondMessage(w, http.StatusOK, "Password changed successfully")
}

// DeleteAccount handles DELETE /api/users/me.
func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	var req deleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	if req.Password == "" {
		respondError(w, r, domain.ValidationError("password is required for account deletion", map[string]any{"field": "password"}))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.users.DeleteAccount(r.Context(), service.DeleteAccountInput{
		UserID:   userID,
		Password: req.Password,
	}); err != nil {
		respondError(w, r, err)
		return
	}

	respondMessage(w, http.StatusOK, "Account scheduled for deletion")
}
