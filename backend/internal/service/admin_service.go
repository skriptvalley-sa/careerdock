package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// AdminService handles admin panel business logic.
type AdminService struct {
	companies   domain.CompanyRepository
	users       domain.UserRepository
	credits     domain.CreditRepository
	payments    domain.PaymentRepository
	auditLog    domain.AuditLogRepository
	store       domain.FileStore
	tx          domain.Transactor
	redisClient *redis.Client
}

// NewAdminService creates a new AdminService.
//
//nolint:revive // parameter list is necessarily long — this is the DI constructor
func NewAdminService(
	companies domain.CompanyRepository,
	users domain.UserRepository,
	credits domain.CreditRepository,
	payments domain.PaymentRepository,
	auditLog domain.AuditLogRepository,
	store domain.FileStore,
	tx domain.Transactor,
	redisClient *redis.Client,
) *AdminService {
	return &AdminService{
		companies:   companies,
		users:       users,
		credits:     credits,
		payments:    payments,
		auditLog:    auditLog,
		store:       store,
		tx:          tx,
		redisClient: redisClient,
	}
}

// --- 5.1: Admin Company CRUD ---

// CreateCompanyInput holds input for admin company creation.
type CreateCompanyInput struct {
	Slug              string
	Name              string
	LogoURL           *string
	Description       *string
	Size              *domain.CompanySize
	Headquarters      *string
	FoundedYear       *int
	CareersPageURL    *string
	GlassdoorURL      *string
	AmbitionboxURL    *string
	LinkedinURL       *string
	TechStack         []string
	Domains           []string
	HiringStatus      domain.HiringStatus
	InterviewPatterns json.RawMessage
	CompensationTier  *string
	HasRSU            bool
	HasRSURefresher   bool
	OfficeModes       []string
	CompensationBands json.RawMessage
}

// CreateCompany creates a new company and logs the action.
func (s *AdminService) CreateCompany(ctx context.Context, adminID uuid.UUID, input CreateCompanyInput, ipAddress string) (*domain.Company, error) {
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
	input.Name = strings.TrimSpace(input.Name)

	if input.Slug == "" {
		return nil, domain.ValidationError("slug is required", map[string]any{"field": "slug"})
	}
	if input.Name == "" {
		return nil, domain.ValidationError("name is required", map[string]any{"field": "name"})
	}

	company := &domain.Company{
		ID:                uuid.Must(uuid.NewV7()),
		Slug:              input.Slug,
		Name:              input.Name,
		LogoURL:           input.LogoURL,
		Description:       input.Description,
		Size:              input.Size,
		Headquarters:      input.Headquarters,
		FoundedYear:       input.FoundedYear,
		CareersPageURL:    input.CareersPageURL,
		GlassdoorURL:      input.GlassdoorURL,
		AmbitionboxURL:    input.AmbitionboxURL,
		LinkedinURL:       input.LinkedinURL,
		TechStack:         input.TechStack,
		Domains:           input.Domains,
		HiringStatus:      input.HiringStatus,
		InterviewPatterns: input.InterviewPatterns,
		CompensationTier:  input.CompensationTier,
		HasRSU:            input.HasRSU,
		HasRSURefresher:   input.HasRSURefresher,
		OfficeModes:       input.OfficeModes,
		CompensationBands: input.CompensationBands,
	}

	if err := s.companies.Create(ctx, company); err != nil {
		return nil, err
	}

	s.logAudit(ctx, adminID, "create", "company", &company.ID, nil, ipAddress)
	return company, nil
}

// UpdateCompanyInput holds input for admin company updates.
type UpdateCompanyInput struct {
	Name              *string
	Slug              *string
	LogoURL           *string
	Description       *string
	Size              *domain.CompanySize
	Headquarters      *string
	FoundedYear       *int
	CareersPageURL    *string
	GlassdoorURL      *string
	AmbitionboxURL    *string
	LinkedinURL       *string
	TechStack         []string
	Domains           []string
	HiringStatus      *domain.HiringStatus
	InterviewPatterns json.RawMessage
	CompensationTier  *string
	HasRSU            *bool
	HasRSURefresher   *bool
	OfficeModes       []string
	CompensationBands json.RawMessage
	LastVerifiedAt    *time.Time
}

