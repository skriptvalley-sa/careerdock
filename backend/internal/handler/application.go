package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// ApplicationHandler handles application HTTP endpoints.
type ApplicationHandler struct {
	apps *service.ApplicationService
}

// NewApplicationHandler creates a new ApplicationHandler.
func NewApplicationHandler(apps *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{apps: apps}
}

// --- Request DTOs ---

type createApplicationRequest struct {
	CompanyID   string  `json:"company_id"`
	RoleTitle   *string `json:"role_title"`
	Status      *string `json:"status"`
	DateApplied *string `json:"date_applied"`
	Notes       *string `json:"notes"`
}

type updateApplicationRequest struct {
	RoleTitle   *string `json:"role_title"`
	Status      *string `json:"status"`
	DateApplied *string `json:"date_applied"`
	Notes       *string `json:"notes"`
}

type createRoundRequest struct {
	RoundNumber   int     `json:"round_number"`
	RoundType     string  `json:"round_type"`
	ScheduledDate *string `json:"scheduled_date"`
	Outcome       *string `json:"outcome"`
	Notes         *string `json:"notes"`
}

type updateRoundRequest struct {
	Outcome       *string `json:"outcome"`
	Notes         *string `json:"notes"`
	ScheduledDate *string `json:"scheduled_date"`
}

// --- Response DTOs ---

type applicationResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	CompanyID   string  `json:"company_id"`
	CompanyName string  `json:"company_name,omitempty"`
	RoleTitle   *string `json:"role_title,omitempty"`
	Status      string  `json:"status"`
	DateApplied *string `json:"date_applied,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type statusHistoryResponse struct {
	ID         string  `json:"id"`
	FromStatus *string `json:"from_status,omitempty"`
	ToStatus   string  `json:"to_status"`
	ChangedAt  string  `json:"changed_at"`
}

type interviewRoundResponse struct {
	ID            string  `json:"id"`
	RoundNumber   int     `json:"round_number"`
	RoundType     string  `json:"round_type"`
	ScheduledDate *string `json:"scheduled_date,omitempty"`
	Outcome       string  `json:"outcome"`
	Notes         *string `json:"notes,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// --- Converters ---

func toApplicationResponse(a *domain.Application) applicationResponse {
	resp := applicationResponse{
		ID:        a.ID.String(),
		UserID:    a.UserID.String(),
		CompanyID: a.CompanyID.String(),
		RoleTitle: a.RoleTitle,
		Status:    string(a.Status),
		Notes:     a.Notes,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
	if a.DateApplied != nil {
		s := a.DateApplied.Format("2006-01-02")
		resp.DateApplied = &s
	}
	return resp
}

func toStatusHistoryResponse(h domain.StatusHistory) statusHistoryResponse {
	resp := statusHistoryResponse{
		ID:        h.ID.String(),
		ToStatus:  string(h.ToStatus),
		ChangedAt: h.ChangedAt.Format(time.RFC3339),
	}
	if h.FromStatus != nil {
		s := string(*h.FromStatus)
		resp.FromStatus = &s
	}
	return resp
}

func toInterviewRoundResponse(r *domain.InterviewRound) interviewRoundResponse {
	resp := interviewRoundResponse{
		ID:          r.ID.String(),
		RoundNumber: r.RoundNumber,
		RoundType:   r.RoundType,
		Outcome:     string(r.Outcome),
		Notes:       r.Notes,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.Format(time.RFC3339),
	}
	if r.ScheduledDate != nil {
		s := r.ScheduledDate.Format("2006-01-02")
		resp.ScheduledDate = &s
	}
	return resp
}

// --- Handlers ---

// ListApplications handles GET /api/applications.
func (h *ApplicationHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var statusFilter *domain.ApplicationStatus
	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.ApplicationStatus(s)
		statusFilter = &status
	}

	apps, err := h.apps.ListApplications(r.Context(), userID, statusFilter, true)
	if err != nil {
		respondError(w, r, err)
		return
	}

	items := make([]applicationResponse, len(apps))
	for i, a := range apps {
		resp := toApplicationResponse(&a.Application)
		resp.CompanyName = a.CompanyName
		items[i] = resp
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: items})
}

// CreateApplication handles POST /api/applications.
func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var req createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid company_id", nil))
		return
	}

	var roleTitle *string
	if req.RoleTitle != nil {
		trimmed := strings.TrimSpace(*req.RoleTitle)
		if trimmed != "" {
			roleTitle = &trimmed
		}
	}

	var status domain.ApplicationStatus
	if req.Status != nil {
		status = domain.ApplicationStatus(*req.Status)
	}

	var dateApplied *time.Time
	if req.DateApplied != nil {
		t, err := time.Parse("2006-01-02", *req.DateApplied)
		if err != nil {
			respondError(w, r, domain.ValidationError("invalid date_applied format (expected YYYY-MM-DD)", nil))
			return
		}
		dateApplied = &t
	}

	userID := middleware.UserIDFromContext(r.Context())
	app, err := h.apps.CreateApplication(r.Context(), service.CreateApplicationInput{
		UserID:      userID,
		CompanyID:   companyID,
		RoleTitle:   roleTitle,
		Status:      status,
		DateApplied: dateApplied,
		Notes:       req.Notes,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, DataResponse{Data: toApplicationResponse(app)})
}

