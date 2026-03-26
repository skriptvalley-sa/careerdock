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

// ResumeRepo implements domain.ResumeRepository using pgx.
type ResumeRepo struct {
	pool *pgxpool.Pool
}

// NewResumeRepo creates a new ResumeRepo.
func NewResumeRepo(pool *pgxpool.Pool) *ResumeRepo {
	return &ResumeRepo{pool: pool}
}

// Create inserts a new resume record.
func (r *ResumeRepo) Create(ctx context.Context, resume *domain.Resume) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO resumes (
			id, user_id, slot_number, file_name, file_size_bytes,
			s3_key, extracted_text, parsed_data, ats_general,
			status, failure_reason, is_default, is_archived, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15
		) RETURNING created_at, updated_at`,
		resume.ID, resume.UserID, resume.SlotNumber, resume.FileName, resume.FileSizeBytes,
		resume.S3Key, resume.ExtractedText, resume.ParsedData, resume.ATSGeneral,
		resume.Status, resume.FailureReason, resume.IsDefault, resume.IsArchived, time.Now(), time.Now(),
	).Scan(&resume.CreatedAt, &resume.UpdatedAt)
	if err != nil {
		return domain.InternalError(err)
	}

	return nil
}

// GetByID retrieves a resume by its ID.
func (r *ResumeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Resume, error) {
	q := getDBTX(ctx, r.pool)

	var resume domain.Resume
	var parsedData, atsGeneral []byte

	err := q.QueryRow(ctx, `
		SELECT id, user_id, slot_number, file_name, file_size_bytes,
		       s3_key, extracted_text, parsed_data, ats_general,
		       status, failure_reason, is_default, is_archived, archived_at, created_at, updated_at
		FROM resumes
		WHERE id = $1`, id,
	).Scan(
		&resume.ID, &resume.UserID, &resume.SlotNumber, &resume.FileName, &resume.FileSizeBytes,
		&resume.S3Key, &resume.ExtractedText, &parsedData, &atsGeneral,
		&resume.Status, &resume.FailureReason, &resume.IsDefault, &resume.IsArchived, &resume.ArchivedAt,
		&resume.CreatedAt, &resume.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NotFound("resume", id)
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}

	if parsedData != nil {
		resume.ParsedData = json.RawMessage(parsedData)
	}
	if atsGeneral != nil {
		resume.ATSGeneral = json.RawMessage(atsGeneral)
	}

	return &resume, nil
}

// ListByUser returns all non-archived resumes for a user, ordered by slot.
func (r *ResumeRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Resume, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT id, user_id, slot_number, file_name, file_size_bytes,
		       s3_key, extracted_text, parsed_data, ats_general,
		       status, failure_reason, is_default, is_archived, archived_at, created_at, updated_at
		FROM resumes
		WHERE user_id = $1 AND is_archived = false
		ORDER BY slot_number`, userID,
	)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var resumes []domain.Resume
	for rows.Next() {
		var res domain.Resume
		var parsedData, atsGeneral []byte

		if err := rows.Scan(
			&res.ID, &res.UserID, &res.SlotNumber, &res.FileName, &res.FileSizeBytes,
			&res.S3Key, &res.ExtractedText, &parsedData, &atsGeneral,
			&res.Status, &res.FailureReason, &res.IsDefault, &res.IsArchived, &res.ArchivedAt,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, domain.InternalError(err)
		}

		if parsedData != nil {
			res.ParsedData = json.RawMessage(parsedData)
		}
		if atsGeneral != nil {
			res.ATSGeneral = json.RawMessage(atsGeneral)
		}

		resumes = append(resumes, res)
	}

	return resumes, nil
}

// Update saves changes to an existing resume (status, parsed_data, ats_general, etc.).
func (r *ResumeRepo) Update(ctx context.Context, resume *domain.Resume) error {
	q := getDBTX(ctx, r.pool)

	now := time.Now()
	tag, err := q.Exec(ctx, `
		UPDATE resumes
		SET slot_number = $2, file_name = $3, file_size_bytes = $4,
		    s3_key = $5, extracted_text = $6, parsed_data = $7,
		    ats_general = $8, status = $9, failure_reason = $10,
		    is_default = $11, is_archived = $12, archived_at = $13, updated_at = $14
		WHERE id = $1`,
		resume.ID, resume.SlotNumber, resume.FileName, resume.FileSizeBytes,
		resume.S3Key, resume.ExtractedText, resume.ParsedData,
		resume.ATSGeneral, resume.Status, resume.FailureReason,
		resume.IsDefault, resume.IsArchived, resume.ArchivedAt, now,
	)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("resume", resume.ID)
	}

	resume.UpdatedAt = now
	return nil
}

// Archive marks a resume as archived.
func (r *ResumeRepo) Archive(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	now := time.Now()
	tag, err := q.Exec(ctx, `
		UPDATE resumes
		SET is_archived = true, archived_at = $2, is_default = false, updated_at = $2
		WHERE id = $1 AND is_archived = false`,
		id, now,
	)
	if err != nil {
		return domain.InternalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("resume", id)
	}

	return nil
}

// GetByUserAndSlot returns the active (non-archived) resume in a specific slot.
func (r *ResumeRepo) GetByUserAndSlot(ctx context.Context, userID uuid.UUID, slot int) (*domain.Resume, error) {
	q := getDBTX(ctx, r.pool)

	var resume domain.Resume
	var parsedData, atsGeneral []byte

	err := q.QueryRow(ctx, `
		SELECT id, user_id, slot_number, file_name, file_size_bytes,
		       s3_key, extracted_text, parsed_data, ats_general,
		       status, failure_reason, is_default, is_archived, archived_at, created_at, updated_at
		FROM resumes
		WHERE user_id = $1 AND slot_number = $2 AND is_archived = false`, userID, slot,
	).Scan(
		&resume.ID, &resume.UserID, &resume.SlotNumber, &resume.FileName, &resume.FileSizeBytes,
		&resume.S3Key, &resume.ExtractedText, &parsedData, &atsGeneral,
		&resume.Status, &resume.FailureReason, &resume.IsDefault, &resume.IsArchived, &resume.ArchivedAt,
		&resume.CreatedAt, &resume.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // slot is empty — not an error
	}
	if err != nil {
		return nil, domain.InternalError(err)
	}

	if parsedData != nil {
		resume.ParsedData = json.RawMessage(parsedData)
	}
	if atsGeneral != nil {
		resume.ATSGeneral = json.RawMessage(atsGeneral)
	}

	return &resume, nil
}

// ClearDefaultForUser removes the is_default flag from all resumes for a user.
func (r *ResumeRepo) ClearDefaultForUser(ctx context.Context, userID uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	_, err := q.Exec(ctx, `
		UPDATE resumes SET is_default = false, updated_at = $2
		WHERE user_id = $1 AND is_default = true`,
		userID, time.Now(),
	)
	if err != nil {
		return domain.InternalError(err)
	}

	return nil
}
