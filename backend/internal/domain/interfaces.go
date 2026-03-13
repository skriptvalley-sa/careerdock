package domain

import (
	"context"
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
}

// CompanyRepository defines data access for the company directory.
type CompanyRepository interface {
	List(ctx context.Context, filter CompanyFilter) ([]Company, string, error) // entities, next cursor, error
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetBySlug(ctx context.Context, slug string) (*Company, error)
	Search(ctx context.Context, query string, filter CompanyFilter) ([]Company, string, error)
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
	UpdateEntry(ctx context.Context, entry *ListEntry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error
}

// ATSCheckRepository defines data access for ATS check results.
type ATSCheckRepository interface {
	Create(ctx context.Context, check *ATSCheck) error
	GetByID(ctx context.Context, id uuid.UUID) (*ATSCheck, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]ATSCheck, error)
	GetByCacheKey(ctx context.Context, cacheKey string) (*ATSCheck, error)
}

// CuratedListRepository defines data access for AI-curated lists.
type CuratedListRepository interface {
	Create(ctx context.Context, list *CuratedList) error
	GetByID(ctx context.Context, id uuid.UUID) (*CuratedList, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]CuratedList, error)
}

// PaymentRepository defines data access for payment records.
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Payment, error)
}

// CreditRepository defines data access for credits and transactions.
type CreditRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID, creditType CreditType) (int, error)
	Allocate(ctx context.Context, userID uuid.UUID, creditType CreditType, amount int) error
	Deduct(ctx context.Context, userID uuid.UUID, creditType CreditType, amount int) error
	LogTransaction(ctx context.Context, txn *CreditTransaction) error
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

// AIProvider abstracts LLM operations behind a common interface.
type AIProvider interface {
	ParseResume(ctx context.Context, req *ParseResumeRequest) (*ParsedResume, error)
	ScoreATSGeneral(ctx context.Context, req *ATSGeneralRequest) (*ATSResult, error)
	ScoreATSCompany(ctx context.Context, req *ATSCompanyRequest) (*ATSResult, error)
	ScoreATSJob(ctx context.Context, req *ATSJobRequest) (*ATSResult, error)
	CurateCompanyList(ctx context.Context, req *CurateListRequest) (*CuratedListResult, error)
	EnrichCompanyProfile(ctx context.Context, req *EnrichRequest) (*CompanyProfile, error)
}

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

// AuditLogFilter defines filtering options for admin audit log.
type AuditLogFilter struct {
	AdminID    *uuid.UUID
	EntityType *string
	EntityID   *uuid.UUID
	Limit      int
	Offset     int
}

// --- AI request/response types (placeholders — fleshed out in Sprint 3) ---

// ParseResumeRequest holds inputs for resume parsing.
type ParseResumeRequest struct {
	ResumeText string
}

// ParsedResume holds AI-extracted structured resume data.
type ParsedResume struct {
	Data []byte // raw JSON
}

// ATSGeneralRequest holds inputs for general ATS scoring.
type ATSGeneralRequest struct {
	ResumeText string
}

// ATSCompanyRequest holds inputs for company-specific ATS scoring.
type ATSCompanyRequest struct {
	ResumeText string
	Company    *Company
}

// ATSJobRequest holds inputs for job-specific ATS scoring.
type ATSJobRequest struct {
	ResumeText     string
	JobDescription string
}

// ATSResult holds an ATS scoring result.
type ATSResult struct {
	Data []byte // raw JSON
}

// CurateListRequest holds inputs for AI-curated company list generation.
type CurateListRequest struct {
	ResumeText  string
	Preferences *User // user profile + prefs
}

// CuratedListResult holds the AI-generated company recommendations.
type CuratedListResult struct {
	Data []byte // raw JSON
}

// EnrichRequest holds inputs for AI company profile enrichment.
type EnrichRequest struct {
	Company *Company
}

// CompanyProfile holds enriched company data.
type CompanyProfile struct {
	Data []byte // raw JSON
}

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
