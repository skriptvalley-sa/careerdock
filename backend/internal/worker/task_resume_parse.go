// Package worker implements Asynq task handlers for async background jobs.
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

// ResumeParseHandler processes uploaded resumes: parses with AI + runs general ATS scoring.
type ResumeParseHandler struct {
	resumeRepo  domain.ResumeRepository
	fileStore   domain.FileStore
	aiProvider  ai.LLMProvider
	aiCache     *ai.ResultCache
	redisClient *redis.Client
}

// NewResumeParseHandler creates a handler for resume:parse_and_score tasks.
func NewResumeParseHandler(
	resumeRepo domain.ResumeRepository,
	fileStore domain.FileStore,
	aiProvider ai.LLMProvider,
	aiCache *ai.ResultCache,
	redisClient *redis.Client,
) *ResumeParseHandler {
	return &ResumeParseHandler{
		resumeRepo:  resumeRepo,
		fileStore:   fileStore,
		aiProvider:  aiProvider,
		aiCache:     aiCache,
		redisClient: redisClient,
	}
}

// Handle processes a resume:parse_and_score task.
//
// Pipeline:
//  1. Load resume from DB
//  2. Parse resume text with AI (check cache first)
//  3. Run general ATS scoring with AI (check cache first)
//  4. Update resume record with parsed_data + ats_general, status → "ready"
func (h *ResumeParseHandler) Handle(ctx context.Context, t *asynq.Task) error {
	resumeID, err := uuid.Parse(string(t.Payload()))
	if err != nil {
		return fmt.Errorf("invalid resume ID in payload: %w", err)
	}

	slog.Info("processing resume parse+score", "resume_id", resumeID)

	// 1. Load resume
	resume, err := h.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return fmt.Errorf("load resume: %w", err)
	}

	resumeText := ""
	if resume.ExtractedText != nil {
		resumeText = *resume.ExtractedText
	}

	if resumeText == "" {
		slog.Warn("resume has no extracted text — marking as failed", "resume_id", resumeID)
		resume.Status = domain.ResumeStatusFailed
		if err := h.resumeRepo.Update(ctx, resume); err != nil {
			return fmt.Errorf("update resume status: %w", err)
		}
		return nil // don't retry — no text to parse
	}

	// 2. Parse resume with AI
	parsedData, err := h.parseResume(ctx, resumeText)
	if err != nil {
		slog.Error("AI resume parsing failed", "resume_id", resumeID, "error", err)
		resume.Status = domain.ResumeStatusFailed
		_ = h.resumeRepo.Update(ctx, resume)
		return fmt.Errorf("parse resume: %w", err) // will be retried by Asynq
	}

	parsedJSON, err := json.Marshal(parsedData)
	if err != nil {
		return fmt.Errorf("marshal parsed data: %w", err)
	}

	// 3. Run general ATS scoring
	var pdfBytes []byte
	pdfBytes, err = h.fileStore.Download(ctx, resume.S3Key)
	if err != nil {
		slog.Warn("failed to download PDF for ATS scoring — using text only",
			"resume_id", resumeID, "error", err)
	}

	atsResult, err := h.scoreATSGeneral(ctx, resumeText, pdfBytes)
	if err != nil {
		slog.Error("AI ATS scoring failed — saving parse results only",
			"resume_id", resumeID, "error", err)
		// Still save parsed data even if ATS scoring fails
		resume.ParsedData = parsedJSON
		resume.Status = domain.ResumeStatusReady
		if err := h.resumeRepo.Update(ctx, resume); err != nil {
			return fmt.Errorf("update resume: %w", err)
		}
		h.publishResumeReady(resume.UserID, resumeID, resume.FileName, 0)
		return nil // don't retry — partial success
	}

	atsJSON, err := json.Marshal(atsResult)
	if err != nil {
		return fmt.Errorf("marshal ATS result: %w", err)
	}

	// 4. Update resume with both results
	resume.ParsedData = parsedJSON
	resume.ATSGeneral = atsJSON
	resume.Status = domain.ResumeStatusReady

	if err := h.resumeRepo.Update(ctx, resume); err != nil {
		return fmt.Errorf("update resume: %w", err)
	}

	slog.Info("resume parse+score complete",
		"resume_id", resumeID,
		"ats_score", atsResult.Score,
		"provider", h.aiProvider.Name(),
	)

	// Publish SSE event to notify connected client
	h.publishResumeReady(resume.UserID, resumeID, resume.FileName, atsResult.Score)

	return nil
}

// parseResume checks cache then calls AI provider for resume parsing.
func (h *ResumeParseHandler) parseResume(ctx context.Context, resumeText string) (*ai.ParsedResume, error) {
	// Check cache
	if h.aiCache != nil {
		cached, _ := h.aiCache.Get(ctx, "resume_parse", ai.CacheKeyForResumeParse(resumeText))
		if cached != nil {
			var result ai.ParsedResume
			if err := json.Unmarshal(cached, &result); err == nil {
				slog.Info("resume parse cache hit")
				return &result, nil
			}
		}
	}

	// Call AI provider
	result, err := h.aiProvider.ParseResume(ctx, &ai.ParseResumeRequest{
		ResumeText: resumeText,
	})
	if err != nil {
		return nil, err
	}

	// Cache result
	if h.aiCache != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = h.aiCache.Set(ctx, "resume_parse", ai.CacheKeyForResumeParse(resumeText), data, ai.CacheTTLResumeParse)
		}
	}

	return result, nil
}

// scoreATSGeneral checks cache then calls AI provider for general ATS scoring.
func (h *ResumeParseHandler) scoreATSGeneral(ctx context.Context, resumeText string, pdfBytes []byte) (*ai.ATSResult, error) {
	// Check cache
	if h.aiCache != nil {
		cached, _ := h.aiCache.Get(ctx, "ats_general", ai.CacheKeyForATSGeneral(resumeText))
		if cached != nil {
			var result ai.ATSResult
			if err := json.Unmarshal(cached, &result); err == nil {
				slog.Info("ATS general cache hit")
				return &result, nil
			}
		}
	}

	// Call AI provider
	result, err := h.aiProvider.ScoreATSGeneral(ctx, &ai.ATSGeneralRequest{
		PDFBytes:   pdfBytes,
		ResumeText: resumeText,
	})
	if err != nil {
		return nil, err
	}

	// Cache result
	if h.aiCache != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = h.aiCache.Set(ctx, "ats_general", ai.CacheKeyForATSGeneral(resumeText), data, ai.CacheTTLATSGeneral)
		}
	}

	return result, nil
}

// publishResumeReady publishes an SSE event to notify the user their resume is ready.
func (h *ResumeParseHandler) publishResumeReady(userID, resumeID uuid.UUID, fileName string, atsScore int) {
	if h.redisClient == nil {
		return
	}

	event := map[string]any{
		"resume_id": resumeID,
		"file_name": fileName,
		"status":    "ready",
		"ats_score": atsScore,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		slog.Warn("failed to marshal SSE event", "error", err)
		return
	}

	sseEvent := map[string]any{
		"type": "resume_ready",
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
