// Package service implements business logic, orchestrating repositories
// and external services. Services own transaction boundaries.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/repository"
)

const (
	bcryptCost = 12

	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour

	emailVerificationTTL = 24 * time.Hour
	passwordResetTTL     = 1 * time.Hour

	// Redis key prefixes
	refreshBlacklistPrefix = "blacklist:"
)

// AuthService handles authentication and authorisation logic.
type AuthService struct {
	users     domain.UserRepository
	tokens    *repository.TokenRepo
	tx        domain.Transactor
	redis     *redis.Client
	jwtSecret []byte
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	users domain.UserRepository,
	tokens *repository.TokenRepo,
	tx domain.Transactor,
	redisClient *redis.Client,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		users:     users,
		tokens:    tokens,
		tx:        tx,
		redis:     redisClient,
		jwtSecret: []byte(jwtSecret),
	}
}

// TokenPair holds the generated JWT tokens and their metadata.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessExp    time.Time
	RefreshExp   time.Time
	RefreshJTI   string // for blacklisting
}

// RegisterInput holds validated registration data.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// LoginInput holds validated login data.
type LoginInput struct {
	Email    string
	Password string
}

// Register creates a new user account and returns the user and a token pair.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, *TokenPair, error) {
	// Validate password
	if err := validatePassword(input.Password); err != nil {
		return nil, nil, err
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		return nil, nil, domain.InternalError(fmt.Errorf("hash password: %w", err))
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:            uuid.Must(uuid.NewV7()),
		Email:         input.Email,
		PasswordHash:  string(hash),
		Name:          input.Name,
		Role:          domain.RoleUser,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Create user and email verification token in a transaction
	var verificationToken string
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.users.Create(ctx, user); err != nil {
			return err
		}

		// Create email verification token
		verificationToken, err = generateSecureToken()
		if err != nil {
			return domain.InternalError(err)
		}

		evToken := &domain.EmailVerificationToken{
			ID:        uuid.Must(uuid.NewV7()),
			UserID:    user.ID,
			Token:     verificationToken,
			ExpiresAt: now.Add(emailVerificationTTL),
			CreatedAt: now,
		}
		return s.tokens.CreateEmailVerificationToken(ctx, evToken)
	})
	if err != nil {
		return nil, nil, err
	}

	// Generate JWT tokens
	pair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Queue email verification email via Asynq
	slog.Info("email verification token created",
		"user_id", user.ID,
		"token_preview", verificationToken[:8]+"...",
	)

	return user, pair, nil
}

// Login authenticates a user by email and password.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*domain.User, *TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) && appErr.Code == domain.ErrCodeNotFound {
			// Don't reveal whether the email exists
			return nil, nil, domain.Unauthorized("invalid email or password")
		}
		return nil, nil, err
	}

	// Check if account is soft-deleted
	if user.DeletedAt != nil {
		return nil, nil, domain.Forbidden("account is suspended")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, domain.Unauthorized("invalid email or password")
	}

	pair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

// Refresh exchanges a valid refresh token for a new token pair.
// The old refresh token is blacklisted.
func (s *AuthService) Refresh(ctx context.Context, refreshTokenStr string) (*domain.User, *TokenPair, error) {
	// Parse and validate the refresh token
	claims, err := s.parseToken(refreshTokenStr)
	if err != nil {
		return nil, nil, domain.Unauthorized("invalid or expired refresh token")
	}

	// Check token type
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, nil, domain.Unauthorized("invalid token type")
	}

	// Check if blacklisted
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil, nil, domain.Unauthorized("invalid refresh token")
	}

	blacklisted, err := s.isTokenBlacklisted(ctx, jti)
	if err != nil {
		return nil, nil, domain.InternalError(err)
	}
	if blacklisted {
		return nil, nil, domain.Unauthorized("refresh token has been revoked")
	}

	// Blacklist the old refresh token
	if err := s.blacklistToken(ctx, jti); err != nil {
		return nil, nil, domain.InternalError(err)
	}

	// Get user
	sub, _ := claims["sub"].(string)
	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, nil, domain.Unauthorized("invalid token subject")
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	// Generate new token pair
	pair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

// Logout blacklists the current refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	claims, err := s.parseToken(refreshTokenStr)
	if err != nil {
		// Token already invalid — treat as success
		return nil //nolint:nilerr // intentional: expired/invalid token logout is a no-op
	}

	jti, _ := claims["jti"].(string)
	if jti != "" {
		if err := s.blacklistToken(ctx, jti); err != nil {
			return domain.InternalError(err)
		}
	}

	return nil
}

// VerifyEmail marks a user's email as verified using the provided token.
func (s *AuthService) VerifyEmail(ctx context.Context, tokenStr string) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		token, err := s.tokens.GetEmailVerificationToken(ctx, tokenStr)
		if err != nil {
			var appErr *domain.AppError
			if errors.As(err, &appErr) && appErr.Code == domain.ErrCodeNotFound {
				return domain.ValidationError("invalid or expired verification token", nil)
			}
			return err
		}

		// Mark token as used
		if err := s.tokens.MarkEmailVerificationTokenUsed(ctx, token.ID); err != nil {
			return err
		}

		// Update user
		user, err := s.users.GetByID(ctx, token.UserID)
		if err != nil {
			return err
		}

		user.EmailVerified = true
		return s.users.Update(ctx, user)
	})
}

