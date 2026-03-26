package handler

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/service"
)

// AdminHandler handles admin panel HTTP endpoints (Sprint 5).
type AdminHandler struct {
	admin *service.AdminService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// --- 5.1: Admin Company CRUD ---

type createCompanyRequest struct {
	Slug              string              `json:"slug"`
	Name              string              `json:"name"`
	LogoURL           *string             `json:"logo_url,omitempty"`
	Description       *string             `json:"description,omitempty"`
	Size              *domain.CompanySize `json:"size,omitempty"`
	Headquarters      *string             `json:"headquarters,omitempty"`
	FoundedYear       *int                `json:"founded_year,omitempty"`
	CareersPageURL    *string             `json:"careers_page_url,omitempty"`
	GlassdoorURL      *string             `json:"glassdoor_url,omitempty"`
	AmbitionboxURL    *string             `json:"ambitionbox_url,omitempty"`
	LinkedinURL       *string             `json:"linkedin_url,omitempty"`
	TechStack         []string            `json:"tech_stack"`
	Domains           []string            `json:"domains"`
	HiringStatus      domain.HiringStatus `json:"hiring_status"`
	InterviewPatterns json.RawMessage     `json:"interview_patterns,omitempty"`
	CompensationTier  *string             `json:"compensation_tier,omitempty"`
	HasRSU            bool                `json:"has_rsu"`
	HasRSURefresher   bool                `json:"has_rsu_refresher"`
	OfficeModes       []string            `json:"office_modes"`
	CompensationBands json.RawMessage     `json:"compensation_bands,omitempty"`
}

// CreateCompany handles POST /api/admin/companies.
func (h *AdminHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var req createCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	adminID := middleware.UserIDFromContext(r.Context())
	input := service.CreateCompanyInput{
		Slug:              req.Slug,
		Name:              req.Name,
		LogoURL:           req.LogoURL,
		Description:       req.Description,
		Size:              req.Size,
		Headquarters:      req.Headquarters,
		FoundedYear:       req.FoundedYear,
		CareersPageURL:    req.CareersPageURL,
		GlassdoorURL:      req.GlassdoorURL,
		AmbitionboxURL:    req.AmbitionboxURL,
		LinkedinURL:       req.LinkedinURL,
		TechStack:         req.TechStack,
		Domains:           req.Domains,
		HiringStatus:      req.HiringStatus,
		InterviewPatterns: req.InterviewPatterns,
		CompensationTier:  req.CompensationTier,
		HasRSU:            req.HasRSU,
		HasRSURefresher:   req.HasRSURefresher,
		OfficeModes:       req.OfficeModes,
		CompensationBands: req.CompensationBands,
	}

	company, err := h.admin.CreateCompany(r.Context(), adminID, input, clientIP(r))
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, DataResponse{Data: company})
}

type updateCompanyRequest struct {
	Name              *string              `json:"name,omitempty"`
	Slug              *string              `json:"slug,omitempty"`
	LogoURL           *string              `json:"logo_url,omitempty"`
	Description       *string              `json:"description,omitempty"`
	Size              *domain.CompanySize  `json:"size,omitempty"`
	Headquarters      *string              `json:"headquarters,omitempty"`
	FoundedYear       *int                 `json:"founded_year,omitempty"`
	CareersPageURL    *string              `json:"careers_page_url,omitempty"`
	GlassdoorURL      *string              `json:"glassdoor_url,omitempty"`
	AmbitionboxURL    *string              `json:"ambitionbox_url,omitempty"`
	LinkedinURL       *string              `json:"linkedin_url,omitempty"`
	TechStack         []string             `json:"tech_stack,omitempty"`
	Domains           []string             `json:"domains,omitempty"`
	HiringStatus      *domain.HiringStatus `json:"hiring_status,omitempty"`
	InterviewPatterns json.RawMessage      `json:"interview_patterns,omitempty"`
	CompensationTier  *string              `json:"compensation_tier,omitempty"`
	HasRSU            *bool                `json:"has_rsu,omitempty"`
	HasRSURefresher   *bool                `json:"has_rsu_refresher,omitempty"`
	OfficeModes       []string             `json:"office_modes,omitempty"`
	CompensationBands json.RawMessage      `json:"compensation_bands,omitempty"`
	LastVerifiedAt    *time.Time           `json:"last_verified_at,omitempty"`
}

