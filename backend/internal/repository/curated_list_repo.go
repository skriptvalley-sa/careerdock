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

// CuratedListRepo implements domain.CuratedListRepository using pgx.
type CuratedListRepo struct {
	pool *pgxpool.Pool
}

// NewCuratedListRepo creates a new CuratedListRepo.
func NewCuratedListRepo(pool *pgxpool.Pool) *CuratedListRepo {
	return &CuratedListRepo{pool: pool}
}

// Create inserts a new curated list record with an empty result placeholder.
func (r *CuratedListRepo) Create(ctx context.Context, list *domain.CuratedList) error {
	q := getDBTX(ctx, r.pool)

	result := list.Result
	if len(result) == 0 {
		result = json.RawMessage("{}")
	}

	now := time.Now()
	err := q.QueryRow(ctx, `
		INSERT INTO curated_lists (
			id, user_id, resume_id, name, preferences_hash, result, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING created_at, updated_at`,
		list.ID, list.UserID, list.ResumeID, list.Name, list.PreferencesHash, result, now,
	).Scan(&list.CreatedAt, &list.UpdatedAt)
	if err != nil {
		return domain.InternalError(err)
	}

	return nil
}

// GetByID retrieves a curated list by its primary key.
func (r *CuratedListRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CuratedList, error) {
	q := getDBTX(ctx, r.pool)

	var list domain.CuratedList
	var result []byte

	err := q.QueryRow(ctx, `
		SELECT id, user_id, resume_id, name, preferences_hash, result, created_at, updated_at
		FROM curated_lists
		WHERE id = $1`, id,
	).Scan(
		&list.ID, &list.UserID, &list.ResumeID, &list.Name,
		&list.PreferencesHash, &result, &list.CreatedAt, &list.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NotFound("curated_list", id)
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}

	if result != nil {
		list.Result = json.RawMessage(result)
	}

	return &list, nil
}

// ListByUser returns all curated lists for a user, newest first.
func (r *CuratedListRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.CuratedList, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT id, user_id, resume_id, name, preferences_hash, result, created_at, updated_at
		FROM curated_lists
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var lists []domain.CuratedList
	for rows.Next() {
		var list domain.CuratedList
		var result []byte

		if err := rows.Scan(
			&list.ID, &list.UserID, &list.ResumeID, &list.Name,
			&list.PreferencesHash, &result, &list.CreatedAt, &list.UpdatedAt,
		); err != nil {
			return nil, domain.InternalError(err)
		}

		if result != nil {
			list.Result = json.RawMessage(result)
		}

		lists = append(lists, list)
	}

	return lists, nil
}

// GetByPreferencesHash returns the most recent curated list matching the given hash.
// Returns nil (not an error) if no match is found.
func (r *CuratedListRepo) GetByPreferencesHash(ctx context.Context, hash string) (*domain.CuratedList, error) {
	q := getDBTX(ctx, r.pool)

	var list domain.CuratedList
	var result []byte

	err := q.QueryRow(ctx, `
		SELECT id, user_id, resume_id, name, preferences_hash, result, created_at, updated_at
		FROM curated_lists
		WHERE preferences_hash = $1
		ORDER BY created_at DESC
		LIMIT 1`, hash,
	).Scan(
		&list.ID, &list.UserID, &list.ResumeID, &list.Name,
		&list.PreferencesHash, &result, &list.CreatedAt, &list.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // cache miss — not an error
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}

	if result != nil {
		list.Result = json.RawMessage(result)
	}

	return &list, nil
}

// UpdateResult stores the AI-generated ranking result for a completed curated list.
func (r *CuratedListRepo) UpdateResult(ctx context.Context, id uuid.UUID, result json.RawMessage) error {
	q := getDBTX(ctx, r.pool)

	tag, err := q.Exec(ctx, `
		UPDATE curated_lists SET result = $2, updated_at = NOW() WHERE id = $1`,
		id, result,
	)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("curated_list", id)
	}

	return nil
}

// Rename updates the name of a curated list.
func (r *CuratedListRepo) Rename(ctx context.Context, id uuid.UUID, name string) error {
	q := getDBTX(ctx, r.pool)

	tag, err := q.Exec(ctx, `
		UPDATE curated_lists SET name = $2, updated_at = NOW() WHERE id = $1`,
		id, name,
	)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("curated_list", id)
	}

	return nil
}

// Delete removes a curated list.
func (r *CuratedListRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	tag, err := q.Exec(ctx, `DELETE FROM curated_lists WHERE id = $1`, id)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("curated_list", id)
	}

	return nil
}
