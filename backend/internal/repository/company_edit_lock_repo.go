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

// CompanyEditLockRepo implements domain.CompanyEditLockRepository.
type CompanyEditLockRepo struct {
	pool *pgxpool.Pool
}

// NewCompanyEditLockRepo creates a new CompanyEditLockRepo.
func NewCompanyEditLockRepo(pool *pgxpool.Pool) *CompanyEditLockRepo {
	return &CompanyEditLockRepo{pool: pool}
}

// GetLock returns the current lock for a company, or nil if none or expired.
func (r *CompanyEditLockRepo) GetLock(ctx context.Context, companyID uuid.UUID) (*domain.CompanyEditLock, error) {
	q := getDBTX(ctx, r.pool)

	var lock domain.CompanyEditLock
	err := q.QueryRow(ctx, `
		SELECT company_id, locked_by, locked_at, expires_at
		FROM company_edit_locks
		WHERE company_id = $1 AND expires_at > NOW()`, companyID,
	).Scan(&lock.CompanyID, &lock.LockedBy, &lock.LockedAt, &lock.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}
	return &lock, nil
}

// AcquireLock creates a lock for a company. Cleans up expired locks first.
func (r *CompanyEditLockRepo) AcquireLock(ctx context.Context, lock *domain.CompanyEditLock) error {
	q := getDBTX(ctx, r.pool)

	// Clean up any expired lock first
	_, _ = q.Exec(ctx, `DELETE FROM company_edit_locks WHERE company_id = $1 AND expires_at <= NOW()`, lock.CompanyID)

	_, err := q.Exec(ctx, `
		INSERT INTO company_edit_locks (company_id, locked_by, locked_at, expires_at)
		VALUES ($1, $2, $3, $4)`,
		lock.CompanyID, lock.LockedBy, lock.LockedAt, lock.ExpiresAt,
	)
	if err != nil {
		// Unique constraint violation = already locked
		return domain.Conflict("company_edit_lock", "company is currently locked for editing")
	}
	return nil
}

// ReleaseLock removes a lock. Only the lock holder can release.
func (r *CompanyEditLockRepo) ReleaseLock(ctx context.Context, companyID, userID uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	tag, err := q.Exec(ctx, `
		DELETE FROM company_edit_locks WHERE company_id = $1 AND locked_by = $2`,
		companyID, userID,
	)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("company_edit_lock", companyID)
	}
	return nil
}

// CreateEdit records a moderator edit.
func (r *CompanyEditLockRepo) CreateEdit(ctx context.Context, edit *domain.CompanyEdit) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO company_edits (id, company_id, user_id, diff, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`,
		edit.ID, edit.CompanyID, edit.UserID, edit.Diff, time.Now(),
	).Scan(&edit.CreatedAt)
	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// GetLatestEdit returns the most recent edit by a user on a company, or nil.
func (r *CompanyEditLockRepo) GetLatestEdit(ctx context.Context, companyID, userID uuid.UUID) (*domain.CompanyEdit, error) {
	q := getDBTX(ctx, r.pool)

	var edit domain.CompanyEdit
	err := q.QueryRow(ctx, `
		SELECT id, company_id, user_id, diff, created_at
		FROM company_edits
		WHERE company_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT 1`, companyID, userID,
	).Scan(&edit.ID, &edit.CompanyID, &edit.UserID, &edit.Diff, &edit.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}
	return &edit, nil
}
