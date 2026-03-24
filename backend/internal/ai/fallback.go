package ai

import (
	"context"
	"fmt"
	"log/slog"
)

// FallbackProvider tries the primary provider, falls back to secondary on failure.
type FallbackProvider struct {
	primary   LLMProvider
	secondary LLMProvider
}

// NewFallbackProvider creates a provider that tries primary first, then secondary.
func NewFallbackProvider(primary, secondary LLMProvider) *FallbackProvider {
	return &FallbackProvider{
		primary:   primary,
		secondary: secondary,
	}
}

// Name returns the provider name.
func (f *FallbackProvider) Name() string {
	return fmt.Sprintf("fallback(%s→%s)", f.primary.Name(), f.secondary.Name())
}

// ParseResume tries primary, falls back to secondary.
func (f *FallbackProvider) ParseResume(ctx context.Context, req *ParseResumeRequest) (*ParsedResume, error) {
	result, err := f.primary.ParseResume(ctx, req)
	if err == nil {
		return result, nil
	}

	slog.Warn("primary provider failed, trying fallback",
		"operation", "parse_resume",
		"primary", f.primary.Name(),
		"primary_error", err.Error(),
	)

	result, err = f.secondary.ParseResume(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("all providers failed: primary=%s, fallback=%s: %w",
			f.primary.Name(), f.secondary.Name(), err)
	}
	return result, nil
}

// ScoreATSCompany tries primary, falls back to secondary.
func (f *FallbackProvider) ScoreATSCompany(ctx context.Context, req *ATSCompanyRequest) (*ATSResult, error) {
	result, err := f.primary.ScoreATSCompany(ctx, req)
	if err == nil {
		return result, nil
	}

	slog.Warn("primary provider failed, trying fallback",
		"operation", "score_ats_company",
		"primary", f.primary.Name(),
		"primary_error", err.Error(),
	)

	result, err = f.secondary.ScoreATSCompany(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("all providers failed: primary=%s, fallback=%s: %w",
			f.primary.Name(), f.secondary.Name(), err)
	}
	return result, nil
}

// ScoreATSJob tries primary, falls back to secondary.
func (f *FallbackProvider) ScoreATSJob(ctx context.Context, req *ATSJobRequest) (*ATSResult, error) {
	result, err := f.primary.ScoreATSJob(ctx, req)
	if err == nil {
		return result, nil
	}

	slog.Warn("primary provider failed, trying fallback",
		"operation", "score_ats_job",
		"primary", f.primary.Name(),
		"primary_error", err.Error(),
	)

	result, err = f.secondary.ScoreATSJob(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("all providers failed: primary=%s, fallback=%s: %w",
			f.primary.Name(), f.secondary.Name(), err)
	}
	return result, nil
}

// CurateCompanyList tries primary, falls back to secondary.
func (f *FallbackProvider) CurateCompanyList(ctx context.Context, req *CurateListRequest) (*CuratedListResult, error) {
	result, err := f.primary.CurateCompanyList(ctx, req)
	if err == nil {
		return result, nil
	}

	slog.Warn("primary provider failed, trying fallback",
		"operation", "curate_company_list",
		"primary", f.primary.Name(),
		"primary_error", err.Error(),
	)

	result, err = f.secondary.CurateCompanyList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("all providers failed: primary=%s, fallback=%s: %w",
			f.primary.Name(), f.secondary.Name(), err)
	}
	return result, nil
}

// ScoreATSGeneral tries primary, falls back to secondary.
func (f *FallbackProvider) ScoreATSGeneral(ctx context.Context, req *ATSGeneralRequest) (*ATSResult, error) {
	result, err := f.primary.ScoreATSGeneral(ctx, req)
	if err == nil {
		return result, nil
	}

	slog.Warn("primary provider failed, trying fallback",
		"operation", "score_ats_general",
		"primary", f.primary.Name(),
		"primary_error", err.Error(),
	)

	result, err = f.secondary.ScoreATSGeneral(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("all providers failed: primary=%s, fallback=%s: %w",
			f.primary.Name(), f.secondary.Name(), err)
	}
	return result, nil
}
