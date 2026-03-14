package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// UserService handles user profile and settings logic.
type UserService struct {
	users domain.UserRepository
	tx    domain.Transactor
}

// NewUserService creates a new UserService.
func NewUserService(
	users domain.UserRepository,
	tx domain.Transactor,
) *UserService {
	return &UserService{users: users, tx: tx}
}

// UpdateProfileInput holds input for profile updates.
type UpdateProfileInput struct {
	UserID              uuid.UUID
	Name                *string
	CurrentTitle        *string
	ExperienceLevel     *domain.ExperienceLevel
	PreferredTechStacks []string
	TargetDomains       []string
	TargetLocations     []string
}

// UpdateProfile updates user profile fields.
func (s *UserService) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if *input.Name == "" || len(*input.Name) > 255 {
			return nil, domain.ValidationError("name must be 1-255 characters", map[string]any{"field": "name"})
		}
		user.Name = *input.Name
	}
	if input.CurrentTitle != nil {
		user.CurrentTitle = input.CurrentTitle
	}
	if input.ExperienceLevel != nil {
		user.ExperienceLevel = input.ExperienceLevel
	}
	if input.PreferredTechStacks != nil {
		user.PreferredTechStacks = input.PreferredTechStacks
	}
	if input.TargetDomains != nil {
		user.TargetDomains = input.TargetDomains
	}
	if input.TargetLocations != nil {
		user.TargetLocations = input.TargetLocations
	}

	user.UpdatedAt = time.Now().UTC()

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePasswordInput holds input for changing password.
type ChangePasswordInput struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
}

// ChangePassword changes the user's password after verifying the current one.
func (s *UserService) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return domain.Unauthorized("current password is incorrect")
	}

	// Validate new password
	if pwErr := validatePassword(input.NewPassword); pwErr != nil {
		return pwErr
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcryptCost)
	if err != nil {
		return domain.InternalError(fmt.Errorf("hash password: %w", err))
	}

	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now().UTC()

	// TODO: Invalidate all other refresh tokens via Redis

	return s.users.Update(ctx, user)
}

// DeleteAccountInput holds input for account deletion.
type DeleteAccountInput struct {
	UserID   uuid.UUID
	Password string
}

// DeleteAccount soft-deletes the user after password confirmation.
func (s *UserService) DeleteAccount(ctx context.Context, input DeleteAccountInput) error {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return domain.Unauthorized("password is incorrect")
	}

	return s.users.SoftDelete(ctx, input.UserID)
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}
