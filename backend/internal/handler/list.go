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

// ListHandler handles list and entry HTTP endpoints.
type ListHandler struct {
	lists     *service.ListService
	companies *service.CompanyService
}

// NewListHandler creates a new ListHandler.
func NewListHandler(lists *service.ListService, companies *service.CompanyService) *ListHandler {
	return &ListHandler{lists: lists, companies: companies}
}

// --- Request DTOs ---

type createListRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateListRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Position    *int    `json:"position"`
}

type createEntryRequest struct {
	CompanyID     string  `json:"company_id"`
	CompanyStatus *string `json:"company_status"`
	RoleTitle     *string `json:"role_title"`
	Status        *string `json:"status"`
	DateApplied   *string `json:"date_applied"`
	Notes         *string `json:"notes"`
}

type updateEntryRequest struct {
	CompanyStatus *string `json:"company_status"`
	Status        *string `json:"status"`
	RoleTitle     *string `json:"role_title"`
	Notes         *string `json:"notes"`
	DateApplied   *string `json:"date_applied"`
	Position      *int    `json:"position"`
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

type listResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Position    int     `json:"position"`
	EntryCount  int     `json:"entry_count"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type entryResponse struct {
	ID            string  `json:"id"`
	ListID        string  `json:"list_id"`
	CompanyID     string  `json:"company_id"`
	CompanyName   string  `json:"company_name"`
	CompanySlug   string  `json:"company_slug,omitempty"`
	CompanyStatus string  `json:"company_status"`
	RoleTitle     *string `json:"role_title,omitempty"`
	Status        string  `json:"status"`
	DateApplied   *string `json:"date_applied,omitempty"`
	Notes         *string `json:"notes,omitempty"`
	Position      int     `json:"position"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
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

func toListResponse(l *domain.UserList, entryCount int) listResponse {
	return listResponse{
		ID:          l.ID.String(),
		Name:        l.Name,
		Description: l.Description,
		Position:    l.Position,
		EntryCount:  entryCount,
		CreatedAt:   l.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   l.UpdatedAt.Format(time.RFC3339),
	}
}

func toEntryResponse(e *domain.ListEntry) entryResponse {
	resp := entryResponse{
		ID:            e.ID.String(),
		ListID:        e.ListID.String(),
		CompanyID:     e.CompanyID.String(),
		CompanyStatus: string(e.CompanyStatus),
		RoleTitle:     e.RoleTitle,
		Status:        string(e.Status),
		Notes:         e.Notes,
		Position:      e.Position,
		CreatedAt:     e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     e.UpdatedAt.Format(time.RFC3339),
	}
	if e.DateApplied != nil {
		s := e.DateApplied.Format("2006-01-02")
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

// --- List handlers ---

// ListLists handles GET /api/lists.
func (h *ListHandler) ListLists(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	lists, err := h.lists.ListsByUser(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	// Build responses with entry counts
	items := make([]listResponse, len(lists))
	for i, l := range lists {
		entries, err := h.lists.ListEntries(r.Context(), l.ID, userID)
		if err != nil {
			respondError(w, r, err)
			return
		}
		items[i] = toListResponse(&l, len(entries))
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: items})
}

// CreateList handles POST /api/lists.
func (h *ListHandler) CreateList(w http.ResponseWriter, r *http.Request) {
	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondError(w, r, domain.ValidationError("name is required", map[string]any{"field": "name"}))
		return
	}
	if len(req.Name) > 255 {
		respondError(w, r, domain.ValidationError("name must be at most 255 characters", map[string]any{"field": "name"}))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	list, err := h.lists.CreateList(r.Context(), service.CreateListInput{
		UserID:      userID,
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, DataResponse{Data: toListResponse(list, 0)})
}

// GetList handles GET /api/lists/{id}.
func (h *ListHandler) GetList(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	list, err := h.lists.GetList(r.Context(), listID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	entries, err := h.lists.ListEntries(r.Context(), listID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	// Batch-resolve company names and slugs for all entries
	companyIDs := make([]uuid.UUID, 0, len(entries))
	seen := make(map[uuid.UUID]bool)
	for _, e := range entries {
		if !seen[e.CompanyID] {
			companyIDs = append(companyIDs, e.CompanyID)
			seen[e.CompanyID] = true
		}
	}
	companyInfo, _ := h.companies.GetNameAndSlugsByIDs(r.Context(), companyIDs)
	if companyInfo == nil {
		companyInfo = map[uuid.UUID]domain.CompanyNameSlug{}
	}

	entryItems := make([]entryResponse, len(entries))
	for i, e := range entries {
		resp := toEntryResponse(&e)
		if info, ok := companyInfo[e.CompanyID]; ok {
			resp.CompanyName = info.Name
			resp.CompanySlug = info.Slug
		}
		entryItems[i] = resp
	}

	type listDetailResponse struct {
		listResponse
		Entries []entryResponse `json:"entries"`
	}

	resp := listDetailResponse{
		listResponse: toListResponse(list, len(entries)),
		Entries:      entryItems,
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: resp})
}

// UpdateList handles PUT /api/lists/{id}.
func (h *ListHandler) UpdateList(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	var req updateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			respondError(w, r, domain.ValidationError("name cannot be empty", map[string]any{"field": "name"}))
			return
		}
		req.Name = &trimmed
	}

	userID := middleware.UserIDFromContext(r.Context())
	list, err := h.lists.UpdateList(r.Context(), service.UpdateListInput{
		ListID:      listID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Position:    req.Position,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: toListResponse(list, 0)})
}

// DeleteList handles DELETE /api/lists/{id}.
func (h *ListHandler) DeleteList(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.lists.DeleteList(r.Context(), listID, userID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Entry handlers ---

// CreateEntry handles POST /api/lists/{id}/entries.
func (h *ListHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	var req createEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	companyID, parseErr := uuid.Parse(req.CompanyID)
	if parseErr != nil {
		respondError(w, r, domain.ValidationError("invalid company_id", map[string]any{"field": "company_id"}))
		return
	}

	// RoleTitle is now optional — trim if provided
	var roleTitle *string
	if req.RoleTitle != nil {
		trimmed := strings.TrimSpace(*req.RoleTitle)
		if trimmed != "" {
			roleTitle = &trimmed
		}
	}

	var companyStatus domain.CompanyTrackingStatus
	if req.CompanyStatus != nil {
		companyStatus = domain.CompanyTrackingStatus(*req.CompanyStatus)
	}

	var status domain.ApplicationStatus
	if req.Status != nil {
		status = domain.ApplicationStatus(*req.Status)
	}

	var dateApplied *time.Time
	if req.DateApplied != nil {
		t, parseErr := time.Parse("2006-01-02", *req.DateApplied)
		if parseErr != nil {
			respondError(w, r, domain.ValidationError("invalid date_applied format (expected YYYY-MM-DD)", map[string]any{"field": "date_applied"}))
			return
		}
		dateApplied = &t
	}

	userID := middleware.UserIDFromContext(r.Context())
	entry, err := h.lists.CreateEntry(r.Context(), service.CreateEntryInput{
		ListID:        listID,
		UserID:        userID,
		CompanyID:     companyID,
		CompanyStatus: companyStatus,
		RoleTitle:     roleTitle,
		Status:        status,
		DateApplied:   dateApplied,
		Notes:         req.Notes,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, DataResponse{Data: toEntryResponse(entry)})
}

// BatchCreateEntries handles POST /api/lists/{id}/entries/batch.
// Accepts { "company_ids": ["uuid", ...] } and creates one entry per company
// with default company_status "marked".
func (h *ListHandler) BatchCreateEntries(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	var req struct {
		CompanyIDs []string `json:"company_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}
	if len(req.CompanyIDs) == 0 {
		respondError(w, r, domain.ValidationError("company_ids is required", nil))
		return
	}
	if len(req.CompanyIDs) > 50 {
		respondError(w, r, domain.ValidationError("cannot add more than 50 companies at once", nil))
		return
	}

	companyIDs := make([]uuid.UUID, 0, len(req.CompanyIDs))
	for _, raw := range req.CompanyIDs {
		cid, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			respondError(w, r, domain.ValidationError("invalid company_id: "+raw, nil))
			return
		}
		companyIDs = append(companyIDs, cid)
	}

	userID := middleware.UserIDFromContext(r.Context())
	entries, err := h.lists.BatchCreateEntries(r.Context(), service.BatchCreateEntriesInput{
		ListID:     listID,
		UserID:     userID,
		CompanyIDs: companyIDs,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	// Resolve company names
	companyNames, _ := h.companies.GetNamesByIDs(r.Context(), companyIDs)
	if companyNames == nil {
		companyNames = map[uuid.UUID]string{}
	}

	items := make([]entryResponse, len(entries))
	for i, e := range entries {
		resp := toEntryResponse(e)
		resp.CompanyName = companyNames[e.CompanyID]
		items[i] = resp
	}

	respondJSON(w, http.StatusCreated, DataResponse{Data: items})
}

// UpdateEntry handles PUT /api/lists/{id}/entries/{entryId}.
func (h *ListHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	_, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	entryID, err := parseUUIDParam(r, "entryId")
	if err != nil {
		respondError(w, r, err)
		return
	}

	var req updateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	var companyStatus *domain.CompanyTrackingStatus
	if req.CompanyStatus != nil {
		cs := domain.CompanyTrackingStatus(*req.CompanyStatus)
		companyStatus = &cs
	}

	var status *domain.ApplicationStatus
	if req.Status != nil {
		s := domain.ApplicationStatus(*req.Status)
		status = &s
	}

	var dateApplied *time.Time
	if req.DateApplied != nil {
		t, parseErr := time.Parse("2006-01-02", *req.DateApplied)
		if parseErr != nil {
			respondError(w, r, domain.ValidationError("invalid date_applied format (expected YYYY-MM-DD)", map[string]any{"field": "date_applied"}))
			return
		}
		dateApplied = &t
	}

	userID := middleware.UserIDFromContext(r.Context())
	entry, err := h.lists.UpdateEntry(r.Context(), service.UpdateEntryInput{
		EntryID:       entryID,
		UserID:        userID,
		CompanyStatus: companyStatus,
		Status:        status,
		RoleTitle:     req.RoleTitle,
		Notes:         req.Notes,
		DateApplied:   dateApplied,
		Position:      req.Position,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: toEntryResponse(entry)})
}

// DeleteEntry handles DELETE /api/lists/{id}/entries/{entryId}.
func (h *ListHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	_, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	entryID, err := parseUUIDParam(r, "entryId")
	if err != nil {
		respondError(w, r, err)
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.lists.DeleteEntry(r.Context(), entryID, userID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEntryHistory handles GET /api/lists/{id}/entries/{entryId}/history.
func (h *ListHandler) GetEntryHistory(w http.ResponseWriter, r *http.Request) {
	_, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	entryID, err := parseUUIDParam(r, "entryId")
	if err != nil {
		respondError(w, r, err)
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	history, err := h.lists.GetEntryHistory(r.Context(), entryID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	items := make([]statusHistoryResponse, len(history))
	for i, h := range history {
		items[i] = toStatusHistoryResponse(h)
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: items})
}

// --- Interview Round handlers ---

// CreateRound handles POST /api/lists/{id}/entries/{entryId}/rounds.
func (h *ListHandler) CreateRound(w http.ResponseWriter, r *http.Request) {
	_, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	entryID, err := parseUUIDParam(r, "entryId")
	if err != nil {
		respondError(w, r, err)
		return
	}

	var req createRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}

	req.RoundType = strings.TrimSpace(req.RoundType)
	if req.RoundType == "" {
		respondError(w, r, domain.ValidationError("round_type is required", map[string]any{"field": "round_type"}))
		return
	}
	if req.RoundNumber < 1 {
		respondError(w, r, domain.ValidationError("round_number must be positive", map[string]any{"field": "round_number"}))
		return
	}

	var scheduledDate *time.Time
	if req.ScheduledDate != nil {
		t, parseErr := time.Parse("2006-01-02", *req.ScheduledDate)
		if parseErr != nil {
			respondError(w, r, domain.ValidationError("invalid scheduled_date format (expected YYYY-MM-DD)", map[string]any{"field": "scheduled_date"}))
			return
		}
		scheduledDate = &t
	}

	var outcome domain.InterviewOutcome
	if req.Outcome != nil {
		outcome = domain.InterviewOutcome(*req.Outcome)
	}

	userID := middleware.UserIDFromContext(r.Context())
	round, err := h.lists.CreateRound(r.Context(), service.CreateRoundInput{
		EntryID:       entryID,
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

// UpdateRound handles PUT /api/lists/{id}/entries/{entryId}/rounds/{roundId}.
func (h *ListHandler) UpdateRound(w http.ResponseWriter, r *http.Request) {
	roundID, err := parseUUIDParam(r, "roundId")
	if err != nil {
		respondError(w, r, err)
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
		t, parseErr := time.Parse("2006-01-02", *req.ScheduledDate)
		if parseErr != nil {
			respondError(w, r, domain.ValidationError("invalid scheduled_date format (expected YYYY-MM-DD)", map[string]any{"field": "scheduled_date"}))
			return
		}
		scheduledDate = &t
	}

	userID := middleware.UserIDFromContext(r.Context())
	round, err := h.lists.UpdateRound(r.Context(), service.UpdateRoundInput{
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

// DeleteRound handles DELETE /api/lists/{id}/entries/{entryId}/rounds/{roundId}.
func (h *ListHandler) DeleteRound(w http.ResponseWriter, r *http.Request) {
	roundID, err := parseUUIDParam(r, "roundId")
	if err != nil {
		respondError(w, r, err)
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.lists.DeleteRound(r.Context(), roundID, userID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDashboard handles GET /api/dashboard.
func (h *ListHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	counts, err := h.lists.GetDashboardCounts(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: counts})
}

// ListEntriesAcrossLists handles GET /api/entries
// Supports two modes:
//   - ?company_id=UUID  → entries for a specific company across all user lists
//   - (no company_id)   → all entries across all lists, optionally ?status=... filtered
func (h *ListHandler) ListEntriesAcrossLists(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	type fullEntryResponse struct {
		entryResponse
		ListName    string `json:"list_name"`
		CompanyName string `json:"company_name"`
	}

	companyIDStr := r.URL.Query().Get("company_id")

	// Mode 1: filter by company
	if companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err != nil {
			respondError(w, r, domain.ValidationError("invalid company_id", nil))
			return
		}

		entries, err := h.lists.ListEntriesByCompanyID(r.Context(), userID, companyID)
		if err != nil {
			respondError(w, r, err)
			return
		}

		items := make([]fullEntryResponse, len(entries))
		for i, e := range entries {
			resp := toEntryResponse(&e.ListEntry)
			items[i] = fullEntryResponse{
				entryResponse: resp,
				ListName:      e.ListName,
			}
		}
		respondJSON(w, http.StatusOK, DataResponse{Data: items})
		return
	}

	// Mode 2: all entries, optional status filter
	var statusFilter *domain.ApplicationStatus
	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.ApplicationStatus(s)
		statusFilter = &status
	}

	entries, err := h.lists.ListAllEntries(r.Context(), userID, statusFilter, true)
	if err != nil {
		respondError(w, r, err)
		return
	}

	items := make([]fullEntryResponse, len(entries))
	for i, e := range entries {
		resp := toEntryResponse(&e.ListEntry)
		items[i] = fullEntryResponse{
			entryResponse: resp,
			ListName:      e.ListName,
			CompanyName:   e.CompanyName,
		}
	}
	respondJSON(w, http.StatusOK, DataResponse{Data: items})
}

// SyncListEntries handles PUT /api/lists/{id}/entries/sync.
// Receives the full desired set of company IDs and reconciles additions + removals.
func (h *ListHandler) SyncListEntries(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	var req struct {
		CompanyIDs []string `json:"company_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("invalid request body", nil))
		return
	}
	if len(req.CompanyIDs) > 50 {
		respondError(w, r, domain.ValidationError("cannot sync more than 50 companies at once", nil))
		return
	}

	companyIDs := make([]uuid.UUID, 0, len(req.CompanyIDs))
	for _, raw := range req.CompanyIDs {
		cid, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			respondError(w, r, domain.ValidationError("invalid company_id: "+raw, nil))
			return
		}
		companyIDs = append(companyIDs, cid)
	}

	userID := middleware.UserIDFromContext(r.Context())
	_, err = h.lists.SyncListEntries(r.Context(), service.SyncListEntriesInput{
		ListID:     listID,
		UserID:     userID,
		CompanyIDs: companyIDs,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	// Return the full updated entry list with company names
	entries, listErr := h.lists.ListEntries(r.Context(), listID, userID)
	if listErr != nil {
		respondError(w, r, listErr)
		return
	}

	cids := make([]uuid.UUID, 0, len(entries))
	seen := make(map[uuid.UUID]bool)
	for _, e := range entries {
		if !seen[e.CompanyID] {
			cids = append(cids, e.CompanyID)
			seen[e.CompanyID] = true
		}
	}
	companyInfo, _ := h.companies.GetNameAndSlugsByIDs(r.Context(), cids)
	if companyInfo == nil {
		companyInfo = map[uuid.UUID]domain.CompanyNameSlug{}
	}

	items := make([]entryResponse, len(entries))
	for i, e := range entries {
		resp := toEntryResponse(&e)
		if info, ok := companyInfo[e.CompanyID]; ok {
			resp.CompanyName = info.Name
			resp.CompanySlug = info.Slug
		}
		items[i] = resp
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: items})
}

// ListsForCompany handles GET /api/lists/by-company/{companyId}.
// Returns the user's lists with a flag indicating whether each list contains the company.
func (h *ListHandler) ListsForCompany(w http.ResponseWriter, r *http.Request) {
	companyIDStr := chi.URLParam(r, "companyId")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		respondError(w, r, domain.ValidationError("invalid companyId", nil))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	flags, err := h.lists.ListsWithCompanyFlag(r.Context(), userID, companyID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: flags})
}

// CompanyListCounts handles GET /api/lists/company-counts.
// Returns a map of company_id → number of lists containing it for the authenticated user.
func (h *ListHandler) CompanyListCounts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	counts, err := h.lists.CompanyListCounts(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	// Convert uuid keys to strings for JSON serialization
	result := make(map[string]int, len(counts))
	for k, v := range counts {
		result[k.String()] = v
	}

	respondJSON(w, http.StatusOK, DataResponse{Data: result})
}

// DeleteEntryByCompany handles DELETE /api/lists/{id}/entries/by-company/{companyId}.
// Removes an entry by list ID and company ID.
func (h *ListHandler) DeleteEntryByCompany(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, r, err)
		return
	}

	companyIDStr := chi.URLParam(r, "companyId")
	companyID, parseErr := uuid.Parse(companyIDStr)
	if parseErr != nil {
		respondError(w, r, domain.ValidationError("invalid companyId", nil))
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.lists.DeleteEntryByCompany(r.Context(), listID, companyID, userID); err != nil {
		respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Helpers ---

// parseUUIDParam extracts a UUID from a chi URL param.
func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	raw := chi.URLParam(r, name)
	if raw == "" {
		return uuid.Nil, domain.ValidationError(name+" is required", nil)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, domain.ValidationError("invalid "+name+" format", nil)
	}
	return id, nil
}
