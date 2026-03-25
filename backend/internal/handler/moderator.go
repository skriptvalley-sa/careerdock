package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// ModeratorHandler handles moderator-specific HTTP endpoints.
type ModeratorHandler struct {
	svc *service.ModeratorService
}

// NewModeratorHandler creates a new ModeratorHandler.
func NewModeratorHandler(svc *service.ModeratorService) *ModeratorHandler {
	return &ModeratorHandler{svc: svc}
}

// --- Request DTOs ---

type generateDraftRequest struct {
	Name        string `json:"name"`
	CareersURL  string `json:"careers_url,omitempty"`
	LinkedinURL string `json:"linkedin_url,omitempty"`
}

type submitDraftRequest struct {
	Slug              string              `json:"slug,omitempty"`
	Name              string              `json:"name"`
	Description       *string             `json:"description,omitempty"`
	Size              *domain.CompanySize `json:"size,omitempty"`
	Headquarters      *string             `json:"headquarters,omitempty"`
	FoundedYear       *int                `json:"founded_year,omitempty"`
	CareersPageURL    *string             `json:"careers_page_url,omitempty"`
	LinkedinURL       *string             `json:"linkedin_url,omitempty"`
	TechStack         []string            `json:"tech_stack"`
	Domains           []string            `json:"domains"`
	HiringStatus      domain.HiringStatus `json:"hiring_status"`
	OfficeModes       []string            `json:"office_modes"`
	CompensationTier  *string             `json:"compensation_tier,omitempty"`
	HasRSU            bool                `json:"has_rsu"`
	HasRSURefresher   bool                `json:"has_rsu_refresher"`
	CompensationBands json.RawMessage     `json:"compensation_bands,omitempty"`
}

type submitEditRequest struct {
	Description      *string              `json:"description,omitempty"`
	Size             *domain.CompanySize  `json:"size,omitempty"`
	Headquarters     *string              `json:"headquarters,omitempty"`
	CareersPageURL   *string              `json:"careers_page_url,omitempty"`
	LinkedinURL      *string              `json:"linkedin_url,omitempty"`
	TechStack        []string             `json:"tech_stack,omitempty"`
	Domains          []string             `json:"domains,omitempty"`
	HiringStatus     *domain.HiringStatus `json:"hiring_status,omitempty"`
	OfficeModes      []string             `json:"office_modes,omitempty"`
	CompensationTier *string              `json:"compensation_tier,omitempty"`
}

// --- Response DTOs ---

type editLockResponse struct {
	CompanyID uuid.UUID `json:"company_id"`
	LockedBy  uuid.UUID `json:"locked_by"`
	LockedAt  time.Time `json:"locked_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// --- Handlers ---

// GenerateCompanyDraft handles POST /api/moderator/companies/generate.
func (h *ModeratorHandler) GenerateCompanyDraft(w http.ResponseWriter, r *http.Request) {
	var req generateDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}

	draft, err := h.svc.GenerateCompanyDraft(r.Context(), req.Name, req.CareersURL, req.LinkedinURL)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": draft})
}

// SubmitCompanyDraft handles POST /api/moderator/companies.
func (h *ModeratorHandler) SubmitCompanyDraft(w http.ResponseWriter, r *http.Request) {
	var req submitDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}
	if req.Name == "" {
		respondError(w, r, domain.ValidationError("name is required", nil))
		return
	}

	company, err := h.svc.SubmitCompanyDraft(r.Context(), service.CreateCompanyInput{
		Slug:              req.Slug,
		Name:              req.Name,
		Description:       req.Description,
		Size:              req.Size,
		Headquarters:      req.Headquarters,
		FoundedYear:       req.FoundedYear,
		CareersPageURL:    req.CareersPageURL,
		LinkedinURL:       req.LinkedinURL,
		TechStack:         req.TechStack,
		Domains:           req.Domains,
		HiringStatus:      req.HiringStatus,
		OfficeModes:       req.OfficeModes,
		CompensationTier:  req.CompensationTier,
		HasRSU:            req.HasRSU,
		HasRSURefresher:   req.HasRSURefresher,
		CompensationBands: req.CompensationBands,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"data": company})
}

// AcquireLock handles POST /api/moderator/companies/{id}/lock.
func (h *ModeratorHandler) AcquireLock(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid company ID", nil))
		return
	}

	lock, err := h.svc.AcquireEditLock(r.Context(), userID, companyID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": editLockResponse{
		CompanyID: lock.CompanyID,
		LockedBy:  lock.LockedBy,
		LockedAt:  lock.LockedAt,
		ExpiresAt: lock.ExpiresAt,
	}})
}

// ReleaseLock handles DELETE /api/moderator/companies/{id}/lock.
func (h *ModeratorHandler) ReleaseLock(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid company ID", nil))
		return
	}

	if err := h.svc.ReleaseEditLock(r.Context(), userID, companyID); err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"message": "ok"})
}

// GetEditStatus handles GET /api/moderator/companies/{id}/lock.
func (h *ModeratorHandler) GetEditStatus(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid company ID", nil))
		return
	}

	lock, err := h.svc.GetEditStatus(r.Context(), companyID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	if lock == nil {
		respondJSON(w, http.StatusOK, map[string]any{"data": nil})
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": editLockResponse{
		CompanyID: lock.CompanyID,
		LockedBy:  lock.LockedBy,
		LockedAt:  lock.LockedAt,
		ExpiresAt: lock.ExpiresAt,
	}})
}

// SubmitEdit handles POST /api/moderator/companies/{id}/edit.
func (h *ModeratorHandler) SubmitEdit(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid company ID", nil))
		return
	}

	var req submitEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}

	company, err := h.svc.SubmitCompanyEdit(r.Context(), userID, companyID, service.UpdateCompanyInput{
		Description:      req.Description,
		Size:             req.Size,
		Headquarters:     req.Headquarters,
		CareersPageURL:   req.CareersPageURL,
		LinkedinURL:      req.LinkedinURL,
		TechStack:        req.TechStack,
		Domains:          req.Domains,
		HiringStatus:     req.HiringStatus,
		OfficeModes:      req.OfficeModes,
		CompensationTier: req.CompensationTier,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": company})
}
