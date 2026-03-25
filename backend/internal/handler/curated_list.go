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

// CuratedListHandler handles curated company list HTTP endpoints.
type CuratedListHandler struct {
	svc *service.CuratedListService
}

// NewCuratedListHandler creates a new CuratedListHandler.
func NewCuratedListHandler(svc *service.CuratedListService) *CuratedListHandler {
	return &CuratedListHandler{svc: svc}
}

// --- Request DTOs ---

type generateListRequest struct {
	ResumeID uuid.UUID `json:"resume_id"`
}

type renameListRequest struct {
	Name string `json:"name"`
}

// --- Response DTO ---

type curatedListResponse struct {
	ID        uuid.UUID       `json:"id"`
	ResumeID  uuid.UUID       `json:"resume_id"`
	Name      *string         `json:"name,omitempty"`
	Result    json.RawMessage `json:"result"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// --- Handlers ---

// GenerateList handles POST /api/curated-lists.
func (h *CuratedListHandler) GenerateList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req generateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}
	if req.ResumeID == uuid.Nil {
		respondError(w, r, domain.ValidationError("resume_id is required", nil))
		return
	}

	list, err := h.svc.GenerateList(r.Context(), userID, req.ResumeID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]any{"data": toCuratedListResponse(list)})
}

// GetList handles GET /api/curated-lists/{id}.
func (h *CuratedListHandler) GetList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid curated list ID", nil))
		return
	}

	list, err := h.svc.GetList(r.Context(), userID, listID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": toCuratedListResponse(list)})
}

// ListByUser handles GET /api/curated-lists/.
func (h *CuratedListHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	lists, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	result := make([]curatedListResponse, 0, len(lists))
	for i := range lists {
		result = append(result, toCuratedListResponse(&lists[i]))
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": result})
}

// RenameList handles PUT /api/curated-lists/{id}.
func (h *CuratedListHandler) RenameList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid curated list ID", nil))
		return
	}

	var req renameListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, domain.ValidationError("Invalid request body", nil))
		return
	}
	if req.Name == "" {
		respondError(w, r, domain.ValidationError("name is required", nil))
		return
	}

	if err := h.svc.Rename(r.Context(), userID, listID, req.Name); err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"message": "ok"})
}

// DeleteList handles DELETE /api/curated-lists/{id}.
func (h *CuratedListHandler) DeleteList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, r, domain.ValidationError("Invalid curated list ID", nil))
		return
	}

	if err := h.svc.Delete(r.Context(), userID, listID); err != nil {
		respondError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"message": "ok"})
}

// --- Converter ---

func toCuratedListResponse(l *domain.CuratedList) curatedListResponse {
	return curatedListResponse{
		ID:        l.ID,
		ResumeID:  l.ResumeID,
		Name:      l.Name,
		Result:    l.Result,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
