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

// ApplicationRepo implements domain.ApplicationRepository using pgx.
type ApplicationRepo struct {
	pool *pgxpool.Pool
}

// NewApplicationRepo creates a new ApplicationRepo.
func NewApplicationRepo(pool *pgxpool.Pool) *ApplicationRepo {
	return &ApplicationRepo{pool: pool}
}

// Create inserts a new application.
func (r *ApplicationRepo) Create(ctx context.Context, app *domain.Application) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO applications (id, user_id, company_id, role_title, status, date_applied, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		app.ID, app.UserID, app.CompanyID, app.RoleTitle,
		string(app.Status), app.DateApplied, app.Notes,
		app.CreatedAt, app.UpdatedAt,
	).Scan(&app.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// GetByID retrieves an application by ID.
func (r *ApplicationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	q := getDBTX(ctx, r.pool)

	app := &domain.Application{}
	err := q.QueryRow(ctx, `
		SELECT id, user_id, company_id, role_title, status, date_applied, notes, created_at, updated_at
		FROM applications
		WHERE id = $1`, id,
	).Scan(&app.ID, &app.UserID, &app.CompanyID, &app.RoleTitle,
		&app.Status, &app.DateApplied, &app.Notes,
		&app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("application", id)
		}
		return nil, domain.InternalError(err)
	}
	return app, nil
}

// ListByUser returns all applications for a user, optionally filtered by status.
func (r *ApplicationRepo) ListByUser(ctx context.Context, userID uuid.UUID, statusFilter *domain.ApplicationStatus, excludeNotApplied bool) ([]domain.ApplicationWithCompany, error) {
	q := getDBTX(ctx, r.pool)

	query := `
		SELECT a.id, a.user_id, a.company_id, a.role_title, a.status,
		       a.date_applied, a.notes, a.created_at, a.updated_at,
		       COALESCE(c.name, '') AS company_name
		FROM applications a
		LEFT JOIN companies c ON c.id = a.company_id
		WHERE a.user_id = $1`
	args := []any{userID}

	if statusFilter != nil {
		query += ` AND a.status = $2`
		args = append(args, string(*statusFilter))
	} else if excludeNotApplied {
		query += ` AND a.status != 'not_applied'`
	}

	query += ` ORDER BY a.updated_at DESC`

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var apps []domain.ApplicationWithCompany
	for rows.Next() {
		var a domain.ApplicationWithCompany
		if err := rows.Scan(&a.ID, &a.UserID, &a.CompanyID, &a.RoleTitle,
			&a.Status, &a.DateApplied, &a.Notes,
			&a.CreatedAt, &a.UpdatedAt, &a.CompanyName); err != nil {
			return nil, domain.InternalError(err)
		}
		apps = append(apps, a)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	if apps == nil {
		apps = []domain.ApplicationWithCompany{}
	}
	return apps, nil
}

// ListByCompany returns all applications for a user+company.
func (r *ApplicationRepo) ListByCompany(ctx context.Context, userID, companyID uuid.UUID) ([]domain.Application, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT id, user_id, company_id, role_title, status, date_applied, notes, created_at, updated_at
		FROM applications
		WHERE user_id = $1 AND company_id = $2
		ORDER BY updated_at DESC`, userID, companyID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var apps []domain.Application
	for rows.Next() {
		var a domain.Application
		if err := rows.Scan(&a.ID, &a.UserID, &a.CompanyID, &a.RoleTitle,
			&a.Status, &a.DateApplied, &a.Notes,
			&a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, domain.InternalError(err)
		}
		apps = append(apps, a)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	if apps == nil {
		apps = []domain.Application{}
	}
	return apps, nil
}

// Update updates an existing application.
func (r *ApplicationRepo) Update(ctx context.Context, app *domain.Application) error {
	q := getDBTX(ctx, r.pool)
	app.UpdatedAt = time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE applications SET
			role_title = $2, status = $3, date_applied = $4, notes = $5, updated_at = $6
		WHERE id = $1
		RETURNING id`,
		app.ID, app.RoleTitle, string(app.Status), app.DateApplied, app.Notes, app.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("application", app.ID)
		}
		return domain.InternalError(err)
	}
	return nil
}

