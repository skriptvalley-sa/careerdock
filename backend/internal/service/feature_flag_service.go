package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// FeatureFlagService handles feature flag business logic.
type FeatureFlagService struct {
	flags domain.FeatureFlagRepository
}

// NewFeatureFlagService creates a new FeatureFlagService.
func NewFeatureFlagService(flags domain.FeatureFlagRepository) *FeatureFlagService {
	return &FeatureFlagService{flags: flags}
}

// IsEnabled returns whether a feature flag is enabled by key.
// Returns false (not an error) if the flag does not exist.
func (s *FeatureFlagService) IsEnabled(ctx context.Context, key string) (bool, error) {
	flag, err := s.flags.GetByKey(ctx, key)
	if err != nil {
		// If not found, treat as disabled rather than error
		var appErr *domain.AppError
		if errors.As(err, &appErr) && appErr.Code == domain.ErrCodeNotFound {
			return false, nil
		}
		return false, err
	}
	return flag.Enabled, nil
}

// GetByKey returns a feature flag by key.
func (s *FeatureFlagService) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	return s.flags.GetByKey(ctx, key)
}

// List returns all feature flags.
func (s *FeatureFlagService) List(ctx context.Context) ([]domain.FeatureFlag, error) {
	return s.flags.List(ctx)
}

// Toggle updates a feature flag's enabled state and optional description.
// This is an admin-only operation.
func (s *FeatureFlagService) Toggle(ctx context.Context, id uuid.UUID, enabled bool, description *string) (*domain.FeatureFlag, error) {
	// Fetch all flags to find the target by ID
	flags, err := s.flags.List(ctx)
	if err != nil {
		return nil, err
	}

	var target *domain.FeatureFlag
	for i := range flags {
		if flags[i].ID == id {
			target = &flags[i]
			break
		}
	}
	if target == nil {
		return nil, domain.NotFound("feature_flag", id)
	}

	target.Enabled = enabled
	if description != nil {
		target.Description = description
	}

	if err := s.flags.Update(ctx, target); err != nil {
		return nil, err
	}
	return target, nil
}
