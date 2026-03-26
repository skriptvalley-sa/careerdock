package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/ai"
	"github.com/skriptvalley/careerdock/internal/domain"
)

const (
	editLockDuration = 10 * time.Minute
	editCooldown     = 10 * time.Minute
)

// ModeratorService handles moderator workflows: AI company generation, editing with locks.
type ModeratorService struct {
	companies  domain.CompanyRepository
	editLocks  domain.CompanyEditLockRepository
	aiProvider ai.LLMProvider
}

// NewModeratorService creates a new ModeratorService.
func NewModeratorService(
	companies domain.CompanyRepository,
	editLocks domain.CompanyEditLockRepository,
	aiProvider ai.LLMProvider,
) *ModeratorService {
	return &ModeratorService{
		companies:  companies,
		editLocks:  editLocks,
		aiProvider: aiProvider,
	}
}

// GenerateCompanyDraft calls the AI provider to generate a draft company profile.
// The input name, careers URL, and LinkedIn URL are guaranteed to appear in the
// result even if the AI leaves those fields empty.
// Returns a Conflict error (without calling AI) if a company with the same slug
// already exists in the directory.
func (s *ModeratorService) GenerateCompanyDraft(ctx context.Context, name, careersURL, linkedinURL string) (*ai.EnrichedCompany, error) {
	if name == "" {
		return nil, domain.ValidationError("company name is required", nil)
	}

	// Duplicate check: derive a slug from the name and see if it already exists.
	// This avoids spending an AI call on a company we already have.
	candidate := slugify(name)
	if existing, err := s.companies.GetBySlug(ctx, candidate); err == nil && existing != nil {
		return nil, domain.ValidationError(
			fmt.Sprintf("%q already exists in the directory. Edit it instead.", existing.Name),
			nil,
		)
	}

	result, err := s.aiProvider.EnrichCompany(ctx, &ai.EnrichCompanyRequest{
		Name:           name,
		CareersPageURL: careersURL,
		LinkedinURL:    linkedinURL,
	})
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("AI company enrichment failed: %w", err))
	}

	// Ensure the input values are never lost: fall back to what the caller
	// provided when the AI returns an empty field.
	if result.Name == "" {
		result.Name = name
	}
	if result.CareersPageURL == "" && careersURL != "" {
		result.CareersPageURL = careersURL
	}
	if result.LinkedinURL == "" && linkedinURL != "" {
		result.LinkedinURL = linkedinURL
	}

	return result, nil
}

// SubmitCompanyDraft creates a company from a reviewed moderator draft.
func (s *ModeratorService) SubmitCompanyDraft(ctx context.Context, input CreateCompanyInput) (*domain.Company, error) {
	if input.Name == "" {
		return nil, domain.ValidationError("name is required", nil)
	}
	if input.Slug == "" {
		input.Slug = slugify(input.Name)
	}

	company := &domain.Company{
		ID:                uuid.Must(uuid.NewV7()),
		Slug:              input.Slug,
		Name:              input.Name,
		Description:       input.Description,
		Size:              input.Size,
		Headquarters:      input.Headquarters,
		FoundedYear:       input.FoundedYear,
		CareersPageURL:    input.CareersPageURL,
		LinkedinURL:       input.LinkedinURL,
		TechStack:         input.TechStack,
		Domains:           input.Domains,
		HiringStatus:      input.HiringStatus,
		OfficeModes:       input.OfficeModes,
		CompensationTier:  input.CompensationTier,
		HasRSU:            input.HasRSU,
		HasRSURefresher:   input.HasRSURefresher,
		CompensationBands: input.CompensationBands,
	}

	if err := s.companies.Create(ctx, company); err != nil {
		return nil, err
	}

	return company, nil
}

// AcquireEditLock attempts to lock a company for editing.
// Checks cooldown: if the user submitted an edit on this company within the last 10 minutes, deny.
func (s *ModeratorService) AcquireEditLock(ctx context.Context, userID, companyID uuid.UUID) (*domain.CompanyEditLock, error) {
	// Check company exists
	if _, err := s.companies.GetByID(ctx, companyID); err != nil {
		return nil, err
	}

	// Check cooldown: has this user edited this company recently?
	latestEdit, err := s.editLocks.GetLatestEdit(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	if latestEdit != nil && time.Since(latestEdit.CreatedAt) < editCooldown {
		remaining := editCooldown - time.Since(latestEdit.CreatedAt)
		return nil, domain.ValidationError(
			fmt.Sprintf("Please wait %d minutes before editing this company again", int(remaining.Minutes())+1),
			map[string]any{"cooldown_remaining_seconds": int(remaining.Seconds())},
		)
	}

	now := time.Now()
	lock := &domain.CompanyEditLock{
		CompanyID: companyID,
		LockedBy:  userID,
		LockedAt:  now,
		ExpiresAt: now.Add(editLockDuration),
	}

	if err := s.editLocks.AcquireLock(ctx, lock); err != nil {
		return nil, err
	}

	return lock, nil
}

// ReleaseEditLock releases a lock on a company.
func (s *ModeratorService) ReleaseEditLock(ctx context.Context, userID, companyID uuid.UUID) error {
	return s.editLocks.ReleaseLock(ctx, companyID, userID)
}

// GetEditStatus returns the lock status for a company.
func (s *ModeratorService) GetEditStatus(ctx context.Context, companyID uuid.UUID) (*domain.CompanyEditLock, error) {
	return s.editLocks.GetLock(ctx, companyID)
}

// SubmitCompanyEdit applies edits to a company and records the diff.
func (s *ModeratorService) SubmitCompanyEdit(ctx context.Context, userID, companyID uuid.UUID, changes UpdateCompanyInput) (*domain.Company, error) {
	company, err := s.companies.GetByID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Apply changes
	if changes.Description != nil {
		company.Description = changes.Description
	}
	if changes.Size != nil {
		company.Size = changes.Size
	}
	if changes.Headquarters != nil {
		company.Headquarters = changes.Headquarters
	}
	if changes.CareersPageURL != nil {
		company.CareersPageURL = changes.CareersPageURL
	}
	if changes.LinkedinURL != nil {
		company.LinkedinURL = changes.LinkedinURL
	}
	if changes.TechStack != nil {
		company.TechStack = changes.TechStack
	}
	if changes.Domains != nil {
		company.Domains = changes.Domains
	}
	if changes.HiringStatus != nil {
		company.HiringStatus = *changes.HiringStatus
	}
	if changes.OfficeModes != nil {
		company.OfficeModes = changes.OfficeModes
	}
	if changes.CompensationTier != nil {
		company.CompensationTier = changes.CompensationTier
	}

	if err := s.companies.Update(ctx, company); err != nil {
		return nil, err
	}

	// Record the edit
	diffJSON, _ := json.Marshal(changes)
	edit := &domain.CompanyEdit{
		ID:        uuid.Must(uuid.NewV7()),
		CompanyID: companyID,
		UserID:    userID,
		Diff:      diffJSON,
	}
	if err := s.editLocks.CreateEdit(ctx, edit); err != nil {
		return nil, err
	}

	// Release the lock
	_ = s.editLocks.ReleaseLock(ctx, companyID, userID)

	return company, nil
}

// slugify creates a URL slug from a company name.
func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