// GetApplication handles GET /api/applications/{id}.
func (h *ApplicationHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid application ID", nil))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	history, err := h.apps.GetApplicationHistory(r.Context(), appID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	items := make([]statusHistoryResponse, len(history))
	for i, sh := range history {
		items[i] = toStatusHistoryResponse(sh)
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: items})
}

// UpdateApplication handles PUT /api/applications/{id}.
func (h *ApplicationHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid application ID", nil))
		return
	}

	var req updateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	var status *domain.ApplicationStatus
	if req.Status != nil {
		s := domain.ApplicationStatus(*req.Status)
		status = &s
	}

	var dateApplied *time.Time
	if req.DateApplied != nil {
		t, err := time.Parse("2006-01-02", *req.DateApplied)
		if err != nil {
			respondError(w, r, domain.ValidationError("invalid date_applied format", nil))
			return
		}
		dateApplied = &t
	}

	userID := middleware.UserIDFromContext(r.Context())
	app, err := h.apps.UpdateApplication(r.Context(), service.UpdateApplicationInput{
		ApplicationID: appID,
		UserID:        userID,
		RoleTitle:     req.RoleTitle,
		Status:        status,
		DateApplied:   dateApplied,
		Notes:         req.Notes,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: toApplicationResponse(app)})
}

// DeleteApplication handles DELETE /api/applications/{id}.
func (h *ApplicationHandler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid application ID", nil))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.apps.DeleteApplication(r.Context(), appID, userID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDashboard handles GET /api/dashboard.
func (h *ApplicationHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	counts, err := h.apps.GetDashboardCounts(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: counts})
}

// ListByCompany handles GET /api/applications/by-company/{companyId}.
func (h *ApplicationHandler) ListByCompany(w http.ResponseWriter, r *http.Request) {
	companyIDStr := chi.URLParam(r, "companyId")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid companyId", nil))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	apps, err := h.apps.ListByCompany(r.Context(), userID, companyID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	items := make([]applicationResponse, len(apps))
	for i, a := range apps {
		items[i] = toApplicationResponse(&a)
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: items})
}

// --- Interview Round handlers ---

// CreateRound handles POST /api/applications/{id}/rounds.
func (h *ApplicationHandler) CreateRound(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid application ID", nil))
		return
	}

	var req createRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	req.RoundType = strings.TrimSpace(req.RoundType)
	if req.RoundType == "" {
		respondError(w, r, domain.ValidationError("round_type is required", nil))
		return
	}
	if req.RoundNumber < 1 {
		respondError(w, r, domain.ValidationError("round_number must be positive", nil))
		return
	}

	var scheduledDate *time.Time
	if req.ScheduledDate != nil {
		t, err := time.Parse("2006-01-02", *req.ScheduledDate)
		if err != nil {
			respondError(w, r, domain.ValidationError("invalid scheduled_date format", nil))
			return
		}
		scheduledDate = &t
	}

	var outcome domain.InterviewOutcome
	if req.Outcome != nil {
		outcome = domain.InterviewOutcome(*req.Outcome)
	}

	userID := middleware.UserIDFromContext(r.Context())
	round, err := h.apps.CreateRound(r.Context(), service.CreateRoundInput{
		ApplicationID: appID,
		UserID:        userID,
		RoundNumber:   req.RoundNumber,
		RoundType:     req.RoundType,
		ScheduledDate: scheduledDate,
		Outcome:       outcome,
		Notes:         req.Notes,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, DataResponse{Data: toInterviewRoundResponse(round)})
}

// UpdateRound handles PUT /api/applications/{id}/rounds/{roundId}.
func (h *ApplicationHandler) UpdateRound(w http.ResponseWriter, r *http.Request) {
	roundID, err := uuid.Parse(chi.URLParam(r, "roundId"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid round ID", nil))
		return
	}

	var req updateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	var outcome *domain.InterviewOutcome
	if req.Outcome != nil {
		o := domain.InterviewOutcome(*req.Outcome)
		outcome = &o
	}

	var scheduledDate *time.Time
	if req.ScheduledDate != nil {
		t, err := time.Parse("2006-01-02", *req.ScheduledDate)
		if err != nil {
			respondError(w, r, domain.ValidationError("invalid scheduled_date format", nil))
			return
		}
		scheduledDate = &t
	}

	userID := middleware.UserIDFromContext(r.Context())
	round, err := h.apps.UpdateRound(r.Context(), service.UpdateRoundInput{
		RoundID:       roundID,
		UserID:        userID,
		Outcome:       outcome,
		Notes:         req.Notes,
		ScheduledDate: scheduledDate,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: toInterviewRoundResponse(round)})
}

// DeleteRound handles DELETE /api/applications/{id}/rounds/{roundId}.
func (h *ApplicationHandler) DeleteRound(w http.ResponseWriter, r *http.Request) {
	roundID, err := uuid.Parse(chi.URLParam(r, "roundId"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid round ID", nil))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.apps.DeleteRound(r.Context(), roundID, userID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
