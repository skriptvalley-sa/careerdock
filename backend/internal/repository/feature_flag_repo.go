package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// FeatureFlagRepo implements domain.FeatureFlagRepository using pgx.
type FeatureFlagRepo struct {
	pool *pgxpool.Pool
}

// NewFeatureFlagRepo creates a new FeatureFlagRepo.
func NewFeatureFlagRepo(pool *pgxpool.Pool) *FeatureFlagRepo {
	return &FeatureFlagRepo{pool: pool}
}

// GetByKey returns a feature flag by its unique key.
func (r *FeatureFlagRepo) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	q := getDBTX(ctx, r.pool)

	flag := &domain.FeatureFlag{}
	err := q.QueryRow(ctx, `
		SELECT id, key, enabled, description, created_at, updated_at
		FROM feature_flags
		WHERE key = $1`, key,
	).Scan(&flag.ID, &flag.Key, &flag.Enabled, &flag.Description,
		&flag.CreatedAt, &flag.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("feature_flag", key)
		}
		return nil, domain.InternalError(err)
	}
	return flag, nil
}

// List returns all feature flags ordered by key.
func (r *FeatureFlagRepo) List(ctx context.Context) ([]domain.FeatureFlag, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT id, key, enabled, description, created_at, updated_at
		FROM feature_flags
		ORDER BY key ASC`)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var flags []domain.FeatureFlag
	for rows.Next() {
		var f domain.FeatureFlag
		if err := rows.Scan(&f.ID, &f.Key, &f.Enabled, &f.Description,
			&f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, domain.InternalError(err)
		}
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	if flags == nil {
		flags = []domain.FeatureFlag{}
	}
	return flags, nil
}

// Update updates a feature flag's enabled state and description.
func (r *FeatureFlagRepo) Update(ctx context.Context, flag *domain.FeatureFlag) error {
	q := getDBTX(ctx, r.pool)
	flag.UpdatedAt = time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE feature_flags SET enabled = $2, description = $3, updated_at = $4
		WHERE id = $1
		RETURNING id`,
		flag.ID, flag.Enabled, flag.Description, flag.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("feature_flag", flag.ID)
		}
		return domain.InternalError(err)
	}
	return nil
}
