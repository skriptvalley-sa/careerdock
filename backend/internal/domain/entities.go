package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a registered user account.
type User struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"` // never serialised
	Name          string     `json:"name"`
	Role          Role       `json:"role"`
	PremiumSince  *time.Time `json:"premium_since,omitempty"`
	EmailVerified bool       `json:"email_verified"`

	// Profile fields used for AI matching
	CurrentTitle        *string          `json:"current_title,omitempty"`
	ExperienceLevel     *ExperienceLevel `json:"experience_level,omitempty"`
	PreferredTechStacks []string         `json:"preferred_tech_stacks"`
	TargetDomains       []string         `json:"target_domains"`
	TargetLocations     []string         `json:"target_locations"`

	DefaultResumeID *uuid.UUID `json:"default_resume_id,omitempty"`

	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// IsPremium returns true if the user has purchased the Starter Pack.
func (u *User) IsPremium() bool {
	return u.PremiumSince != nil
}

// Company represents a tech company in the directory.
type Company struct {
	ID                uuid.UUID       `json:"id"`
	Slug              string          `json:"slug"`
	Name              string          `json:"name"`
	LogoURL           *string         `json:"logo_url,omitempty"`
	Description       *string         `json:"description,omitempty"`
	Size              *CompanySize    `json:"size,omitempty"`
	Headquarters      *string         `json:"headquarters,omitempty"`
	FoundedYear       *int            `json:"founded_year,omitempty"`
	CareersPageURL    *string         `json:"careers_page_url,omitempty"`
	GlassdoorURL      *string         `json:"glassdoor_url,omitempty"`
	AmbitionboxURL    *string         `json:"ambitionbox_url,omitempty"`
	LinkedinURL       *string         `json:"linkedin_url,omitempty"`
	TechStack         []string        `json:"tech_stack"`
	Domains           []string        `json:"domains"`
	HiringStatus      HiringStatus    `json:"hiring_status"`
	InterviewPatterns json.RawMessage `json:"interview_patterns,omitempty"`
	CompensationTier  *string         `json:"compensation_tier,omitempty"`
	HasRSU            bool            `json:"has_rsu"`
	HasRSURefresher   bool            `json:"has_rsu_refresher"`
	CompensationBands json.RawMessage `json:"compensation_bands,omitempty"`
	LastVerifiedAt    *time.Time      `json:"last_verified_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// CompanyEdit represents a moderator-submitted edit to a company profile.
type CompanyEdit struct {
	ID          uuid.UUID         `json:"id"`
	CompanyID   uuid.UUID         `json:"company_id"`
	SubmittedBy uuid.UUID         `json:"submitted_by"`
	ReviewedBy  *uuid.UUID        `json:"reviewed_by,omitempty"`
	Status      CompanyEditStatus `json:"status"`
	Changes     json.RawMessage   `json:"changes"`
	ReviewNotes *string           `json:"review_notes,omitempty"`
	ReviewedAt  *time.Time        `json:"reviewed_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// UserList represents a user-created company tracking list.
type UserList struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListEntry is a company + role pair within a list, with application tracking.
type ListEntry struct {
	ID          uuid.UUID         `json:"id"`
	ListID      uuid.UUID         `json:"list_id"`
	CompanyID   uuid.UUID         `json:"company_id"`
	RoleTitle   string            `json:"role_title"`
	Status      ApplicationStatus `json:"status"`
	DateApplied *time.Time        `json:"date_applied,omitempty"`
	Notes       *string           `json:"notes,omitempty"`
	Position    int               `json:"position"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// StatusHistory tracks every application status change.
type StatusHistory struct {
	ID          uuid.UUID          `json:"id"`
	ListEntryID uuid.UUID          `json:"list_entry_id"`
	FromStatus  *ApplicationStatus `json:"from_status,omitempty"`
	ToStatus    ApplicationStatus  `json:"to_status"`
	ChangedAt   time.Time          `json:"changed_at"`
}

// InterviewRound tracks a single interview round within an application.
type InterviewRound struct {
	ID            uuid.UUID        `json:"id"`
	ListEntryID   uuid.UUID        `json:"list_entry_id"`
	RoundNumber   int              `json:"round_number"`
	RoundType     string           `json:"round_type"`
	ScheduledDate *time.Time       `json:"scheduled_date,omitempty"`
	Outcome       InterviewOutcome `json:"outcome"`
	Notes         *string          `json:"notes,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// Resume represents a user's uploaded resume with extracted content.
type Resume struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	SlotNumber    int             `json:"slot_number"`
	FileName      string          `json:"file_name"`
	FileSizeBytes int             `json:"file_size_bytes"`
	S3Key         string          `json:"-"`
	ExtractedText *string         `json:"-"` // large text — excluded from default JSON
	ParsedData    json.RawMessage `json:"parsed_data,omitempty"`
	ATSGeneral    json.RawMessage `json:"ats_general,omitempty"`
	Status        ResumeStatus    `json:"status"`
	IsDefault     bool            `json:"is_default"`
	IsArchived    bool            `json:"is_archived"`
	ArchivedAt    *time.Time      `json:"archived_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ATSCheck represents an ATS scoring result (company or job).
type ATSCheck struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"user_id"`
	ResumeID       uuid.UUID       `json:"resume_id"`
	CheckType      ATSCheckType    `json:"check_type"`
	CompanyID      *uuid.UUID      `json:"company_id,omitempty"`
	JobDescription *string         `json:"job_description,omitempty"`
	Result         json.RawMessage `json:"result"`
	CacheKey       string          `json:"-"`
	CreatedAt      time.Time       `json:"created_at"`
}

// CuratedList represents an AI-generated company recommendation.
type CuratedList struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	ResumeID        uuid.UUID       `json:"resume_id"`
	PreferencesHash string          `json:"-"`
	Result          json.RawMessage `json:"result"`
	CreatedAt       time.Time       `json:"created_at"`
}

