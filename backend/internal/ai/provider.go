// Package ai implements the LLM provider abstraction for CareerDock.
//
// It defines a common interface (LLMProvider) with implementations for
// Claude (primary) and OpenAI (fallback), plus a FallbackProvider that
// tries Claude first and falls back to OpenAI on failure.
//
// All AI operations are asynchronous — invoked from Asynq worker tasks,
// never from HTTP request handlers.
package ai

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LLMProvider abstracts the LLM backend (Claude, OpenAI).
type LLMProvider interface {
	// ParseResume extracts structured data from raw resume text.
	ParseResume(ctx context.Context, req *ParseResumeRequest) (*ParsedResume, error)

	// ScoreATSGeneral evaluates a resume's general ATS compatibility.
	ScoreATSGeneral(ctx context.Context, req *ATSGeneralRequest) (*ATSResult, error)

	// ScoreATSCompany evaluates resume fit against a specific company profile.
	ScoreATSCompany(ctx context.Context, req *ATSCompanyRequest) (*ATSResult, error)

	// ScoreATSJob evaluates resume fit against a specific job description.
	ScoreATSJob(ctx context.Context, req *ATSJobRequest) (*ATSResult, error)

	// CurateCompanyList ranks companies by fit for a candidate profile.
	CurateCompanyList(ctx context.Context, req *CurateListRequest) (*CuratedListResult, error)

	// EnrichCompany infers company attributes from its name and website.
	EnrichCompany(ctx context.Context, req *EnrichCompanyRequest) (*EnrichedCompany, error)

	// Name returns the provider name (e.g., "claude", "openai").
	Name() string
}

// --- Request types ---

// ParseResumeRequest contains the raw text extracted from a PDF.
type ParseResumeRequest struct {
	ResumeText string
}

// ATSGeneralRequest evaluates general ATS compatibility.
// PDFBytes is the primary input for Claude (native PDF processing).
// ResumeText is the fallback for OpenAI (text only).
type ATSGeneralRequest struct {
	PDFBytes   []byte
	ResumeText string
}

// --- Response types ---

// ParsedResume holds AI-extracted structured resume data.
type ParsedResume struct {
	Name              string           `json:"name"`
	Email             string           `json:"email,omitempty"`
	Phone             string           `json:"phone,omitempty"`
	Summary           string           `json:"summary"`
	YearsOfExperience float64          `json:"years_of_experience"`
	Skills            SkillSet         `json:"skills"`
	Experience        []WorkExperience `json:"experience"`
	Education         []Education      `json:"education"`
	Domains           []string         `json:"domains"`
	RoleLevel         string           `json:"role_level"`
	TokensUsed        TokenUsage       `json:"tokens_used"`
}

// SkillSet categorises technical skills.
type SkillSet struct {
	Languages  []string `json:"languages"`
	Frameworks []string `json:"frameworks"`
	Tools      []string `json:"tools"`
	Databases  []string `json:"databases"`
	Cloud      []string `json:"cloud"`
}

// WorkExperience represents a single work experience entry.
type WorkExperience struct {
	Company     string  `json:"company"`
	Title       string  `json:"title"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Description string  `json:"description"`
}

// Education represents an educational qualification.
type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Year        int    `json:"year"`
}

// ATSResult holds an ATS scoring result.
type ATSResult struct {
	Score       int                    `json:"score"`
	Breakdown   map[string]ScoreDetail `json:"breakdown"`
	Suggestions []string               `json:"suggestions"`
	TokensUsed  TokenUsage             `json:"tokens_used"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// ScoreDetail holds the score and feedback for a single ATS category.
type ScoreDetail struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// TokenUsage tracks token consumption for cost tracking.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// --- Types used by Sprint 4 (declared here for interface completeness) ---

// ATSCompanyRequest evaluates resume fit against a specific company.
type ATSCompanyRequest struct {
	PDFBytes   []byte
	ResumeText string
	Company    *CompanyProfile
}

// ATSJobRequest evaluates resume fit against a job description.
type ATSJobRequest struct {
	PDFBytes       []byte
	ResumeText     string
	JobDescription string
}

// CurateListRequest generates a ranked company list.
type CurateListRequest struct {
	ParsedResume *ParsedResume
	Companies    []CompanySummary
}

// CompanyProfile provides company data for ATS scoring.
type CompanyProfile struct {
	Name             string   `json:"name"`
	Size             string   `json:"size"`
	TechStack        []string `json:"tech_stack"`
	Domains          []string `json:"domains"`
	CompensationTier string   `json:"compensation_tier"`
}

// CompanySummary provides a company snapshot for curation.
type CompanySummary struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	TechStack        []string  `json:"tech_stack"`
	Domains          []string  `json:"domains"`
	Headquarters     string    `json:"headquarters"`
	Size             string    `json:"size"`
	CompensationTier string    `json:"compensation_tier"`
	HiringStatus     string    `json:"hiring_status"`
}

// CuratedListResult holds the AI-ranked list of best-fit companies.
type CuratedListResult struct {
	Companies   []RankedCompany `json:"companies"`
	GeneratedAt time.Time       `json:"generated_at"`
	TokensUsed  TokenUsage      `json:"tokens_used"`
}

// RankedCompany holds a single company's fit assessment within a curated list.
type RankedCompany struct {
	CompanyID      uuid.UUID `json:"company_id"`
	Name           string    `json:"name"`
	MatchScore     int       `json:"match_score"`
	MatchReasons   []string  `json:"match_reasons"`
	Recommendation string    `json:"recommendation"`
}

// EnrichCompanyRequest holds company info for AI enrichment.
type EnrichCompanyRequest struct {
	Name           string `json:"name"`
	CareersPageURL string `json:"careers_page_url,omitempty"`
	LinkedinURL    string `json:"linkedin_url,omitempty"`
}

// EnrichedCompany holds AI-inferred company attributes.
type EnrichedCompany struct {
	// Identity
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Headquarters   string `json:"headquarters"`
	FoundedYear    int    `json:"founded_year"`
	CareersPageURL string `json:"careers_page_url"`
	LinkedinURL    string `json:"linkedin_url"`

	// Profile
	Description  string   `json:"description"`
	TechStack    []string `json:"tech_stack"`
	Domains      []string `json:"domains"`
	Size         string   `json:"size"`
	HiringStatus string   `json:"hiring_status"`
	OfficeModes  []string `json:"office_modes"`

	// Compensation
	CompensationTier string `json:"compensation_tier"`

	TokensUsed TokenUsage `json:"tokens_used"`
}

// MarshalCuratedListResult serialises a CuratedListResult to JSON for storage.
func MarshalCuratedListResult(r *CuratedListResult) (json.RawMessage, error) {
	return json.Marshal(r)
}

// MarshalParsedResume serialises a ParsedResume to JSON for storage.
func MarshalParsedResume(pr *ParsedResume) (json.RawMessage, error) {
	return json.Marshal(pr)
}

// MarshalATSResult serialises an ATSResult to JSON for storage.
func MarshalATSResult(ar *ATSResult) (json.RawMessage, error) {
	return json.Marshal(ar)
}
