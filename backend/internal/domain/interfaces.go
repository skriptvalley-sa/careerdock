package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// --- Repository interfaces ---

// UserRepository defines data access for user accounts.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// UndoSoftDelete restores a soft-deleted user (clears deleted_at).
	UndoSoftDelete(ctx context.Context, id uuid.UUID) error
	// HardDeleteExpired permanently deletes users soft-deleted before cutoff.
	// Returns the number of rows deleted.
	HardDeleteExpired(ctx context.Context, cutoff time.Time) (int64, error)
	// GetByIDIncludeDeleted retrieves a user by ID, including soft-deleted users.
	GetByIDIncludeDeleted(ctx context.Context, id uuid.UUID) (*User, error)
	// ListUsers returns users matching the filter (admin use).
	ListUsers(ctx context.Context, filter UserFilter) ([]User, int, error) // users, total count, error
}

// CompanyRepository defines data access for the company directory.
type CompanyRepository interface {
	List(ctx context.Context, filter CompanyFilter) ([]Company, string, error) // entities, next cursor, error
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetBySlug(ctx context.Context, slug string) (*Company, error)
	Search(ctx context.Context, query string, filter CompanyFilter) ([]Company, string, error)
	GetNamesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error)
	GetNameAndSlugsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]CompanyNameSlug, error)
	// ListAll returns all companies as compact summaries for AI curation (capped at 500).
	ListAll(ctx context.Context) ([]Company, error)
	Create(ctx context.Context, company *Company) error
	Update(ctx context.Context, company *Company) error
}

// ResumeRepository defines data access for resumes.
type ResumeRepository interface {
	Create(ctx context.Context, resume *Resume) error
	GetByID(ctx context.Context, id uuid.UUID) (*Resume, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Resume, error)
	Update(ctx context.Context, resume *Resume) error
	Archive(ctx context.Context, id uuid.UUID) error
	GetByUserAndSlot(ctx context.Context, userID uuid.UUID, slot int) (*Resume, error)
	ClearDefaultForUser(ctx context.Context, userID uuid.UUID) error
}

// ListRepository defines data access for user lists and entries.
type ListRepository interface {
	CreateList(ctx context.Context, list *UserList) error
	GetListByID(ctx context.Context, id uuid.UUID) (*UserList, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]UserList, error)
	UpdateList(ctx context.Context, list *UserList) error
	DeleteList(ctx context.Context, id uuid.UUID) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)

	CreateEntry(ctx context.Context, entry *ListEntry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*ListEntry, error)
	ListEntries(ctx context.Context, listID uuid.UUID) ([]ListEntry, error)
	ListEntriesByCompanyID(ctx context.Context, userID, companyID uuid.UUID) ([]ListEntryWithList, error)
	ListAllEntries(ctx context.Context, userID uuid.UUID, statusFilter *ApplicationStatus, excludeNotApplied bool) ([]ListEntryFull, error)
	ListsWithCompanyFlag(ctx context.Context, userID, companyID uuid.UUID) ([]ListCompanyFlag, error)
	CompanyListCounts(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error)
	UpdateEntry(ctx context.Context, entry *ListEntry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	DeleteEntryByCompany(ctx context.Context, listID, companyID uuid.UUID) error

	CreateStatusHistory(ctx context.Context, h *StatusHistory) error
	ListStatusHistory(ctx context.Context, entryID uuid.UUID) ([]StatusHistory, error)

	CreateInterviewRound(ctx context.Context, round *InterviewRound) error
	GetInterviewRoundByID(ctx context.Context, id uuid.UUID) (*InterviewRound, error)
	UpdateInterviewRound(ctx context.Context, round *InterviewRound) error
	DeleteInterviewRound(ctx context.Context, id uuid.UUID) error
}

// ATSCheckRepository defines data access for ATS check results.
type ATSCheckRepository interface {
	Create(ctx context.Context, check *ATSCheck) error
	GetByID(ctx context.Context, id uuid.UUID) (*ATSCheck, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]ATSCheck, error)
	GetByCacheKey(ctx context.Context, cacheKey string) (*ATSCheck, error)
	// UpdateResult stores the AI-generated score result for a completed ATS check.
	UpdateResult(ctx context.Context, id uuid.UUID, result json.RawMessage) error
}

// CuratedListRepository defines data access for AI-curated lists.
type CuratedListRepository interface {
	Create(ctx context.Context, list *CuratedList) error
	GetByID(ctx context.Context, id uuid.UUID) (*CuratedList, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]CuratedList, error)
	// GetByPreferencesHash returns the most recent list with the given hash, or nil if not found.
	GetByPreferencesHash(ctx context.Context, hash string) (*CuratedList, error)
	// UpdateResult stores the AI-generated ranking result for a completed curated list.
	UpdateResult(ctx context.Context, id uuid.UUID, result json.RawMessage) error
}

