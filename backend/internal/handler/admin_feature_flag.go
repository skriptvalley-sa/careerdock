package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/service"
)

// AdminFeatureFlagHandler handles admin feature flag HTTP endpoints.
type AdminFeatureFlagHandler struct {
	flags *service.FeatureFlagService
}

// NewAdminFeatureFlagHandler creates a new AdminFeatureFlagHandler.
func NewAdminFeatureFlagHandler(flags *service.FeatureFlagService) *AdminFeatureFlagHandler {
	return &AdminFeatureFlagHandler{flags: flags}
}

// --- Request/Response DTOs ---

type toggleFlagRequest struct {
	Enabled     bool    `json:"enabled"`
	Description *string `json:"description,omitempty"`
}

type featureFlagResponse struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Enabled     bool      `json:"enabled"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func toFeatureFlagResponse(f *domain.FeatureFlag) featureFlagResponse {
	return featureFlagResponse{
		ID:          f.ID,
		Key:         f.Key,
		Enabled:     f.Enabled,
		Description: f.Description,
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// --- Handlers ---

// ListFlags handles GET /api/admin/feature-flags.
func (h *AdminFeatureFlagHandler) ListFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.flags.List(r.Context())
	if err != nil {
		respondError(w, r, err)
		return
	}

	resp := make([]featureFlagResponse, len(flags))
	for i := range flags {
		resp[i] = toFeatureFlagResponse(&flags[i])
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: resp})
}

// ToggleFlag handles PUT /api/admin/feature-flags/{id}.
func (h *AdminFeatureFlagHandler) ToggleFlag(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid flag id", map[string]any{"field": "id"}))
		return
	}

	var req toggleFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	flag, err := h.flags.Toggle(r.Context(), id, req.Enabled, req.Description)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: toFeatureFlagResponse(flag)})
}