// UpdateCompany updates an existing company and logs the action.
func (s *AdminService) UpdateCompany(ctx context.Context, adminID uuid.UUID, companyID uuid.UUID, input UpdateCompanyInput, ipAddress string) (*domain.Company, error) {
	company, err := s.companies.GetByID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.ValidationError("name cannot be empty", map[string]any{"field": "name"})
		}
		company.Name = name
	}
	if input.Slug != nil {
		slug := strings.TrimSpace(strings.ToLower(*input.Slug))
		if slug == "" {
			return nil, domain.ValidationError("slug cannot be empty", map[string]any{"field": "slug"})
		}
		company.Slug = slug
	}
	if input.LogoURL != nil {
		company.LogoURL = input.LogoURL
	}
	if input.Description != nil {
		company.Description = input.Description
	}
	if input.Size != nil {
		company.Size = input.Size
	}
	if input.Headquarters != nil {
		company.Headquarters = input.Headquarters
	}
	if input.FoundedYear != nil {
		company.FoundedYear = input.FoundedYear
	}
	if input.CareersPageURL != nil {
		company.CareersPageURL = input.CareersPageURL
	}
	if input.GlassdoorURL != nil {
		company.GlassdoorURL = input.GlassdoorURL
	}
	if input.AmbitionboxURL != nil {
		company.AmbitionboxURL = input.AmbitionboxURL
	}
	if input.LinkedinURL != nil {
		company.LinkedinURL = input.LinkedinURL
	}
	if input.TechStack != nil {
		company.TechStack = input.TechStack
	}
	if input.Domains != nil {
		company.Domains = input.Domains
	}
	if input.HiringStatus != nil {
		company.HiringStatus = *input.HiringStatus
	}
	if input.InterviewPatterns != nil {
		company.InterviewPatterns = input.InterviewPatterns
	}
	if input.CompensationTier != nil {
		company.CompensationTier = input.CompensationTier
	}
	if input.HasRSU != nil {
		company.HasRSU = *input.HasRSU
	}
	if input.HasRSURefresher != nil {
		company.HasRSURefresher = *input.HasRSURefresher
	}
	if input.OfficeModes != nil {
		company.OfficeModes = input.OfficeModes
	}
	if input.CompensationBands != nil {
		company.CompensationBands = input.CompensationBands
	}
	if input.LastVerifiedAt != nil {
		company.LastVerifiedAt = input.LastVerifiedAt
	}

	if err := s.companies.Update(ctx, company); err != nil {
		return nil, err
	}

	s.logAudit(ctx, adminID, "update", "company", &companyID, nil, ipAddress)
	return company, nil
}

// DeleteCompany hard-deletes a company from the directory and logs the action.
func (s *AdminService) DeleteCompany(ctx context.Context, adminID uuid.UUID, companyID uuid.UUID, ipAddress string) error {
	if err := s.companies.Delete(ctx, companyID); err != nil {
		return err
	}
	s.logAudit(ctx, adminID, "delete", "company", &companyID, nil, ipAddress)
	return nil
}

// UploadCompanyLogo uploads a logo to S3 and returns the key.
func (s *AdminService) UploadCompanyLogo(ctx context.Context, adminID uuid.UUID, companyID uuid.UUID, data []byte, contentType string, ipAddress string) (string, error) {
	// Validate company exists
	if _, err := s.companies.GetByID(ctx, companyID); err != nil {
		return "", err
	}

	key := "logos/" + companyID.String()
	if err := s.store.Upload(ctx, key, data, contentType); err != nil {
		return "", domain.InternalError(err)
	}

	s.logAudit(ctx, adminID, "upload_logo", "company", &companyID, nil, ipAddress)
	return key, nil
}

// --- 5.2: Admin User Management ---

// ListUsers returns users matching the filter.
func (s *AdminService) ListUsers(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	return s.users.ListUsers(ctx, filter)
}

// AdminUpdateUserInput holds input for admin user updates.
type AdminUpdateUserInput struct {
	Role         *domain.Role
	PremiumSince *time.Time // set to grant premium, nil to revoke
	SetPremium   *bool      // explicit flag: true = grant, false = revoke
	Banned       *bool      // true = soft-delete (ban), false = undo (unban)
}

