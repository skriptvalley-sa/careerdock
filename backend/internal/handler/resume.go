package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// ResumeHandler handles resume-related HTTP endpoints.
type ResumeHandler struct {
	resumeSvc *service.ResumeService
}

// NewResumeHandler creates a new ResumeHandler.
func NewResumeHandler(resumeSvc *service.ResumeService) *ResumeHandler {
	return &ResumeHandler{resumeSvc: resumeSvc}
}

// --- Response DTOs ---

type resumeListResponse struct {
	ID              uuid.UUID           `json:"id"`
	SlotNumber      int                 `json:"slot_number"`
	FileName        string              `json:"file_name"`
	FileSizeBytes   int                 `json:"file_size_bytes"`
	Status          domain.ResumeStatus `json:"status"`
	FailureReason   *string             `json:"failure_reason,omitempty"`
	IsDefault       bool                `json:"is_default"`
	ATSGeneralScore *int                `json:"ats_general_score,omitempty"`
	ParsedSummary   *parsedSummary      `json:"parsed_data_summary,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type parsedSummary struct {
	YearsOfExperience float64  `json:"years_of_experience"`
	RoleLevel         string   `json:"role_level"`
	TopSkills         []string `json:"top_skills"`
	Domains           []string `json:"domains"`
}

type resumeDetailResponse struct {
	ID            uuid.UUID           `json:"id"`
	SlotNumber    int                 `json:"slot_number"`
	FileName      string              `json:"file_name"`
	FileSizeBytes int                 `json:"file_size_bytes"`
	Status        domain.ResumeStatus `json:"status"`
	IsDefault     bool                `json:"is_default"`
	ParsedData    json.RawMessage     `json:"parsed_data,omitempty"`
	ATSGeneral    json.RawMessage     `json:"ats_general,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type downloadURLResponse struct {
	DownloadURL string `json:"download_url"`
	ExpiresIn   int    `json:"expires_in_seconds"`
}

// --- Handlers ---

// ListResumes handles GET /api/resumes.
func (h *ResumeHandler) ListResumes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	resumes, err := h.resumeSvc.ListResumes(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	result := make([]resumeListResponse, 0, len(resumes))
	for i := range resumes {
		result = append(result, toResumeListResponse(&resumes[i]))
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": result})
}

// UploadResume handles POST /api/resumes (multipart/form-data).
func (h *ResumeHandler) UploadResume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	// Parse multipart form (max 6 MB to account for overhead)
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		respondError(w, r, domain.ValidationError("Invalid multipart form", nil))
		return
	}

	// Read slot_number
	slotStr := r.FormValue("slot_number")
	slotNumber, err := strconv.Atoi(slotStr)
	if err != nil || slotNumber < 1 || slotNumber > 3 {
		respondError(w, r, domain.ValidationError("slot_number must be 1, 2, or 3", nil))
		return
	}

	// Read file
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

	resume, err := h.resumeSvc.UploadResume(r.Context(), userID, header.Filename, slotNumber, fileData)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"data": toResumeDetailResponse(resume)})
}

// GetResume handles GET /api/resumes/{id}.
func (h *ResumeHandler) GetResume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	resumeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid resume ID", nil))
		return
	}

	resume, err := h.resumeSvc.GetResume(r.Context(), userID, resumeID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": toResumeDetailResponse(resume)})
}

// SetDefaultResume handles PUT /api/resumes/{id}/default.
func (h *ResumeHandler) SetDefaultResume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	resumeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid resume ID", nil))
		return
	}

	if err := h.resumeSvc.SetDefault(r.Context(), userID, resumeID); err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"message": "Resume set as default"}})
}

// ArchiveResume handles DELETE /api/resumes/{id}.
func (h *ResumeHandler) ArchiveResume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	resumeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid resume ID", nil))
		return
	}

	if err := h.resumeSvc.ArchiveResume(r.Context(), userID, resumeID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RetryResume handles POST /api/resumes/{id}/retry.
// Resets a failed resume to "parsing" and re-enqueues the AI processing task.
func (h *ResumeHandler) RetryResume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	resumeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid resume ID", nil))
		return
	}

	if err := h.resumeSvc.RetryParsing(r.Context(), userID, resumeID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetResumeDownloadURL handles GET /api/resumes/{id}/download.
func (h *ResumeHandler) GetResumeDownloadURL(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	resumeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid resume ID", nil))
		return
	}

	url, err := h.resumeSvc.GetDownloadURL(r.Context(), userID, resumeID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": downloadURLResponse{
		DownloadURL: url,
		ExpiresIn:   900, // 15 minutes
	}})
}

// --- Converters ---

func toResumeListResponse(r *domain.Resume) resumeListResponse {
	resp := resumeListResponse{
		ID:            r.ID,
		SlotNumber:    r.SlotNumber,
		FileName:      r.FileName,
		FileSizeBytes: r.FileSizeBytes,
		Status:        r.Status,
		FailureReason: r.FailureReason,
		IsDefault:     r.IsDefault,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}

	// Extract top-level ATS score if available
	if len(r.ATSGeneral) > 0 {
		var ats struct {
			Score int `json:"score"`
		}
		if err := json.Unmarshal(r.ATSGeneral, &ats); err == nil && ats.Score > 0 {
			resp.ATSGeneralScore = &ats.Score
		}
	}

	// Extract summary from parsed data if available
	if len(r.ParsedData) > 0 {
		var parsed struct {
			YearsOfExperience float64 `json:"years_of_experience"`
			RoleLevel         string  `json:"role_level"`
			Skills            struct {
				Languages  []string `json:"languages"`
				Frameworks []string `json:"frameworks"`
			} `json:"skills"`
			Domains []string `json:"domains"`
		}
		if err := json.Unmarshal(r.ParsedData, &parsed); err == nil {
			topSkills := make([]string, 0)
			// Take up to 3 from languages + frameworks
			for _, s := range parsed.Skills.Languages {
				if len(topSkills) >= 3 {
					break
				}
				topSkills = append(topSkills, s)
			}
			for _, s := range parsed.Skills.Frameworks {
				if len(topSkills) >= 3 {
					break
				}
				topSkills = append(topSkills, s)
			}

			resp.ParsedSummary = &parsedSummary{
				YearsOfExperience: parsed.YearsOfExperience,
				RoleLevel:         parsed.RoleLevel,
				TopSkills:         topSkills,
				Domains:           parsed.Domains,
			}
		}
	}

	return resp
}

func toResumeDetailResponse(r *domain.Resume) resumeDetailResponse {
	return resumeDetailResponse{
		ID:            r.ID,
		SlotNumber:    r.SlotNumber,
		FileName:      r.FileName,
		FileSizeBytes: r.FileSizeBytes,
		Status:        r.Status,
		IsDefault:     r.IsDefault,
		ParsedData:    r.ParsedData,
		ATSGeneral:    r.ATSGeneral,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
