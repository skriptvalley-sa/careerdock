package handler

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/service"
)

// CompanyHandler handles company directory HTTP endpoints.
type CompanyHandler struct {
	companies *service.CompanyService
}

// NewCompanyHandler creates a new CompanyHandler.
func NewCompanyHandler(companies *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companies: companies}
}

// --- Response DTOs ---

// companyListItem is the summary returned in list responses.
type companyListItem struct {
	ID               string   `json:"id"`
	Slug             string   `json:"slug"`
	Name             string   `json:"name"`
	LogoURL          *string  `json:"logo_url,omitempty"`
	Description      *string  `json:"description,omitempty"`
	Size             *string  `json:"size,omitempty"`
	Headquarters     *string  `json:"headquarters,omitempty"`
	TechStack        []string `json:"tech_stack"`
	Domains          []string `json:"domains"`
	HiringStatus     string   `json:"hiring_status"`
	CompensationTier *string  `json:"compensation_tier,omitempty"`
	HasRSU           bool     `json:"has_rsu"`
	HasRSURefresher  bool     `json:"has_rsu_refresher"`
	OfficeModes      []string `json:"office_modes"`
	UpdatedAt        string   `json:"updated_at"`
}

func toCompanyListItem(c domain.Company) companyListItem {
	item := companyListItem{
		ID:               c.ID.String(),
		Slug:             c.Slug,
		Name:             c.Name,
		LogoURL:          c.LogoURL,
		Description:      c.Description,
		Headquarters:     c.Headquarters,
		TechStack:        c.TechStack,
		Domains:          c.Domains,
		HiringStatus:     string(c.HiringStatus),
		CompensationTier: c.CompensationTier,
		HasRSU:           c.HasRSU,
		HasRSURefresher:  c.HasRSURefresher,
		OfficeModes:      c.OfficeModes,
		UpdatedAt:        c.UpdatedAt.Format(time.RFC3339),
	}
	if c.Size != nil {
		s := string(*c.Size)
		item.Size = &s
	}
	if item.TechStack == nil {
		item.TechStack = []string{}
	}
	if item.Domains == nil {
		item.Domains = []string{}
	}
	if item.OfficeModes == nil {
		item.OfficeModes = []string{}
	}
	return item
}

