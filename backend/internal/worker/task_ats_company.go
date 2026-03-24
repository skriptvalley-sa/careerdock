package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/skriptvalley/careerdock/internal/ai"
	"github.com/skriptvalley/careerdock/internal/domain"
)

// ATSCompanyHandler processes ats:company_check tasks.
// It downloads the resume PDF, fetches the company profile, calls the AI for scoring,
// stores the result, and publishes an SSE notification.
type ATSCompanyHandler struct {
	atsRepo     domain.ATSCheckRepository
	resumeRepo  domain.ResumeRepository
	companyRepo domain.CompanyRepository
	fileStore   domain.FileStore
	aiProvider  ai.LLMProvider
	aiCache     *ai.ResultCache
	redisClient *redis.Client
}

// NewATSCompanyHandler creates a handler for ats:company_check tasks.
func NewATSCompanyHandler(
	atsRepo domain.ATSCheckRepository,
	resumeRepo domain.ResumeRepository,
	companyRepo domain.CompanyRepository,
	fileStore domain.FileStore,
	aiProvider ai.LLMProvider,
	aiCache *ai.ResultCache,
	redisClient *redis.Client,
) *ATSCompanyHandler {
	return &ATSCompanyHandler{
		atsRepo:     atsRepo,
		resumeRepo:  resumeRepo,
		companyRepo: companyRepo,
		fileStore:   fileStore,
		aiProvider:  aiProvider,
		aiCache:     aiCache,
		redisClient: redisClient,
	}
}

// Handle processes an ats:company_check task.
//
// Pipeline:
//  1. Load ATSCheck from DB
//  2. Load Resume from DB
//  3. Download PDF from S3 (degrade to text-only on failure)
//  4. Load Company profile
//  5. Check AI cache; call AI if miss
//  6. Cache result
//  7. Update ATSCheck.Result in DB
//  8. Publish ats_company_complete SSE event
func (h *ATSCompanyHandler) Handle(ctx context.Context, t *asynq.Task) error {
	checkID, err := uuid.Parse(string(t.Payload()))
	if err != nil {
		return fmt.Errorf("invalid check ID in payload: %w", err)
	}

	slog.Info("processing ATS company check", "check_id", checkID)

	// 1. Load ATSCheck
	check, err := h.atsRepo.GetByID(ctx, checkID)
	if err != nil {
		return fmt.Errorf("load ATS check: %w", err)
	}
	if check.CompanyID == nil {
		return fmt.Errorf("ATS check %s has no company_id", checkID)
	}

	// 2. Load Resume
	resume, err := h.resumeRepo.GetByID(ctx, check.ResumeID)
	if err != nil {
		return fmt.Errorf("load resume: %w", err)
	}

	resumeText := ""
	if resume.ExtractedText != nil {
		resumeText = *resume.ExtractedText
	}

	// 3. Download PDF (best-effort)
	var pdfBytes []byte
	pdfBytes, err = h.fileStore.Download(ctx, resume.S3Key)
	if err != nil {
		slog.Warn("failed to download PDF for company ATS — using text only",
			"check_id", checkID, "error", err)
	}

	// 4. Load Company
	company, err := h.companyRepo.GetByID(ctx, *check.CompanyID)
	if err != nil {
		return fmt.Errorf("load company: %w", err)
	}

	companyProfile := &ai.CompanyProfile{
		Name:      company.Name,
		TechStack: company.TechStack,
		Domains:   company.Domains,
	}
	if company.Size != nil {
		companyProfile.Size = string(*company.Size)
	}
	if company.CompensationTier != nil {
		companyProfile.CompensationTier = *company.CompensationTier
	}

	// 5. Check AI cache
	cacheKey := ai.CacheKeyForATSCompany(resumeText, check.CompanyID.String())
	result, err := h.scoreWithCache(ctx, cacheKey, func() (*ai.ATSResult, error) {
		return h.aiProvider.ScoreATSCompany(ctx, &ai.ATSCompanyRequest{
			PDFBytes:   pdfBytes,
			ResumeText: resumeText,
			Company:    companyProfile,
		})
	}, ai.CacheTTLATSCompany)
	if err != nil {
		return fmt.Errorf("score ATS company: %w", err)
	}

	// 7. Persist result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal ATS result: %w", err)
	}

	if err := h.atsRepo.UpdateResult(ctx, checkID, resultJSON); err != nil {
		return fmt.Errorf("update ATS check result: %w", err)
	}

	slog.Info("ATS company check complete",
		"check_id", checkID,
		"score", result.Score,
		"provider", h.aiProvider.Name(),
	)

	// 8. Publish SSE event
	h.publishATSComplete(check.UserID, checkID, check.ResumeID, "ats_company_complete", result.Score)

	return nil
}

// scoreWithCache checks the AI result cache and calls fn on a miss.
func (h *ATSCompanyHandler) scoreWithCache(
	ctx context.Context,
	cacheKey string,
	fn func() (*ai.ATSResult, error),
	ttl time.Duration,
) (*ai.ATSResult, error) {
	if h.aiCache != nil {
		if cached, _ := h.aiCache.Get(ctx, "ats_company", cacheKey); cached != nil {
			var result ai.ATSResult
			if err := json.Unmarshal(cached, &result); err == nil {
				slog.Info("ATS company cache hit")
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
			_ = h.aiCache.Set(ctx, "ats_company", cacheKey, data, ttl)
		}
	}

	return result, nil
}

// publishATSComplete publishes an SSE event for a completed ATS check.
func (h *ATSCompanyHandler) publishATSComplete(
	userID, checkID, resumeID uuid.UUID,
	eventType string,
	score int,
) {
	if h.redisClient == nil {
		return
	}

	event := map[string]any{
		"check_id":  checkID,
		"resume_id": resumeID,
		"score":     score,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		slog.Warn("failed to marshal SSE event", "error", err)
		return
	}

	sseEvent := map[string]any{
		"type": eventType,
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