// Payment represents a Razorpay transaction.
type Payment struct {
	ID                uuid.UUID     `json:"id"`
	UserID            uuid.UUID     `json:"user_id"`
	RazorpayOrderID   string        `json:"razorpay_order_id"`
	RazorpayPaymentID *string       `json:"razorpay_payment_id,omitempty"`
	AmountPaise       int           `json:"amount_paise"`
	Currency          string        `json:"currency"`
	ProductType       ProductType   `json:"product_type"`
	Status            PaymentStatus `json:"status"`
	ReceiptNumber     *string       `json:"receipt_number,omitempty"`
	RefundReason      *string       `json:"refund_reason,omitempty"`
	RefundedAt        *time.Time    `json:"refunded_at,omitempty"`
	RefundedBy        *uuid.UUID    `json:"refunded_by,omitempty"`
	WebhookReceivedAt *time.Time    `json:"webhook_received_at,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// UserCredit represents a user's balance for a specific credit type.
type UserCredit struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	CreditType CreditType `json:"credit_type"`
	Balance    int        `json:"balance"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CreditTransaction is an immutable audit record for a credit change.
type CreditTransaction struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	CreditType   CreditType `json:"credit_type"`
	Amount       int        `json:"amount"` // positive = credit, negative = debit
	BalanceAfter int        `json:"balance_after"`
	Reason       string     `json:"reason"`
	ReferenceID  *uuid.UUID `json:"reference_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// FeatureFlag represents a platform-wide feature toggle.
type FeatureFlag struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Enabled     bool      `json:"enabled"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuditLogEntry represents an admin action in the audit trail.
type AuditLogEntry struct {
	ID         uuid.UUID       `json:"id"`
	AdminID    uuid.UUID       `json:"admin_id"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *uuid.UUID      `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// Notification represents a user notification.
type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Message   *string         `json:"message,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// EmailVerificationToken is a single-use token for email verification.
type EmailVerificationToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Token     string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// PasswordResetToken is a single-use token for password reset.
type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Token     string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
