package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/pdf"
)

const (
	taskATSCompanyCheck = "ats:company_check"
	taskATSJobCheck     = "ats:job_check"
	taskATSResumeCheck  = "ats:resume_check"

	minJobDescriptionLength = 100
	maxJobDescriptionLength = 10000
)

// ATSService handles ATS check requests (company, job, and resume-only variants).
type ATSService struct {
	atsRepo     domain.ATSCheckRepository
	resumeRepo  domain.ResumeRepository
	companyRepo domain.CompanyRepository
	creditRepo  domain.CreditRepository
	fileStore   domain.FileStore
	txr         domain.Transactor
	asynq       *asynq.Client
}

// NewATSService creates a new ATSService.
func NewATSService(
	atsRepo domain.ATSCheckRepository,
	resumeRepo domain.ResumeRepository,
	companyRepo domain.CompanyRepository,
	creditRepo domain.CreditRepository,
	fileStore domain.FileStore,
	txr domain.Transactor,
	asynqClient *asynq.Client,
) *ATSService {
	return &ATSService{
		atsRepo:     atsRepo,
		resumeRepo:  resumeRepo,
		companyRepo: companyRepo,
		creditRepo:  creditRepo,
		fileStore:   fileStore,
		txr:         txr,
		asynq:       asynqClient,
	}
}

// CheckCompany creates a company-specific ATS check, deducts a credit, and enqueues the worker task.
//
// Pipeline:
//  1. Verify resume exists, belongs to user, and is ready
//  2. Verify company exists
//  3. Compute deduplication cache key
//  4. Return existing check if cache hit (no credit charge)
//  5. Check ats_check credit balance
//  6. Create ATSCheck record (result = {})
//  7. Deduct credit
//  8. Enqueue ats:company_check worker task
func (s *ATSService) CheckCompany(ctx context.Context, userID, resumeID, companyID uuid.UUID) (*domain.ATSCheck, error) {
	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, domain.NotFound("resume", resumeID)
	}
	if resume.Status != domain.ResumeStatusReady {
		return nil, domain.ValidationError("Resume is not ready for ATS scoring", map[string]any{
			"resume_id": resumeID,
			"status":    resume.Status,
		})
	}

	if _, err := s.companyRepo.GetByID(ctx, companyID); err != nil {
		return nil, err
	}

	cacheKey := hashCacheKey(resumeID.String() + ":" + companyID.String())

	existing, err := s.atsRepo.GetByCacheKey(ctx, cacheKey)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("cache key lookup: %w", err))
	}
	if existing != nil {
		slog.Info("returning cached ATS company check", "check_id", existing.ID)
		return existing, nil
	}

	balance, err := s.creditRepo.GetBalance(ctx, userID, domain.CreditATSCheck)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("check credit balance: %w", err))
	}
	if balance < 1 {
		return nil, domain.InsufficientCredits(domain.CreditATSCheck)
	}

	checkID := uuid.Must(uuid.NewV7())
	check := &domain.ATSCheck{
		ID:        checkID,
		UserID:    userID,
		ResumeID:  &resumeID,
		CheckType: domain.ATSCheckCompany,
		CompanyID: &companyID,
		Result:    json.RawMessage("{}"),
		CacheKey:  cacheKey,
	}

	if err := s.atsRepo.Create(ctx, check); err != nil {
		return nil, err
	}

	if err := s.deductATSCredit(ctx, userID, checkID, "ATS check — company"); err != nil {
		slog.Error("failed to deduct ATS credit after check creation",
			"user_id", userID, "check_id", checkID, "error", err)
	}

	if err := s.enqueueTask(taskATSCompanyCheck, checkID); err != nil {
		slog.Error("failed to enqueue ATS company check task",
			"check_id", checkID, "error", err)
	}

	return check, nil
}

