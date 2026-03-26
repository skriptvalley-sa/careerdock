package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// ATSHandler handles ATS check HTTP endpoints.
type ATSHandler struct {
	atsSvc *service.ATSService
}

// NewATSHandler creates a new ATSHandler.
func NewATSHandler(atsSvc *service.ATSService) *ATSHandler {
	return &ATSHandler{atsSvc: atsSvc}
}

// --- Request DTOs ---

type checkCompanyRequest struct {
	ResumeID  uuid.UUID `json:"resume_id"`
	CompanyID uuid.UUID `json:"company_id"`
}

type checkJobRequest struct {
	ResumeID       uuid.UUID `json:"resume_id"`
	JobDescription string    `json:"job_description"`
}

type checkResumeRequest struct {
	ResumeID uuid.UUID `json:"resume_id"`
}

// --- Response DTO ---

type atsCheckResponse struct {
	ID          uuid.UUID           `json:"id"`
	CheckType   domain.ATSCheckType `json:"check_type"`
	ResumeID    *uuid.UUID          `json:"resume_id,omitempty"` // nil for temp-upload checks
	CompanyID   *uuid.UUID          `json:"company_id,omitempty"`
	CompanyName *string             `json:"company_name,omitempty"`
	Result      json.RawMessage     `json:"result"`
	CreatedAt   time.Time           `json:"created_at"`
}

// --- Handlers ---

// CheckCompany handles POST /api/ats/company.
func (h *ATSHandler) CheckCompany(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req checkCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}
	if req.ResumeID == uuid.Nil {
		respondError(w, r, domain.ValidationError("resume_id is required", nil))
		return
	}
	if req.CompanyID == uuid.Nil {
		respondError(w, r, domain.ValidationError("company_id is required", nil))
		return
	}

	check, err := h.atsSvc.CheckCompany(r.Context(), userID, req.ResumeID, req.CompanyID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]any{"data": toATSCheckResponse(check)})
}

// CheckJob handles POST /api/ats/job.
func (h *ATSHandler) CheckJob(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req checkJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}
	if req.ResumeID == uuid.Nil {
		respondError(w, r, domain.ValidationError("resume_id is required", nil))
		return
	}
	if req.JobDescription == "" {
		respondError(w, r, domain.ValidationError("job_description is required", nil))
		return
	}

	check, err := h.atsSvc.CheckJob(r.Context(), userID, req.ResumeID, req.JobDescription)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]any{"data": toATSCheckResponse(check)})
}

// CheckResume handles POST /api/ats/resume.
func (h *ATSHandler) CheckResume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req checkResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}
	if req.ResumeID == uuid.Nil {
		respondError(w, r, domain.ValidationError("resume_id is required", nil))
		return
	}

	check, err := h.atsSvc.CheckResume(r.Context(), userID, req.ResumeID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]any{"data": toATSCheckResponse(check)})
}

// CheckResumeTempUpload handles POST /api/ats/resume/upload (multipart/form-data).
// Accepts a PDF without requiring an existing resume slot, uploads it to a
// temporary S3 path, and enqueues a resume-only ATS check.
func (h *ATSHandler) CheckResumeTempUpload(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	if err := r.ParseMultipartForm(6 << 20); err != nil {
		respondError(w, r, domain.ValidationError("Invalid multipart form", nil))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, r, domain.ValidationError("Missing file field", nil))
		return
	}
	defer func() { _ = file.Close() }()

	fileData, err := io.ReadAll(file)
	if err != nil {
		respondError(w, r, domain.InternalError(err))
		return
	}

	check, err := h.atsSvc.CheckResumeTempUpload(r.Context(), userID, fileData, header.Filename)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]any{"data": toATSCheckResponse(check)})
}

// GetCheck handles GET /api/ats/{id}.
func (h *ATSHandler) GetCheck(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	checkID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid ATS check ID", nil))
		return
	}

	check, err := h.atsSvc.GetCheck(r.Context(), userID, checkID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": toATSCheckResponse(check)})
}

// ListChecks handles GET /api/ats/.
func (h *ATSHandler) ListChecks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	checks, err := h.atsSvc.ListChecks(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	result := make([]atsCheckResponse, 0, len(checks))
	for i := range checks {
		result = append(result, toATSCheckResponse(&checks[i]))
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": result})
}

// --- Converter ---

func toATSCheckResponse(c *domain.ATSCheck) atsCheckResponse {
	return atsCheckResponse{
		ID:          c.ID,
		CheckType:   c.CheckType,
		ResumeID:    c.ResumeID, // *uuid.UUID — nil for temp-upload checks
		CompanyID:   c.CompanyID,
		CompanyName: c.CompanyName,
		Result:      c.Result,
		CreatedAt:   c.CreatedAt,
	}
}
