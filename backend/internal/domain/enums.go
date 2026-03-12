// Package domain defines business entities, interfaces, errors, and enum
// constants. It is the dependency-inversion hub — every other package depends
// on domain, but domain depends on nothing except stdlib.
package domain

// Role represents a user's permission level.
type Role string

// Role constants.
const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

// ApplicationStatus represents a list entry's application pipeline stage.
type ApplicationStatus string

// ApplicationStatus constants.
const (
	StatusNotApplied  ApplicationStatus = "not_applied"
	StatusApplied     ApplicationStatus = "applied"
	StatusPhoneScreen ApplicationStatus = "phone_screen"
	StatusInterview   ApplicationStatus = "interview"
	StatusOffer       ApplicationStatus = "offer"
	StatusRejected    ApplicationStatus = "rejected"
	StatusAccepted    ApplicationStatus = "accepted"
	StatusWithdrawn   ApplicationStatus = "withdrawn"
)

// CreditType represents a category of purchasable credit.
type CreditType string

// CreditType constants.
const (
	CreditResumeUpload CreditType = "resume_upload"
	CreditATSCheck     CreditType = "ats_check"
	CreditCuratedList  CreditType = "curated_list"
	CreditCVGeneration CreditType = "cv_generation"
)

// ResumeStatus represents the processing state of a resume.
type ResumeStatus string

// ResumeStatus constants.
const (
	ResumeStatusUploading  ResumeStatus = "uploading"
	ResumeStatusExtracting ResumeStatus = "extracting"
	ResumeStatusParsing    ResumeStatus = "parsing"
	ResumeStatusReady      ResumeStatus = "ready"
	ResumeStatusFailed     ResumeStatus = "failed"
)

// PaymentStatus represents the lifecycle of a payment.
type PaymentStatus string

// PaymentStatus constants.
const (
	PaymentStatusCreated  PaymentStatus = "created"
	PaymentStatusCaptured PaymentStatus = "captured"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// ProductType represents a purchasable product.
type ProductType string

// ProductType constants.
const (
	ProductStarterPack  ProductType = "starter_pack"
	ProductResumeUpload ProductType = "resume_upload"
	ProductATSBundle    ProductType = "ats_bundle"
	ProductCVGeneration ProductType = "cv_generation"
	ProductRebuyPack    ProductType = "rebuy_pack"
)

// ATSCheckType represents the type of ATS check.
type ATSCheckType string

// ATSCheckType constants.
const (
	ATSCheckCompany ATSCheckType = "company"
	ATSCheckJob     ATSCheckType = "job"
)

// CompanySize represents the scale of a company.
type CompanySize string

// CompanySize constants.
const (
	CompanySizeStartup    CompanySize = "startup"
	CompanySizeSmall      CompanySize = "small"
	CompanySizeMid        CompanySize = "mid"
	CompanySizeLarge      CompanySize = "large"
	CompanySizeEnterprise CompanySize = "enterprise"
)

// HiringStatus represents whether a company is actively hiring.
type HiringStatus string

// HiringStatus constants.
const (
	HiringActive  HiringStatus = "active"
	HiringPaused  HiringStatus = "paused"
	HiringUnknown HiringStatus = "unknown"
)

// ExperienceLevel represents a user's seniority level.
type ExperienceLevel string

// ExperienceLevel constants.
const (
	ExpFresher   ExperienceLevel = "fresher"
	ExpJunior    ExperienceLevel = "junior"
	ExpMid       ExperienceLevel = "mid"
	ExpSenior    ExperienceLevel = "senior"
	ExpStaffPlus ExperienceLevel = "staff_plus"
)

// CompanyEditStatus represents the moderation state of a company edit.
type CompanyEditStatus string

// CompanyEditStatus constants.
const (
	EditStatusPending  CompanyEditStatus = "pending"
	EditStatusApproved CompanyEditStatus = "approved"
	EditStatusRejected CompanyEditStatus = "rejected"
)

// InterviewOutcome represents the result of an interview round.
type InterviewOutcome string

// InterviewOutcome constants.
const (
	OutcomePassed  InterviewOutcome = "passed"
	OutcomeFailed  InterviewOutcome = "failed"
	OutcomePending InterviewOutcome = "pending"
)