// Delete removes an application. Status history and interview rounds cascade.
func (r *ApplicationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `DELETE FROM applications WHERE id = $1 RETURNING id`, id).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("application", id)
		}
		return domain.InternalError(err)
	}
	return nil
}

// CountByStatus returns application counts grouped by status for a user's dashboard.
func (r *ApplicationRepo) CountByStatus(ctx context.Context, userID uuid.UUID) (map[domain.ApplicationStatus]int, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT status, COUNT(*)::int
		FROM applications
		WHERE user_id = $1
		GROUP BY status`, userID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	counts := make(map[domain.ApplicationStatus]int)
	for rows.Next() {
		var status domain.ApplicationStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, domain.InternalError(err)
		}
		counts[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return counts, nil
}

// --- Status History ---

// CreateStatusHistory inserts a new status history record.
func (r *ApplicationRepo) CreateStatusHistory(ctx context.Context, h *domain.StatusHistory) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO application_status_history (id, application_id, from_status, to_status, changed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		h.ID, h.ApplicationID, h.FromStatus, string(h.ToStatus), h.ChangedAt,
	).Scan(&h.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// ListStatusHistory returns status changes for an application, chronological.
func (r *ApplicationRepo) ListStatusHistory(ctx context.Context, applicationID uuid.UUID) ([]domain.StatusHistory, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT id, application_id, from_status, to_status, changed_at
		FROM application_status_history
		WHERE application_id = $1
		ORDER BY changed_at ASC`, applicationID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var history []domain.StatusHistory
	for rows.Next() {
		var h domain.StatusHistory
		if err := rows.Scan(&h.ID, &h.ApplicationID, &h.FromStatus, &h.ToStatus, &h.ChangedAt); err != nil {
			return nil, domain.InternalError(err)
		}
		history = append(history, h)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return history, nil
}

// --- Interview Rounds ---

// CreateInterviewRound inserts a new interview round.
func (r *ApplicationRepo) CreateInterviewRound(ctx context.Context, round *domain.InterviewRound) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO interview_rounds (id, application_id, round_number, round_type, scheduled_date, outcome, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		round.ID, round.ApplicationID, round.RoundNumber, round.RoundType,
		round.ScheduledDate, string(round.Outcome), round.Notes,
		round.CreatedAt, round.UpdatedAt,
	).Scan(&round.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// GetInterviewRoundByID retrieves an interview round by ID.
func (r *ApplicationRepo) GetInterviewRoundByID(ctx context.Context, id uuid.UUID) (*domain.InterviewRound, error) {
	q := getDBTX(ctx, r.pool)

	round := &domain.InterviewRound{}
	err := q.QueryRow(ctx, `
		SELECT id, application_id, round_number, round_type, scheduled_date, outcome, notes, created_at, updated_at
		FROM interview_rounds
		WHERE id = $1`, id,
	).Scan(&round.ID, &round.ApplicationID, &round.RoundNumber, &round.RoundType,
		&round.ScheduledDate, &round.Outcome, &round.Notes,
		&round.CreatedAt, &round.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("interview round", id)
		}
		return nil, domain.InternalError(err)
	}
	return round, nil
}

// UpdateInterviewRound updates an existing interview round.
func (r *ApplicationRepo) UpdateInterviewRound(ctx context.Context, round *domain.InterviewRound) error {
	q := getDBTX(ctx, r.pool)
	round.UpdatedAt = time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE interview_rounds SET
			outcome = $2, notes = $3, scheduled_date = $4, updated_at = $5
		WHERE id = $1
		RETURNING id`,
		round.ID, string(round.Outcome), round.Notes, round.ScheduledDate, round.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("interview round", round.ID)
		}
		return domain.InternalError(err)
	}
	return nil
}

// DeleteInterviewRound removes an interview round.
func (r *ApplicationRepo) DeleteInterviewRound(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `DELETE FROM interview_rounds WHERE id = $1 RETURNING id`, id).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("interview round", id)
		}
		return domain.InternalError(err)
	}
	return nil
}
