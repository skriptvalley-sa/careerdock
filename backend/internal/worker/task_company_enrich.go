package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/skriptvalley/careerdock/internal/ai"
	"github.com/skriptvalley/careerdock/internal/domain"
)

// CompanyEnrichPayload is the JSON payload for a company enrichment task.
type CompanyEnrichPayload struct {
	CompanyID string `json:"company_id"`
}

// CompanyEnrichHandler enriches a single company's profile using AI.
type CompanyEnrichHandler struct {
	companyRepo domain.CompanyRepository
	aiProvider  ai.LLMProvider
}

// NewCompanyEnrichHandler creates a handler for admin:company_enrich tasks.
func NewCompanyEnrichHandler(
	companyRepo domain.CompanyRepository,
	aiProvider ai.LLMProvider,
) *CompanyEnrichHandler {
	return &CompanyEnrichHandler{
		companyRepo: companyRepo,
		aiProvider:  aiProvider,
	}
}

// Handle processes a company enrichment task.
func (h *CompanyEnrichHandler) Handle(ctx context.Context, t *asynq.Task) error {
	var payload CompanyEnrichPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	companyID, err := uuid.Parse(payload.CompanyID)
	if err != nil {
		return fmt.Errorf("parse company ID: %w", err)
	}

	company, err := h.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return fmt.Errorf("get company: %w", err)
	}

	slog.Info("enriching company", "company_id", company.ID, "name", company.Name)

	// Build enrichment request from existing company data
	req := &ai.EnrichCompanyRequest{
		Name: company.Name,
	}
	if company.CareersPageURL != nil {
		req.CareersPageURL = *company.CareersPageURL
	}
	if company.LinkedinURL != nil {
		req.LinkedinURL = *company.LinkedinURL
	}

	result, err := h.aiProvider.EnrichCompany(ctx, req)
	if err != nil {
		slog.Error("company enrichment failed",
			"company_id", company.ID,
			"name", company.Name,
			"error", err,
		)
		return fmt.Errorf("enrich company %s: %w", company.Name, err)
	}

	// Update company with enriched data (only fill gaps, don't overwrite existing data)
	updated := false

	if len(company.TechStack) == 0 && len(result.TechStack) > 0 {
		company.TechStack = result.TechStack
		updated = true
	}
	if len(company.Domains) == 0 && len(result.Domains) > 0 {
		company.Domains = result.Domains
		updated = true
	}
	if company.Size == nil && result.Size != "" && result.Size != "unknown" {
		size := domain.CompanySize(result.Size)
		company.Size = &size
		updated = true
	}
	if company.HiringStatus == "unknown" && result.HiringStatus != "unknown" && result.HiringStatus != "" {
		company.HiringStatus = domain.HiringStatus(result.HiringStatus)
		updated = true
	}
	if company.Description == nil && result.Description != "" {
		company.Description = &result.Description
		updated = true
	}

	if updated {
		now := time.Now().UTC()
		company.LastVerifiedAt = &now
		company.UpdatedAt = now

		if err := h.companyRepo.Update(ctx, company); err != nil {
			return fmt.Errorf("update company: %w", err)
		}

		slog.Info("company enriched successfully",
			"company_id", company.ID,
			"name", company.Name,
			"tokens_in", result.TokensUsed.InputTokens,
			"tokens_out", result.TokensUsed.OutputTokens,
		)
	} else {
		slog.Info("company already has complete data, skipping enrichment",
			"company_id", company.ID,
			"name", company.Name,
		)
	}

	return nil
}

// CompanyRefreshHandler iterates over all companies and enqueues enrichment tasks.
type CompanyRefreshHandler struct {
	companyRepo domain.CompanyRepository
	asynqClient *asynq.Client
}

// NewCompanyRefreshHandler creates a handler for admin:company_refresh tasks.
func NewCompanyRefreshHandler(
	companyRepo domain.CompanyRepository,
	asynqClient *asynq.Client,
) *CompanyRefreshHandler {
	return &CompanyRefreshHandler{
		companyRepo: companyRepo,
		asynqClient: asynqClient,
	}
}

// Handle enqueues enrichment tasks for all companies needing a refresh.
// Companies not verified in the last 7 days are refreshed.
func (h *CompanyRefreshHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	companies, err := h.companyRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -7)
	var enqueued int

	for _, c := range companies {
		// Skip recently verified companies
		if c.LastVerifiedAt != nil && c.LastVerifiedAt.After(cutoff) {
			continue
		}

		payload, err := json.Marshal(CompanyEnrichPayload{CompanyID: c.ID.String()})
		if err != nil {
			slog.Error("marshal enrich payload", "company_id", c.ID, "error", err)
			continue
		}

		_, err = h.asynqClient.Enqueue(
			asynq.NewTask("admin:company_enrich", payload),
			asynq.Queue("low"),
			asynq.MaxRetry(2),
		)
		if err != nil {
			slog.Error("enqueue company enrich", "company_id", c.ID, "error", err)
			continue
		}

		enqueued++
	}

	slog.Info("company refresh: enqueued enrichment tasks",
		"total", len(companies),
		"enqueued", enqueued,
	)
	return nil
}
