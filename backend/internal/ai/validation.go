package ai

import (
	"context"
	"fmt"
	"log/slog"
)

// ATSCompanyCategories lists the required breakdown keys for company ATS scoring.
var ATSCompanyCategories = []string{
	"tech_stack_match",
	"domain_fit",
	"seniority_fit",
	"keyword_density",
	"impact_metrics",
}

// ATSJobCategories lists the required breakdown keys for job ATS scoring.
var ATSJobCategories = []string{
	"required_skills",
	"preferred_skills",
	"experience_level",
	"domain_relevance",
	"keyword_density",
}

// ValidateATSResult checks an ATSResult for correctness:
//   - Overall score is 0–100
//   - All requiredCategories are present in the breakdown
//   - Each category score is 0–100
//   - Suggestions list has between 1 and 7 items
func ValidateATSResult(result *ATSResult, requiredCategories []string) error {
	if result == nil {
		return fmt.Errorf("ATS result is nil")
	}

	if result.Score < 0 || result.Score > 100 {
		return fmt.Errorf("overall score %d is out of bounds [0, 100]", result.Score)
	}

	if result.Breakdown == nil {
		return fmt.Errorf("breakdown is nil")
	}

	for _, cat := range requiredCategories {
		detail, ok := result.Breakdown[cat]
		if !ok {
			return fmt.Errorf("missing required breakdown category: %q", cat)
		}
		if detail.Score < 0 || detail.Score > 100 {
			return fmt.Errorf("breakdown category %q score %d is out of bounds [0, 100]", cat, detail.Score)
		}
	}

	if len(result.Suggestions) < 1 || len(result.Suggestions) > 7 {
		return fmt.Errorf("suggestions count %d is out of bounds [1, 7]", len(result.Suggestions))
	}

	return nil
}

// ValidateATSResultRetry wraps an AI call with retry logic on validation failures.
// fn is called up to (1 + maxRetries) times total.
// Returns the first valid result, or the last error if all attempts fail.
func ValidateATSResultRetry(
	_ context.Context,
	maxRetries int,
	categories []string,
	fn func() (*ATSResult, error),
) (*ATSResult, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			slog.Warn("retrying AI ATS call after validation failure",
				"attempt", attempt,
				"max_retries", maxRetries,
				"last_error", lastErr.Error(),
			)
		}

		result, err := fn()
		if err != nil {
			lastErr = err
			continue
		}

		if err := ValidateATSResult(result, categories); err != nil {
			lastErr = fmt.Errorf("validation failed (attempt %d): %w", attempt+1, err)
			continue
		}

		return result, nil
	}

	return nil, fmt.Errorf("all %d AI attempts failed: %w", maxRetries+1, lastErr)
}