// UpdateCompany handles PUT /api/admin/companies/{id}.
func (h *AdminHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid company id", map[string]any{"field": "id"}))
		return
	}

	var req updateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	adminID := middleware.UserIDFromContext(r.Context())
	input := service.UpdateCompanyInput{
		Name:              req.Name,
		Slug:              req.Slug,
		LogoURL:           req.LogoURL,
		Description:       req.Description,
		Size:              req.Size,
		Headquarters:      req.Headquarters,
		FoundedYear:       req.FoundedYear,
		CareersPageURL:    req.CareersPageURL,
		GlassdoorURL:      req.GlassdoorURL,
		AmbitionboxURL:    req.AmbitionboxURL,
		LinkedinURL:       req.LinkedinURL,
		TechStack:         req.TechStack,
		Domains:           req.Domains,
		HiringStatus:      req.HiringStatus,
		InterviewPatterns: req.InterviewPatterns,
		CompensationTier:  req.CompensationTier,
		HasRSU:            req.HasRSU,
		HasRSURefresher:   req.HasRSURefresher,
		OfficeModes:       req.OfficeModes,
		CompensationBands: req.CompensationBands,
		LastVerifiedAt:    req.LastVerifiedAt,
	}

	company, err := h.admin.UpdateCompany(r.Context(), adminID, companyID, input, clientIP(r))
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: company})
}

// DeleteCompany handles DELETE /api/admin/companies/{id}.
func (h *AdminHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid company id", map[string]any{"field": "id"}))
		return
	}

	adminID := middleware.UserIDFromContext(r.Context())
	if err := h.admin.DeleteCompany(r.Context(), adminID, companyID, clientIP(r)); err != nil {
		respondError(w, r, err)
		return
	}

	respondMessage(w, http.StatusOK, "company deleted")
}

const maxLogoSize = 2 * 1024 * 1024 // 2 MB

// UploadCompanyLogo handles POST /api/admin/companies/{id}/logo.
func (h *AdminHandler) UploadCompanyLogo(w http.ResponseWriter, r *http.Request) {
	companyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid company id", map[string]any{"field": "id"}))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxLogoSize)

	file, header, err := r.FormFile("logo")
	if err != nil {
		respondError(w, r, domain.ValidationError("logo file is required (max 2MB)", map[string]any{"field": "logo"}))
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		respondError(w, r, domain.ValidationError("file must be an image", map[string]any{"field": "logo"}))
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		respondError(w, r, domain.ValidationError("failed to read logo file", nil))
		return
	}

	adminID := middleware.UserIDFromContext(r.Context())
	key, err := h.admin.UploadCompanyLogo(r.Context(), adminID, companyID, data, contentType, clientIP(r))
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: map[string]string{"key": key}})
}

// --- 5.2: Admin User Management ---

// ListUsers handles GET /api/admin/users.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	filter := domain.UserFilter{
		Query:  r.URL.Query().Get("q"),
		Limit:  parseIntParam(r, "limit", 50),
		Offset: parseIntParam(r, "offset", 0),
	}
	if roleStr := r.URL.Query().Get("role"); roleStr != "" {
		role := domain.Role(roleStr)
		filter.Role = &role
	}

	users, total, err := h.admin.ListUsers(r.Context(), filter)
	if err != nil {
		respondError(w, r, err)
		return
	}

	type adminUserResponse struct {
		ID            uuid.UUID   `json:"id"`
		Email         string      `json:"email"`
		Name          string      `json:"name"`
		Role          domain.Role `json:"role"`
		PremiumSince  *time.Time  `json:"premium_since,omitempty"`
		EmailVerified bool        `json:"email_verified"`
		DeletedAt     *time.Time  `json:"deleted_at,omitempty"`
		CreatedAt     time.Time   `json:"created_at"`
		UpdatedAt     time.Time   `json:"updated_at"`
	}

	resp := make([]adminUserResponse, len(users))
	for i, u := range users {
		resp[i] = adminUserResponse{
			ID:            u.ID,
			Email:         u.Email,
			Name:          u.Name,
			Role:          u.Role,
			PremiumSince:  u.PremiumSince,
			EmailVerified: u.EmailVerified,
			DeletedAt:     u.DeletedAt,
			CreatedAt:     u.CreatedAt,
			UpdatedAt:     u.UpdatedAt,
		}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":  resp,
		"total": total,
	})
}