// ForgotPassword creates a password reset token and queues a reset email.
// Always returns nil to prevent user enumeration.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal whether the email exists — intentionally return nil.
		return nil //nolint:nilerr // security: prevent user enumeration
	}

	tokenStr, err := generateSecureToken()
	if err != nil {
		slog.Error("failed to generate password reset token", "error", err)
		return nil //nolint:nilerr // security: don't expose internal failures
	}

	now := time.Now().UTC()
	resetToken := &domain.PasswordResetToken{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: now.Add(passwordResetTTL),
		CreatedAt: now,
	}

	if err := s.tokens.CreatePasswordResetToken(ctx, resetToken); err != nil {
		slog.Error("failed to create password reset token", "error", err)
		return nil
	}

	// TODO: Queue password reset email via Asynq
	slog.Info("password reset token created",
		"user_id", user.ID,
		"token_preview", tokenStr[:8]+"...",
	)

	return nil
}

// ResetPasswordInput holds validated password reset data.
type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

// ResetPassword resets a user's password using a valid reset token.
func (s *AuthService) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	if err := validatePassword(input.NewPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcryptCost)
	if err != nil {
		return domain.InternalError(fmt.Errorf("hash password: %w", err))
	}

	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		token, err := s.tokens.GetPasswordResetToken(ctx, input.Token)
		if err != nil {
			var appErr *domain.AppError
			if errors.As(err, &appErr) && appErr.Code == domain.ErrCodeNotFound {
				return domain.ValidationError("invalid or expired reset token", nil)
			}
			return err
		}

		// Mark token as used
		if err := s.tokens.MarkPasswordResetTokenUsed(ctx, token.ID); err != nil {
			return err
		}

		// Update password
		user, err := s.users.GetByID(ctx, token.UserID)
		if err != nil {
			return err
		}

		user.PasswordHash = string(hash)
		// TODO: Invalidate all refresh tokens for the user in Redis
		return s.users.Update(ctx, user)
	})
}

// GetUserByID retrieves a user by ID (used by middleware / handlers).
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

// ValidateAccessToken parses and validates an access token, returning the claims.
func (s *AuthService) ValidateAccessToken(tokenStr string) (uuid.UUID, domain.Role, error) {
	claims, err := s.parseToken(tokenStr)
	if err != nil {
		return uuid.Nil, "", domain.Unauthorized("invalid or expired access token")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "access" {
		return uuid.Nil, "", domain.Unauthorized("invalid token type")
	}

	sub, _ := claims["sub"].(string)
	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, "", domain.Unauthorized("invalid token subject")
	}

	role := domain.Role(claims["role"].(string))
	return userID, role, nil
}

// --- Internal helpers ---

func (s *AuthService) generateTokenPair(user *domain.User) (*TokenPair, error) {
	now := time.Now().UTC()
	accessExp := now.Add(accessTokenTTL)
	refreshExp := now.Add(refreshTokenTTL)
	refreshJTI := uuid.Must(uuid.NewV7()).String()

	// Access token
	accessClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": string(user.Role),
		"type": "access",
		"iat":  now.Unix(),
		"exp":  accessExp.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("sign access token: %w", err))
	}

	// Refresh token
	refreshClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"type": "refresh",
		"jti":  refreshJTI,
		"iat":  now.Unix(),
		"exp":  refreshExp.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("sign refresh token: %w", err))
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		AccessExp:    accessExp,
		RefreshExp:   refreshExp,
		RefreshJTI:   refreshJTI,
	}, nil
}

func (s *AuthService) parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) blacklistToken(ctx context.Context, jti string) error {
	key := refreshBlacklistPrefix + jti
	return s.redis.Set(ctx, key, "1", refreshTokenTTL).Err()
}

func (s *AuthService) isTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := refreshBlacklistPrefix + jti
	result, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// validatePassword enforces the password policy:
// min 8, max 72 (bcrypt limit), at least 1 uppercase, 1 lowercase, 1 digit.
func validatePassword(password string) *domain.AppError {
	if len(password) < 8 {
		return domain.ValidationError("password must be at least 8 characters", map[string]any{
			"field": "password",
			"rule":  "min_length",
		})
	}
	if len(password) > 72 {
		return domain.ValidationError("password must be at most 72 characters", map[string]any{
			"field": "password",
			"rule":  "max_length",
		})
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return domain.ValidationError("password must contain at least 1 uppercase letter, 1 lowercase letter, and 1 digit", map[string]any{
			"field": "password",
			"rule":  "complexity",
		})
	}

	return nil
}

// generateSecureToken creates a 32-byte hex-encoded random token.
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
