package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// ATSCheckRepo implements domain.ATSCheckRepository using pgx.
type ATSCheckRepo struct {
	pool *pgxpool.Pool
}

// NewATSCheckRepo creates a new ATSCheckRepo.
func NewATSCheckRepo(pool *pgxpool.Pool) *ATSCheckRepo {
	return &ATSCheckRepo{pool: pool}
}

// Create inserts a new ATS check record with an empty result placeholder.
func (r *ATSCheckRepo) Create(ctx context.Context, check *domain.ATSCheck) error {
	q := getDBTX(ctx, r.pool)

	result := check.Result
	if len(result) == 0 {
		result = json.RawMessage("{}")
	}

	err := q.QueryRow(ctx, `
		INSERT INTO ats_checks (
			id, user_id, resume_id, check_type, company_id,
			job_description, result, cache_key, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9
		) RETURNING created_at`,
		check.ID, check.UserID, check.ResumeID, check.CheckType, check.CompanyID,
		check.JobDescription, result, check.CacheKey, time.Now(),
	).Scan(&check.CreatedAt)
	if err != nil {
		return domain.InternalError(err)
	}

	return nil
}

// GetByID retrieves an ATS check by its primary key, with company name if applicable.
func (r *ATSCheckRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ATSCheck, error) {
	q := getDBTX(ctx, r.pool)

	var check domain.ATSCheck
	var result []byte

	err := q.QueryRow(ctx, `
		SELECT ac.id, ac.user_id, ac.resume_id, ac.check_type, ac.company_id,
		       ac.job_description, ac.result, ac.cache_key, ac.created_at,
		       c.name
		FROM ats_checks ac
		LEFT JOIN companies c ON c.id = ac.company_id
		WHERE ac.id = $1`, id,
	).Scan(
		&check.ID, &check.UserID, &check.ResumeID, &check.CheckType, &check.CompanyID,
		&check.JobDescription, &result, &check.CacheKey, &check.CreatedAt,
		&check.CompanyName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NotFound("ats_check", id)
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}

	if result != nil {
		check.Result = json.RawMessage(result)
	}

	return &check, nil
}

// ListByUser returns all ATS checks for a user, newest first, with company names.
func (r *ATSCheckRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.ATSCheck, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT ac.id, ac.user_id, ac.resume_id, ac.check_type, ac.company_id,
		       ac.job_description, ac.result, ac.cache_key, ac.created_at,
		       c.name
		FROM ats_checks ac
		LEFT JOIN companies c ON c.id = ac.company_id
		WHERE ac.user_id = $1
		ORDER BY ac.created_at DESC`, userID,
	)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var checks []domain.ATSCheck
	for rows.Next() {
		var check domain.ATSCheck
		var result []byte

		if err := rows.Scan(
			&check.ID, &check.UserID, &check.ResumeID, &check.CheckType, &check.CompanyID,
			&check.JobDescription, &result, &check.CacheKey, &check.CreatedAt,
			&check.CompanyName,
		); err != nil {
			return nil, domain.InternalError(err)
		}

		if result != nil {
			check.Result = json.RawMessage(result)
		}

		checks = append(checks, check)
	}

	return checks, nil
}

// GetByCacheKey looks up an existing ATS check by its deduplication key.
// Returns nil (not an error) if no match is found.
func (r *ATSCheckRepo) GetByCacheKey(ctx context.Context, cacheKey string) (*domain.ATSCheck, error) {
	q := getDBTX(ctx, r.pool)

	var check domain.ATSCheck
	var result []byte

	err := q.QueryRow(ctx, `
		SELECT id, user_id, resume_id, check_type, company_id,
		       job_description, result, cache_key, created_at
		FROM ats_checks
		WHERE cache_key = $1
		LIMIT 1`, cacheKey,
	).Scan(
		&check.ID, &check.UserID, &check.ResumeID, &check.CheckType, &check.CompanyID,
		&check.JobDescription, &result, &check.CacheKey, &check.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // cache miss — not an error
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}

	if result != nil {
		check.Result = json.RawMessage(result)
	}

	return &check, nil
}

// UpdateResult stores the AI-generated score result for a completed ATS check.
func (r *ATSCheckRepo) UpdateResult(ctx context.Context, id uuid.UUID, result json.RawMessage) error {
	q := getDBTX(ctx, r.pool)

	tag, err := q.Exec(ctx, `
		UPDATE ats_checks SET result = $2 WHERE id = $1`,
		id, result,
	)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("ats_check", id)
	}

	return nil
}