type updateUserRequest struct {
	Role       *domain.Role `json:"role,omitempty"`
	SetPremium *bool        `json:"set_premium,omitempty"`
	Banned     *bool        `json:"banned,omitempty"`
}

// UpdateUser handles PUT /api/admin/users/{id}.
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid user id", map[string]any{"field": "id"}))
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	adminID := middleware.UserIDFromContext(r.Context())
	input := service.AdminUpdateUserInput{
		Role:       req.Role,
		SetPremium: req.SetPremium,
		Banned:     req.Banned,
	}

	user, err := h.admin.UpdateUser(r.Context(), adminID, userID, input, clientIP(r))
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: user})
}

// --- 5.3: Admin Credit Management ---

type allocateCreditsRequest struct {
	CreditType string `json:"credit_type"`
	Amount     int    `json:"amount"`
	Reason     string `json:"reason"`
}

// AllocateCredits handles POST /api/admin/users/{id}/credits.
func (h *AdminHandler) AllocateCredits(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid user id", map[string]any{"field": "id"}))
		return
	}

	var req allocateCreditsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	ct, err := service.ValidateCreditType(req.CreditType)
	if err != nil {
		respondError(w, r, err)
		return
	}

	adminID := middleware.UserIDFromContext(r.Context())
	input := service.AdminAllocateCreditsInput{
		UserID:     userID,
		CreditType: ct,
		Amount:     req.Amount,
		Reason:     req.Reason,
	}

	if err := h.admin.AllocateCredits(r.Context(), adminID, input, clientIP(r)); err != nil {
		respondError(w, r, err)
		return
	}

	respondMessage(w, http.StatusOK, "credits allocated")
}

// --- 5.4: Admin Payment & Transaction Logs ---

// ListPayments handles GET /api/admin/payments.
func (h *AdminHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	filter := domain.PaymentFilter{
		Limit:  parseIntParam(r, "limit", 50),
		Offset: parseIntParam(r, "offset", 0),
	}
	if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			respondError(w, r, domain.ValidationError("invalid user_id", map[string]any{"field": "user_id"}))
			return
		}
		filter.UserID = &uid
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.PaymentStatus(statusStr)
		filter.Status = &status
	}

	payments, total, err := h.admin.ListPayments(r.Context(), filter)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":  payments,
		"total": total,
	})
}

// ListCreditTransactions handles GET /api/admin/credits/transactions.
func (h *AdminHandler) ListCreditTransactions(w http.ResponseWriter, r *http.Request) {
	filter := domain.CreditTransactionFilter{
		Limit:  parseIntParam(r, "limit", 50),
		Offset: parseIntParam(r, "offset", 0),
	}
	if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			respondError(w, r, domain.ValidationError("invalid user_id", map[string]any{"field": "user_id"}))
			return
		}
		filter.UserID = &uid
	}
	if ctStr := r.URL.Query().Get("credit_type"); ctStr != "" {
		ct, err := service.ValidateCreditType(ctStr)
		if err != nil {
			respondError(w, r, err)
			return
		}
		filter.CreditType = &ct
	}

	txns, total, err := h.admin.ListCreditTransactions(r.Context(), filter)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":  txns,
		"total": total,
	})
}

// --- Helpers ---

// parseIntParam parses an integer query parameter with a default value.
func parseIntParam(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// clientIP extracts the client IP from the request, preferring X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (client)
		if idx := strings.IndexByte(xff, ','); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xff := r.Header.Get("X-Real-IP"); xff != "" {
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
