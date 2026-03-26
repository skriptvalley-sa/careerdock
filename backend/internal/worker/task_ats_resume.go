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
	"github.com/skriptvalley/careerdock/internal/pdf"
)

// ATSResumeHandler processes ats:resume_check tasks.
// It downloads the resume PDF, calls the AI for general ATS scoring,
// stores the result, and publishes an SSE notification.
type ATSResumeHandler struct {
	atsRepo     domain.ATSCheckRepository
	resumeRepo  domain.ResumeRepository
	fileStore   domain.FileStore
	aiProvider  ai.LLMProvider
	aiCache     *ai.ResultCache
	redisClient *redis.Client
}

// NewATSResumeHandler creates a handler for ats:resume_check tasks.
func NewATSResumeHandler(
	atsRepo domain.ATSCheckRepository,
	resumeRepo domain.ResumeRepository,
	fileStore domain.FileStore,
	aiProvider ai.LLMProvider,
	aiCache *ai.ResultCache,
	redisClient *redis.Client,
) *ATSResumeHandler {
	return &ATSResumeHandler{
		atsRepo:     atsRepo,
		resumeRepo:  resumeRepo,
		fileStore:   fileStore,
		aiProvider:  aiProvider,
		aiCache:     aiCache,
		redisClient: redisClient,
	}
}

// Handle processes an ats:resume_check task.
//
// Pipeline:
//  1. Load ATSCheck from DB
//     2a. If ResumeID set: load Resume from DB, download PDF by S3Key
//     2b. If ResumeID nil: download temp PDF directly from TempS3Key
//  3. Check AI cache; call ScoreATSGeneral if miss
//  4. Update ATSCheck.Result in DB
//  5. Delete temp S3 file (if applicable)
//  6. Publish ats_resume_complete SSE event
func (h *ATSResumeHandler) Handle(ctx context.Context, t *asynq.Task) error {
	checkID, err := uuid.Parse(string(t.Payload()))
	if err != nil {
		return fmt.Errorf("invalid check ID in payload: %w", err)
	}

	slog.Info("processing ATS resume check", "check_id", checkID)

	// 1. Load ATSCheck
	check, err := h.atsRepo.GetByID(ctx, checkID)
	if err != nil {
		return fmt.Errorf("load ATS check: %w", err)
	}

	var resumeText string
	var pdfBytes []byte
	var tempKeyToDelete string

	switch {
	case check.ResumeID != nil:
		// 2a. Slot-based resume path
		resume, loadErr := h.resumeRepo.GetByID(ctx, *check.ResumeID)
		if loadErr != nil {
			return fmt.Errorf("load resume: %w", loadErr)
		}
		if resume.ExtractedText != nil {
			resumeText = *resume.ExtractedText
		}
		if dlBytes, dlErr := h.fileStore.Download(ctx, resume.S3Key); dlErr != nil {
			slog.Warn("failed to download PDF for resume ATS — using text only",
				"check_id", checkID, "error", dlErr)
		} else {
			pdfBytes = dlBytes
		}
	case check.TempS3Key != nil:
		// 2b. Temp upload path — download the PDF directly
		tempKeyToDelete = *check.TempS3Key
		pdfBytes, err = h.fileStore.Download(ctx, tempKeyToDelete)
		if err != nil {
			return fmt.Errorf("download temp PDF: %w", err)
		}
		// Extract text from downloaded PDF bytes
		if text, extractErr := pdf.ExtractText(pdfBytes); extractErr == nil {
			resumeText = text
		} else {
			slog.Warn("text extraction failed for temp PDF — scoring with PDF bytes only",
				"check_id", checkID, "error", extractErr)
		}
	default:
		return fmt.Errorf("ats check %s has neither resume_id nor temp_s3_key", checkID)
	}

	// 3. Check AI cache + call AI
	cacheKey := ai.CacheKeyForATSGeneral(resumeText)

	result, err := h.scoreWithCache(ctx, cacheKey, func() (*ai.ATSResult, error) {
		return h.aiProvider.ScoreATSGeneral(ctx, &ai.ATSGeneralRequest{
			PDFBytes:   pdfBytes,
			ResumeText: resumeText,
		})
	}, ai.CacheTTLATSGeneral)
	if err != nil {
		return fmt.Errorf("score ATS resume: %w", err)
	}

	// 4. Persist result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal ATS result: %w", err)
	}

	if err := h.atsRepo.UpdateResult(ctx, checkID, resultJSON); err != nil {
		return fmt.Errorf("update ATS check result: %w", err)
	}

	slog.Info("ATS resume check complete",
		"check_id", checkID,
		"score", result.Score,
		"provider", h.aiProvider.Name(),
	)

	// 5. Clean up temp file (best-effort — don't fail the task if delete errors)
	if tempKeyToDelete != "" {
		if delErr := h.fileStore.Delete(ctx, tempKeyToDelete); delErr != nil {
			slog.Warn("failed to delete temp ATS PDF", "key", tempKeyToDelete, "error", delErr)
		}
	}

	// 6. Publish SSE event
	h.publishATSComplete(check.UserID, checkID, check.ResumeID, "ats_resume_complete", result.Score)

	return nil
}

// scoreWithCache checks the AI result cache and calls fn on a miss.
func (h *ATSResumeHandler) scoreWithCache(
	ctx context.Context,
	cacheKey string,
	fn func() (*ai.ATSResult, error),
	ttl time.Duration,
) (*ai.ATSResult, error) {
	if h.aiCache != nil {
		if cached, _ := h.aiCache.Get(ctx, "ats_general", cacheKey); cached != nil {
			var result ai.ATSResult
			if err := json.Unmarshal(cached, &result); err == nil {
				slog.Info("ATS resume cache hit")
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
			_ = h.aiCache.Set(ctx, "ats_general", cacheKey, data, ttl)
		}
	}

	return result, nil
}

// publishATSComplete publishes an SSE event for a completed ATS resume check.
func (h *ATSResumeHandler) publishATSComplete(
	userID, checkID uuid.UUID,
	resumeID *uuid.UUID,
	eventType string,
	score int,
) {
	if h.redisClient == nil {
		return
	}

	event := map[string]any{
		"check_id": checkID,
		"score":    score,
	}
	if resumeID != nil {
		event["resume_id"] = resumeID
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