// CheckJob creates a job-description ATS check, deducts a credit, and enqueues the worker task.
func (s *ATSService) CheckJob(ctx context.Context, userID, resumeID uuid.UUID, jobDescription string) (*domain.ATSCheck, error) {
	if len(jobDescription) < minJobDescriptionLength {
		return nil, domain.ValidationError(
			fmt.Sprintf("Job description must be at least %d characters", minJobDescriptionLength),
			map[string]any{"length": len(jobDescription)},
		)
	}
	if len(jobDescription) > maxJobDescriptionLength {
		return nil, domain.ValidationError(
			fmt.Sprintf("Job description must not exceed %d characters", maxJobDescriptionLength),
			map[string]any{"length": len(jobDescription)},
		)
	}

	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, domain.NotFound("resume", resumeID)
	}
	if resume.Status != domain.ResumeStatusReady {
		return nil, domain.ValidationError("Resume is not ready for ATS scoring", map[string]any{
			"resume_id": resumeID,
			"status":    resume.Status,
		})
	}

	jdHash := hashCacheKey(jobDescription)
	cacheKey := hashCacheKey(resumeID.String() + ":job:" + jdHash)

	existing, err := s.atsRepo.GetByCacheKey(ctx, cacheKey)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("cache key lookup: %w", err))
	}
	if existing != nil {
		slog.Info("returning cached ATS job check", "check_id", existing.ID)
		return existing, nil
	}

	balance, err := s.creditRepo.GetBalance(ctx, userID, domain.CreditATSCheck)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("check credit balance: %w", err))
	}
	if balance < 1 {
		return nil, domain.InsufficientCredits(domain.CreditATSCheck)
	}

	checkID := uuid.Must(uuid.NewV7())
	check := &domain.ATSCheck{
		ID:             checkID,
		UserID:         userID,
		ResumeID:       &resumeID,
		CheckType:      domain.ATSCheckJob,
		JobDescription: &jobDescription,
		Result:         json.RawMessage("{}"),
		CacheKey:       cacheKey,
	}

	if err := s.atsRepo.Create(ctx, check); err != nil {
		return nil, err
	}

	if err := s.deductATSCredit(ctx, userID, checkID, "ATS check — job"); err != nil {
		slog.Error("failed to deduct ATS credit after check creation",
			"user_id", userID, "check_id", checkID, "error", err)
	}

	if err := s.enqueueTask(taskATSJobCheck, checkID); err != nil {
		slog.Error("failed to enqueue ATS job check task",
			"check_id", checkID, "error", err)
	}

	return check, nil
}

// CheckResume creates a resume-only ATS check (no company or job description).
func (s *ATSService) CheckResume(ctx context.Context, userID, resumeID uuid.UUID) (*domain.ATSCheck, error) {
	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, domain.NotFound("resume", resumeID)
	}
	if resume.Status != domain.ResumeStatusReady {
		return nil, domain.ValidationError("Resume is not ready for ATS scoring", map[string]any{
			"resume_id": resumeID,
			"status":    resume.Status,
		})
	}

	cacheKey := hashCacheKey(resumeID.String() + ":resume")

	existing, err := s.atsRepo.GetByCacheKey(ctx, cacheKey)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("cache key lookup: %w", err))
	}
	if existing != nil {
		slog.Info("returning cached ATS resume check", "check_id", existing.ID)
		return existing, nil
	}

	balance, err := s.creditRepo.GetBalance(ctx, userID, domain.CreditATSCheck)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("check credit balance: %w", err))
	}
	if balance < 1 {
		return nil, domain.InsufficientCredits(domain.CreditATSCheck)
	}

	checkID := uuid.Must(uuid.NewV7())
	check := &domain.ATSCheck{
		ID:        checkID,
		UserID:    userID,
		ResumeID:  &resumeID,
		CheckType: domain.ATSCheckResume,
		Result:    json.RawMessage("{}"),
		CacheKey:  cacheKey,
	}

	if err := s.atsRepo.Create(ctx, check); err != nil {
		return nil, err
	}

	if err := s.deductATSCredit(ctx, userID, checkID, "ATS check — resume"); err != nil {
		slog.Error("failed to deduct ATS credit after check creation",
			"user_id", userID, "check_id", checkID, "error", err)
	}

	if err := s.enqueueTask(taskATSResumeCheck, checkID); err != nil {
		slog.Error("failed to enqueue ATS resume check task",
			"check_id", checkID, "error", err)
	}

	return check, nil
}

