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

// UserRepo implements domain.UserRepository using pgx.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create inserts a new user. The user.ID must be set by the caller (UUID v7).
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO users (
			id, email, password_hash, name, role,
			email_verified, current_title, experience_level,
			preferred_tech_stacks, target_domains, target_locations,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13
		) RETURNING id`,
		user.ID, user.Email, user.PasswordHash, user.Name, string(user.Role),
		user.EmailVerified, user.CurrentTitle, experienceLevelToString(user.ExperienceLevel),
		user.PreferredTechStacks, user.TargetDomains, user.TargetLocations,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.Conflict("user", "email already registered")
		}
		return domain.InternalError(err)
	}
	return nil
}

// GetByID retrieves a user by ID, excluding soft-deleted users.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	q := getDBTX(ctx, r.pool)
	return scanUser(q.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role,
			premium_since, email_verified, current_title, experience_level,
			preferred_tech_stacks, target_domains, target_locations,
			default_resume_id, deleted_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`, id))
}

// GetByEmail retrieves a user by email, excluding soft-deleted users.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := getDBTX(ctx, r.pool)
	return scanUser(q.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role,
			premium_since, email_verified, current_title, experience_level,
			preferred_tech_stacks, target_domains, target_locations,
			default_resume_id, deleted_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`, email))
}

// Update persists changes to an existing user. Sets updated_at to now.
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	q := getDBTX(ctx, r.pool)
	user.UpdatedAt = time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE users SET
			name = $2,
			role = $3,
			premium_since = $4,
			email_verified = $5,
			current_title = $6,
			experience_level = $7,
			preferred_tech_stacks = $8,
			target_domains = $9,
			target_locations = $10,
			default_resume_id = $11,
			password_hash = $12,
			updated_at = $13
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id`,
		user.ID,
		user.Name,
		string(user.Role),
		user.PremiumSince,
		user.EmailVerified,
		user.CurrentTitle,
		experienceLevelToString(user.ExperienceLevel),
		user.PreferredTechStacks,
		user.TargetDomains,
		user.TargetLocations,
		user.DefaultResumeID,
		user.PasswordHash,
		user.UpdatedAt,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("user", user.ID)
		}
		return domain.InternalError(err)
	}
	return nil
}

// SoftDelete marks a user as deleted (30-day grace period).
func (r *UserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)
	now := time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE users SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id`, id, now).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("user", id)
		}
		return domain.InternalError(err)
	}
	return nil
}

// --- Token repositories for auth flow ---

// TokenRepo implements email verification and password reset token operations.
type TokenRepo struct {
	pool *pgxpool.Pool
}

// NewTokenRepo creates a new TokenRepo.
func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{pool: pool}
}

// CreateEmailVerificationToken inserts a new email verification token.
func (r *TokenRepo) CreateEmailVerificationToken(ctx context.Context, token *domain.EmailVerificationToken) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO email_verification_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt,
	).Scan(&token.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// GetEmailVerificationToken retrieves an unused, non-expired token.
func (r *TokenRepo) GetEmailVerificationToken(ctx context.Context, tokenStr string) (*domain.EmailVerificationToken, error) {
	q := getDBTX(ctx, r.pool)

	t := &domain.EmailVerificationToken{}
	err := q.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM email_verification_tokens
		WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()`, tokenStr,
	).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("email verification token", tokenStr)
		}
		return nil, domain.InternalError(err)
	}
	return t, nil
}

// MarkEmailVerificationTokenUsed marks a token as used.
func (r *TokenRepo) MarkEmailVerificationTokenUsed(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)
	now := time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE email_verification_tokens SET used_at = $2
		WHERE id = $1
		RETURNING id`, id, now).Scan(&returnedID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// CreatePasswordResetToken inserts a new password reset token.
func (r *TokenRepo) CreatePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt,
	).Scan(&token.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// GetPasswordResetToken retrieves an unused, non-expired password reset token.
func (r *TokenRepo) GetPasswordResetToken(ctx context.Context, tokenStr string) (*domain.PasswordResetToken, error) {
	q := getDBTX(ctx, r.pool)

	t := &domain.PasswordResetToken{}
	err := q.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()`, tokenStr,
	).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("password reset token", tokenStr)
		}
		return nil, domain.InternalError(err)
	}
	return t, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used.
func (r *TokenRepo) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)
	now := time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE password_reset_tokens SET used_at = $2
		WHERE id = $1
		RETURNING id`, id, now).Scan(&returnedID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// --- Helpers ---

// scanUser scans a single user row.
func scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	var expLevel *string

	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role,
		&u.PremiumSince, &u.EmailVerified, &u.CurrentTitle, &expLevel,
		&u.PreferredTechStacks, &u.TargetDomains, &u.TargetLocations,
		&u.DefaultResumeID, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("user", nil)
		}
		return nil, domain.InternalError(err)
	}

	if expLevel != nil {
		el := domain.ExperienceLevel(*expLevel)
		u.ExperienceLevel = &el
	}

	return u, nil
}

// experienceLevelToString converts *ExperienceLevel to *string for DB.
func experienceLevelToString(el *domain.ExperienceLevel) *string {
	if el == nil {
		return nil
	}
	s := string(*el)
	return &s
}

// isUniqueViolation checks if a pgx error is a unique constraint violation.
func isUniqueViolation(err error) bool {
	// pgx wraps the Postgres error code.
	// Code 23505 = unique_violation
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