// UpdateUser updates a user's admin-managed fields and logs the action.
func (s *AdminService) UpdateUser(ctx context.Context, adminID uuid.UUID, userID uuid.UUID, input AdminUpdateUserInput, ipAddress string) (*domain.User, error) {
	// Use GetByIDIncludeDeleted to support unbanning
	user, err := s.users.GetByIDIncludeDeleted(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Handle ban/unban
	if input.Banned != nil {
		if *input.Banned {
			// Ban = soft delete
			if user.DeletedAt == nil {
				if err := s.users.SoftDelete(ctx, userID); err != nil {
					return nil, err
				}
				s.logAudit(ctx, adminID, "ban", "user", &userID, nil, ipAddress)
			}
			// Re-fetch after soft delete
			user, err = s.users.GetByIDIncludeDeleted(ctx, userID)
			if err != nil {
				return nil, err
			}
			return user, nil
		}
		// Unban = undo soft delete
		if user.DeletedAt != nil {
			if err := s.users.UndoSoftDelete(ctx, userID); err != nil {
				return nil, err
			}
			s.logAudit(ctx, adminID, "unban", "user", &userID, nil, ipAddress)
			user.DeletedAt = nil
		}
	}

	// Refetch active user for further updates
	if user.DeletedAt != nil {
		return nil, domain.ValidationError("cannot update a banned user — unban first", nil)
	}

	changed := false
	if input.Role != nil {
		user.Role = *input.Role
		changed = true
	}
	if input.SetPremium != nil {
		if *input.SetPremium {
			if user.PremiumSince == nil {
				now := time.Now().UTC()
				user.PremiumSince = &now
				changed = true
			}
		} else {
			if user.PremiumSince != nil {
				user.PremiumSince = nil
				changed = true
			}
		}
	}

	if changed {
		if err := s.users.Update(ctx, user); err != nil {
			return nil, err
		}
		s.logAudit(ctx, adminID, "update", "user", &userID, nil, ipAddress)
	}

	return user, nil
}

// --- 5.3: Admin Credit Management ---

// AdminAllocateCreditsInput holds input for manual credit allocation.
type AdminAllocateCreditsInput struct {
	UserID     uuid.UUID
	CreditType domain.CreditType
	Amount     int
	Reason     string
}

// AllocateCredits manually allocates credits to a user with audit logging.
func (s *AdminService) AllocateCredits(ctx context.Context, adminID uuid.UUID, input AdminAllocateCreditsInput, ipAddress string) error {
	if input.Amount <= 0 {
		return domain.ValidationError("amount must be positive", map[string]any{"field": "amount"})
	}
	if input.Reason == "" {
		return domain.ValidationError("reason is required", map[string]any{"field": "reason"})
	}

	// Verify user exists
	if _, err := s.users.GetByID(ctx, input.UserID); err != nil {
		return err
	}

	err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.credits.Allocate(txCtx, input.UserID, input.CreditType, input.Amount); err != nil {
			return err
		}

		newBalance, err := s.credits.GetBalance(txCtx, input.UserID, input.CreditType)
		if err != nil {
			return err
		}

		txn := &domain.CreditTransaction{
			ID:           uuid.Must(uuid.NewV7()),
			UserID:       input.UserID,
			CreditType:   input.CreditType,
			Amount:       input.Amount,
			BalanceAfter: newBalance,
			Reason:       "admin: " + input.Reason,
			CreatedAt:    time.Now().UTC(),
		}
		if err := s.credits.LogTransaction(txCtx, txn); err != nil {
			return err
		}

		details, _ := json.Marshal(map[string]any{
			"user_id":     input.UserID,
			"credit_type": input.CreditType,
			"amount":      input.Amount,
			"reason":      input.Reason,
		})
		s.logAudit(txCtx, adminID, "allocate_credits", "user", &input.UserID, details, ipAddress)
		return nil
	})
	if err != nil {
		return err
	}

	// Notify the target user's connected browser via SSE so the credit widget
	// updates immediately without requiring a hard refresh.
	s.publishCreditsUpdated(input.UserID, string(input.CreditType))
	return nil
}

// publishCreditsUpdated fires a credits_updated SSE event to the target user.
func (s *AdminService) publishCreditsUpdated(userID uuid.UUID, creditType string) {
	if s.redisClient == nil {
		return
	}

	payload, err := json.Marshal(map[string]any{
		"credit_type": creditType,
	})
	if err != nil {
		slog.Warn("admin: failed to marshal credits_updated SSE payload", "error", err)
		return
	}

	event, err := json.Marshal(map[string]any{
		"type": "credits_updated",
		"data": json.RawMessage(payload),
	})
	if err != nil {
		slog.Warn("admin: failed to marshal credits_updated SSE event", "error", err)
		return
	}

	channel := fmt.Sprintf("sse:user:%s", userID)
	if pubErr := s.redisClient.Publish(context.Background(), channel, event).Err(); pubErr != nil {
		slog.Warn("admin: failed to publish credits_updated SSE", "error", pubErr, "user_id", userID)
	}
}

// --- 5.4: Admin Payment & Transaction Logs ---

// ListPayments returns payments matching the filter (admin use).
func (s *AdminService) ListPayments(ctx context.Context, filter domain.PaymentFilter) ([]domain.Payment, int, error) {
	return s.payments.ListAll(ctx, filter)
}

// ListCreditTransactions returns credit transactions matching the filter (admin use).
func (s *AdminService) ListCreditTransactions(ctx context.Context, filter domain.CreditTransactionFilter) ([]domain.CreditTransaction, int, error) {
	return s.credits.ListAllTransactions(ctx, filter)
}

// --- Audit log helper ---

func (s *AdminService) logAudit(ctx context.Context, adminID uuid.UUID, action, entityType string, entityID *uuid.UUID, details json.RawMessage, ipAddress string) {
	var ip *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	entry := &domain.AuditLogEntry{
		ID:         uuid.Must(uuid.NewV7()),
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		IPAddress:  ip,
		CreatedAt:  time.Now().UTC(),
	}
	// Best-effort audit logging — don't fail the request if audit fails
	_ = s.auditLog.Create(ctx, entry)
}
