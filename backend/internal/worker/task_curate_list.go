package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/skriptvalley/careerdock/internal/ai"
	"github.com/skriptvalley/careerdock/internal/domain"
)

// CurateListHandler processes ai:curate_company_list tasks.
// It builds a candidate profile from the parsed resume, fetches all companies,
// calls the AI to rank the best-fit companies, stores the result, and publishes
// a curated_list_complete SSE notification.
type CurateListHandler struct {
	curatedListRepo domain.CuratedListRepository
	resumeRepo      domain.ResumeRepository
	companyRepo     domain.CompanyRepository
	aiProvider      ai.LLMProvider
	aiCache         *ai.ResultCache
	redisClient     *redis.Client
}

// NewCurateListHandler creates a handler for ai:curate_company_list tasks.
func NewCurateListHandler(
	curatedListRepo domain.CuratedListRepository,
	resumeRepo domain.ResumeRepository,
	companyRepo domain.CompanyRepository,
	aiProvider ai.LLMProvider,
	aiCache *ai.ResultCache,
	redisClient *redis.Client,
) *CurateListHandler {
	return &CurateListHandler{
		curatedListRepo: curatedListRepo,
		resumeRepo:      resumeRepo,
		companyRepo:     companyRepo,
		aiProvider:      aiProvider,
		aiCache:         aiCache,
		redisClient:     redisClient,
	}
}

// Handle processes an ai:curate_company_list task.
//
// Pipeline:
//  1. Load CuratedList from DB
//  2. Load Resume from DB; unmarshal parsed_data
//  3. Fetch all companies (capped at 500)
//  4. Build candidate profile + company summaries
//  5. Check AI cache; call AI if miss
//  6. Cache result
//  7. Update CuratedList.Result in DB
//  8. Publish curated_list_complete SSE event
func (h *CurateListHandler) Handle(ctx context.Context, t *asynq.Task) error {
	listID, err := uuid.Parse(string(t.Payload()))
	if err != nil {
		return fmt.Errorf("invalid list ID in payload: %w", err)
	}

	slog.Info("processing curated company list", "list_id", listID)

	// 1. Load CuratedList
	list, err := h.curatedListRepo.GetByID(ctx, listID)
	if err != nil {
		return fmt.Errorf("load curated list: %w", err)
	}

	// 2. Load Resume and unmarshal parsed_data
	resume, err := h.resumeRepo.GetByID(ctx, list.ResumeID)
	if err != nil {
		return fmt.Errorf("load resume: %w", err)
	}

	if len(resume.ParsedData) == 0 {
		return fmt.Errorf("resume %s has no parsed data", list.ResumeID)
	}

	var parsed ai.ParsedResume
	if err := json.Unmarshal(resume.ParsedData, &parsed); err != nil {
		return fmt.Errorf("unmarshal parsed resume: %w", err)
	}

	// 3. Fetch all companies
	companies, err := h.companyRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list all companies: %w", err)
	}

	// 4. Build company summaries
	summaries := make([]ai.CompanySummary, 0, len(companies))
	for _, c := range companies {
		s := ai.CompanySummary{
			ID:        c.ID,
			Name:      c.Name,
			TechStack: c.TechStack,
			Domains:   c.Domains,
		}
		if c.Headquarters != nil {
			s.Headquarters = *c.Headquarters
		}
		if c.Size != nil {
			s.Size = string(*c.Size)
		}
		if c.CompensationTier != nil {
			s.CompensationTier = *c.CompensationTier
		}
		s.HiringStatus = string(c.HiringStatus)
		summaries = append(summaries, s)
	}

	// 5. Check AI cache
	cacheKey := ai.CacheKeyForCuratedList(list.PreferencesHash)
	result, err := h.curateWithCache(ctx, cacheKey, func() (*ai.CuratedListResult, error) {
		return h.aiProvider.CurateCompanyList(ctx, &ai.CurateListRequest{
			ParsedResume: &parsed,
			Companies:    summaries,
		})
	})
	if err != nil {
		return fmt.Errorf("curate company list: %w", err)
	}

	// 7. Persist result
	resultJSON, err := ai.MarshalCuratedListResult(result)
	if err != nil {
		return fmt.Errorf("marshal curated list result: %w", err)
	}

	if err := h.curatedListRepo.UpdateResult(ctx, listID, resultJSON); err != nil {
		return fmt.Errorf("update curated list result: %w", err)
	}

	slog.Info("curated list complete",
		"list_id", listID,
		"companies_ranked", len(result.Companies),
		"provider", h.aiProvider.Name(),
	)

	// 8. Publish SSE event
	h.publishCuratedListComplete(list.UserID, listID, list.ResumeID, len(result.Companies))

	return nil
}

// curateWithCache checks the AI result cache and calls fn on a miss.
func (h *CurateListHandler) curateWithCache(
	ctx context.Context,
	cacheKey string,
	fn func() (*ai.CuratedListResult, error),
) (*ai.CuratedListResult, error) {
	if h.aiCache != nil {
		if cached, _ := h.aiCache.Get(ctx, "curate_list", cacheKey); cached != nil {
			var result ai.CuratedListResult
			if err := json.Unmarshal(cached, &result); err == nil {
				slog.Info("curated list cache hit")
				return &result, nil
			}
		}
	}

	result, err := fn()
	if err != nil {
		return nil, err
	}

	if h.aiCache != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = h.aiCache.Set(ctx, "curate_list", cacheKey, data, ai.CacheTTLCuratedList)
		}
	}

	return result, nil
}

// publishCuratedListComplete publishes an SSE event for a completed curated list.
func (h *CurateListHandler) publishCuratedListComplete(
	userID, listID, resumeID uuid.UUID,
	companiesRanked int,
) {
	if h.redisClient == nil {
		return
	}

	event := map[string]any{
		"list_id":          listID,
		"resume_id":        resumeID,
		"companies_ranked": companiesRanked,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		slog.Warn("failed to marshal SSE event", "error", err)
		return
	}

	sseEvent := map[string]any{
		"type": "curated_list_complete",
		"data": json.RawMessage(payload),
	}

	eventJSON, err := json.Marshal(sseEvent)
	if err != nil {
		slog.Warn("failed to marshal SSE wrapper", "error", err)
		return
	}

	channel := fmt.Sprintf("sse:user:%s", userID)
	if err := h.redisClient.Publish(context.Background(), channel, eventJSON).Err(); err != nil {
		slog.Warn("failed to publish SSE event", "error", err, "user_id", userID)
	}
}