// PaymentRepository defines data access for payment records.
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error
	UpdateWebhookCapture(ctx context.Context, id uuid.UUID, razorpayPaymentID string, webhookReceivedAt time.Time) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Payment, error)
	// ListAll returns payments matching the filter (admin use).
	ListAll(ctx context.Context, filter PaymentFilter) ([]Payment, int, error) // payments, total count, error
}

// CreditRepository defines data access for credits and transactions.
type CreditRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID, creditType CreditType) (int, error)
	GetAllBalances(ctx context.Context, userID uuid.UUID) (map[CreditType]int, error)
	Allocate(ctx context.Context, userID uuid.UUID, creditType CreditType, amount int) error
	Deduct(ctx context.Context, userID uuid.UUID, creditType CreditType, amount int) error
	LogTransaction(ctx context.Context, txn *CreditTransaction) error
	ListTransactionsByUser(ctx context.Context, userID uuid.UUID, creditType *CreditType, limit int) ([]CreditTransaction, error)
	// ListAllTransactions returns credit transactions matching the filter (admin use).
	ListAllTransactions(ctx context.Context, filter CreditTransactionFilter) ([]CreditTransaction, int, error)
}

// NotificationRepository defines data access for user notifications.
type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
}

// FeatureFlagRepository defines data access for feature flags.
type FeatureFlagRepository interface {
	GetByKey(ctx context.Context, key string) (*FeatureFlag, error)
	List(ctx context.Context) ([]FeatureFlag, error)
	Update(ctx context.Context, flag *FeatureFlag) error
}

// AuditLogRepository defines data access for the admin audit log.
type AuditLogRepository interface {
	Create(ctx context.Context, entry *AuditLogEntry) error
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLogEntry, error)
}

// --- Transaction support ---

// Transactor wraps multiple repository calls in a database transaction.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// --- External service interfaces ---

// AIProvider is defined in internal/ai.LLMProvider.
// Keeping a comment here for cross-reference. The concrete interface lives
// in the ai package to avoid circular imports with the prompt/type layer.

// PaymentGateway abstracts the payment provider (Razorpay).
type PaymentGateway interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)
	VerifyPayment(ctx context.Context, req *VerifyPaymentRequest) (*PaymentVerification, error)
}

// EmailSender abstracts transactional email delivery.
type EmailSender interface {
	Send(ctx context.Context, msg *EmailMessage) error
}

// FileStore abstracts object storage (S3/MinIO).
type FileStore interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) error
	Download(ctx context.Context, key string) ([]byte, error)
	GenerateSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// --- Filter types ---

// CompanyFilter defines filtering options for company listing.
type CompanyFilter struct {
	Cursor            string
	Limit             int
	Sizes             []CompanySize // filter by one or more sizes
	HiringStatus      *HiringStatus // single hiring status filter
	TechStack         []string      // must have ALL (AND / @> operator)
	Domains           []string      // any match (OR / && operator)
	CompensationTiers []string      // filter by one or more tiers
	HasRSU            *bool         // filter by RSU availability
	Headquarters      string        // partial match (ILIKE)
	Sort              string        // name, size, compensation_tier, updated_at
	Order             string        // asc (default), desc
}

// UserFilter defines filtering options for admin user listing.
type UserFilter struct {
	Query  string // search by name or email (ILIKE)
	Role   *Role
	Limit  int
	Offset int
}

// PaymentFilter defines filtering options for admin payment listing.
type PaymentFilter struct {
	UserID *uuid.UUID
	Status *PaymentStatus
	Limit  int
	Offset int
}

// CreditTransactionFilter defines filtering options for admin credit transaction listing.
type CreditTransactionFilter struct {
	UserID     *uuid.UUID
	CreditType *CreditType
	Limit      int
	Offset     int
}

// AuditLogFilter defines filtering options for admin audit log.
type AuditLogFilter struct {
	AdminID    *uuid.UUID
	EntityType *string
	EntityID   *uuid.UUID
	Limit      int
	Offset     int
}

// Note: AI request/response types are defined in internal/ai/provider.go.
// The domain.AIProvider interface above uses those types indirectly — the
// actual LLM abstraction is internal/ai.LLMProvider. The domain-level
// AIProvider is kept for reference but may be removed in a future cleanup.

// --- Payment request/response types (placeholders — fleshed out in Sprint 3) ---

// CreateOrderRequest holds inputs for creating a payment order.
type CreateOrderRequest struct {
	UserID      uuid.UUID
	AmountPaise int
	ProductType ProductType
}

// Order represents a created payment order.
type Order struct {
	RazorpayOrderID string
	AmountPaise     int
	Currency        string
}

// VerifyPaymentRequest holds inputs for verifying a payment.
type VerifyPaymentRequest struct {
	OrderID   string
	PaymentID string
	Signature string
}

// PaymentVerification holds the result of payment verification.
type PaymentVerification struct {
	Verified bool
}

// --- Email types ---

// EmailMessage represents a transactional email.
type EmailMessage struct {
	To      string
	Subject string
	HTML    string
}