// companyDetailResponse is the full profile returned for a single company.
type companyDetailResponse struct {
	ID                string   `json:"id"`
	Slug              string   `json:"slug"`
	Name              string   `json:"name"`
	LogoURL           *string  `json:"logo_url,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Size              *string  `json:"size,omitempty"`
	Headquarters      *string  `json:"headquarters,omitempty"`
	FoundedYear       *int     `json:"founded_year,omitempty"`
	CareersPageURL    *string  `json:"careers_page_url,omitempty"`
	GlassdoorURL      *string  `json:"glassdoor_url,omitempty"`
	AmbitionboxURL    *string  `json:"ambitionbox_url,omitempty"`
	LinkedinURL       *string  `json:"linkedin_url,omitempty"`
	TechStack         []string `json:"tech_stack"`
	Domains           []string `json:"domains"`
	HiringStatus      string   `json:"hiring_status"`
	InterviewPatterns any      `json:"interview_patterns,omitempty"`
	CompensationTier  *string  `json:"compensation_tier,omitempty"`
	HasRSU            bool     `json:"has_rsu"`
	HasRSURefresher   bool     `json:"has_rsu_refresher"`
	OfficeModes       []string `json:"office_modes"`
	CompensationBands any      `json:"compensation_bands,omitempty"`
	LastVerifiedAt    *string  `json:"last_verified_at,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

func toCompanyDetailResponse(c *domain.Company) companyDetailResponse {
	resp := companyDetailResponse{
		ID:               c.ID.String(),
		Slug:             c.Slug,
		Name:             c.Name,
		LogoURL:          c.LogoURL,
		Description:      c.Description,
		Headquarters:     c.Headquarters,
		FoundedYear:      c.FoundedYear,
		CareersPageURL:   c.CareersPageURL,
		GlassdoorURL:     c.GlassdoorURL,
		AmbitionboxURL:   c.AmbitionboxURL,
		LinkedinURL:      c.LinkedinURL,
		TechStack:        c.TechStack,
		Domains:          c.Domains,
		HiringStatus:     string(c.HiringStatus),
		CompensationTier: c.CompensationTier,
		HasRSU:           c.HasRSU,
		HasRSURefresher:  c.HasRSURefresher,
		OfficeModes:      c.OfficeModes,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        c.UpdatedAt.Format(time.RFC3339),
	}

	if c.Size != nil {
		s := string(*c.Size)
		resp.Size = &s
	}
	if resp.TechStack == nil {
		resp.TechStack = []string{}
	}
	if resp.Domains == nil {
		resp.Domains = []string{}
	}
	if resp.OfficeModes == nil {
		resp.OfficeModes = []string{}
	}
	if c.LastVerifiedAt != nil {
		s := c.LastVerifiedAt.Format(time.RFC3339)
		resp.LastVerifiedAt = &s
	}
	if len(c.InterviewPatterns) > 0 {
		resp.InterviewPatterns = c.InterviewPatterns
	}
	if len(c.CompensationBands) > 0 {
		resp.CompensationBands = c.CompensationBands
	}

	return resp
}

// --- Handlers ---

// ListCompanies handles GET /api/companies.
func (h *CompanyHandler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	filter := parseCompanyFilter(r)
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	companies, nextCursor, err := h.companies.List(r.Context(), query, filter)
	if err != nil {
		respondError(w, r, err)
		return
	}

	items := make([]companyListItem, len(companies))
	for i, c := range companies {
		items[i] = toCompanyListItem(c)
	}

	resp := PaginatedResponse{
		Data: items,
		Pagination: &Pagination{
			NextCursor: nextCursor,
			HasMore:    nextCursor != "",
		},
	}

	// Cache headers for CDN (5 min)
	w.Header().Set("Cache-Control", "public, max-age=300")
	setETag(w, r, resp)

	respondJSON(w, http.StatusOK, resp)
}

// GetCompany handles GET /api/companies/{slug}.
func (h *CompanyHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondError(w, r, domain.ValidationError("slug is required", nil))
		return
	}

	company, err := h.companies.GetBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, r, err)
		return
	}

	resp := DataResponse{Data: toCompanyDetailResponse(company)}

	// Cache headers for CDN (10 min)
	w.Header().Set("Cache-Control", "public, max-age=600")
	setETag(w, r, resp)

	respondJSON(w, http.StatusOK, resp)
}

// --- Query param parsing ---

func parseCompanyFilter(r *http.Request) domain.CompanyFilter {
	q := r.URL.Query()

	filter := domain.CompanyFilter{
		Cursor:       q.Get("cursor"),
		Sort:         q.Get("sort"),
		Order:        q.Get("order"),
		Headquarters: strings.TrimSpace(q.Get("headquarters")),
	}

	// Limit
	if l, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = l
	}

	// Size (comma-separated)
	if sizes := q.Get("size"); sizes != "" {
		for _, s := range strings.Split(sizes, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				filter.Sizes = append(filter.Sizes, domain.CompanySize(s))
			}
		}
	}

	// Hiring status
	if hs := q.Get("hiring_status"); hs != "" {
		h := domain.HiringStatus(hs)
		filter.HiringStatus = &h
	}

	// Tech stack (comma-separated, AND match)
	if ts := q.Get("tech_stack"); ts != "" {
		for _, t := range strings.Split(ts, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				filter.TechStack = append(filter.TechStack, t)
			}
		}
	}

	// Domains (comma-separated, OR match)
	if doms := q.Get("domains"); doms != "" {
		for _, d := range strings.Split(doms, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				filter.Domains = append(filter.Domains, d)
			}
		}
	}

	// Compensation tiers (comma-separated)
	if ct := q.Get("compensation_tier"); ct != "" {
		for _, t := range strings.Split(ct, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				filter.CompensationTiers = append(filter.CompensationTiers, t)
			}
		}
	}

	// Has RSU
	if rsu := q.Get("has_rsu"); rsu != "" {
		val := rsu == "true" || rsu == "1"
		filter.HasRSU = &val
	}

	return filter
}

// setETag computes and sets an ETag header. Returns true if the client's
// If-None-Match matches (304 Not Modified should be sent).
func setETag(w http.ResponseWriter, r *http.Request, data any) bool {
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%v", data)
	etag := fmt.Sprintf(`"%x"`, h.Sum(nil)[:8])

	w.Header().Set("ETag", etag)

	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}