// CheckResumeTempUpload creates a resume-only ATS check from an inline PDF upload.
// The PDF is stored at a temporary S3 key (not a user resume slot), scored by the
// worker, and deleted from S3 once the result is stored.
//
// Pipeline:
//  1. Validate PDF (≤5 MB)
//  2. Check ats_check credit balance
//  3. Extract text from PDF
//  4. Upload PDF to temp S3 key
//  5. Compute cache key from extracted text
//  6. Return existing check if cache hit (no credit charge)
//  7. Create ATSCheck record (result = {}, TempS3Key set)
//  8. Deduct credit
//  9. Enqueue ats:resume_check worker task
func (s *ATSService) CheckResumeTempUpload(
	ctx context.Context,
	userID uuid.UUID,
	fileData []byte,
	fileName string,
) (*domain.ATSCheck, error) {
	const maxSize = 5 * 1024 * 1024
	if len(fileData) == 0 {
		return nil, domain.ValidationError("File is empty", nil)
	}
	if len(fileData) > maxSize {
		return nil, domain.ValidationError("File exceeds maximum size of 5 MB", nil)
	}
	if !pdf.IsPDF(fileData) {
		return nil, domain.ValidationError("File must be a PDF", map[string]any{"file_name": fileName})
	}

	balance, err := s.creditRepo.GetBalance(ctx, userID, domain.CreditATSCheck)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("check credit balance: %w", err))
	}
	if balance < 1 {
		return nil, domain.InsufficientCredits(domain.CreditATSCheck)
	}

	// Extract text for cache key computation and AI scoring.
	extractedText, _ := pdf.ExtractText(fileData)

	cacheKey := hashCacheKey(userID.String() + ":temp:" + hashCacheKey(extractedText))

	existing, err := s.atsRepo.GetByCacheKey(ctx, cacheKey)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("cache key lookup: %w", err))
	}
	if existing != nil {
		slog.Info("returning cached ATS temp upload check", "check_id", existing.ID)
		return existing, nil
	}

	checkID := uuid.Must(uuid.NewV7())
	tempKey := fmt.Sprintf("ats-temp/%s/%s.pdf", userID, checkID)

	if err := s.fileStore.Upload(ctx, tempKey, fileData, "application/pdf"); err != nil {
		return nil, domain.InternalError(fmt.Errorf("upload temp PDF: %w", err))
	}

	check := &domain.ATSCheck{
		ID:        checkID,
		UserID:    userID,
		ResumeID:  nil, // no slot resume — temp path
		TempS3Key: &tempKey,
		CheckType: domain.ATSCheckResume,
		Result:    json.RawMessage("{}"),
		CacheKey:  cacheKey,
	}

	if err := s.atsRepo.Create(ctx, check); err != nil {
		_ = s.fileStore.Delete(ctx, tempKey)
		return nil, err
	}

	if err := s.deductATSCredit(ctx, userID, checkID, "ATS check — resume (upload)"); err != nil {
		slog.Error("failed to deduct ATS credit after temp upload check",
			"user_id", userID, "check_id", checkID, "error", err)
	}

	if err := s.enqueueTask(taskATSResumeCheck, checkID); err != nil {
		slog.Error("failed to enqueue ATS temp upload check task",
			"check_id", checkID, "error", err)
	}

	return check, nil
}

// GetCheck returns a single ATS check by ID, verifying ownership.
func (s *ATSService) GetCheck(ctx context.Context, userID, checkID uuid.UUID) (*domain.ATSCheck, error) {
	check, err := s.atsRepo.GetByID(ctx, checkID)
	if err != nil {
		return nil, err
	}
	if check.UserID != userID {
		return nil, domain.NotFound("ats_check", checkID)
	}
	return check, nil
}

// ListChecks returns all ATS checks for a user, newest first.
func (s *ATSService) ListChecks(ctx context.Context, userID uuid.UUID) ([]domain.ATSCheck, error) {
	return s.atsRepo.ListByUser(ctx, userID)
}

// --- Internal helpers ---

func (s *ATSService) deductATSCredit(ctx context.Context, userID, checkID uuid.UUID, reason string) error {
	return s.txr.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.creditRepo.Deduct(txCtx, userID, domain.CreditATSCheck, 1); err != nil {
			return err
		}

		balance, err := s.creditRepo.GetBalance(txCtx, userID, domain.CreditATSCheck)
		if err != nil {
			return err
		}

		return s.creditRepo.LogTransaction(txCtx, &domain.CreditTransaction{
			ID:           uuid.Must(uuid.NewV7()),
			UserID:       userID,
			CreditType:   domain.CreditATSCheck,
			Amount:       -1,
			BalanceAfter: balance,
			Reason:       reason,
			ReferenceID:  &checkID,
			CreatedAt:    time.Now(),
		})
	})
}

func (s *ATSService) enqueueTask(taskType string, checkID uuid.UUID) error {
	payload := []byte(checkID.String())
	task := asynq.NewTask(taskType, payload)
	_, err := s.asynq.Enqueue(task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
	)
	return err
}

// hashCacheKey returns a short hex SHA-256 digest of the input string.
func hashCacheKey(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}
