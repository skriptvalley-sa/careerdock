package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/pdf"
)

const (
	maxResumeSizeMB    = 5
	maxResumeSizeBytes = maxResumeSizeMB * 1024 * 1024
	maxResumeSlots     = 3
)

// ResumeService handles resume upload, listing, and lifecycle management.
type ResumeService struct {
	resumeRepo domain.ResumeRepository
	userRepo   domain.UserRepository
	creditRepo domain.CreditRepository
	fileStore  domain.FileStore
	txr        domain.Transactor
	asynq      *asynq.Client
}

// NewResumeService creates a new ResumeService.
func NewResumeService(
	resumeRepo domain.ResumeRepository,
	userRepo domain.UserRepository,
	creditRepo domain.CreditRepository,
	fileStore domain.FileStore,
	txr domain.Transactor,
	asynqClient *asynq.Client,
) *ResumeService {
	return &ResumeService{
		resumeRepo: resumeRepo,
		userRepo:   userRepo,
		creditRepo: creditRepo,
		fileStore:  fileStore,
		txr:        txr,
		asynq:      asynqClient,
	}
}

// UploadResume validates, stores, and enqueues processing for a resume.
//
// Pipeline:
//  1. Validate file (PDF, ≤5 MB, valid slot)
//  2. Check resume_upload credits
//  3. Archive existing resume in slot (if occupied)
//  4. Upload to S3
//  5. Extract text from PDF
//  6. Create resume record (status = "parsing")
//  7. Deduct 1 resume_upload credit
//  8. Queue resume:parse_and_score worker task
func (s *ResumeService) UploadResume(
	ctx context.Context,
	userID uuid.UUID,
	fileName string,
	slotNumber int,
	fileData []byte,
) (*domain.Resume, error) {
	// 1. Validate file
	if err := s.validateUpload(fileName, slotNumber, fileData); err != nil {
		return nil, err
	}

	// 2. Check credits
	balance, err := s.creditRepo.GetBalance(ctx, userID, domain.CreditResumeUpload)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("check credit balance: %w", err))
	}
	if balance < 1 {
		return nil, domain.InsufficientCredits(domain.CreditResumeUpload)
	}

	// 3. Archive existing resume in slot (if any)
	existing, err := s.resumeRepo.GetByUserAndSlot(ctx, userID, slotNumber)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("check existing slot: %w", err))
	}
	if existing != nil {
		if archErr := s.resumeRepo.Archive(ctx, existing.ID); archErr != nil {
			return nil, domain.InternalError(fmt.Errorf("archive existing resume: %w", archErr))
		}
		if existing.IsDefault {
			if udErr := s.clearUserDefaultResume(ctx, userID); udErr != nil {
				slog.Warn("failed to clear default resume", "user_id", userID, "error", udErr)
			}
		}
	}

	// 4. Upload to S3
	resumeID := uuid.Must(uuid.NewV7())
	s3Key := fmt.Sprintf("%s/%s.pdf", userID, resumeID)
	if err := s.fileStore.Upload(ctx, s3Key, fileData, "application/pdf"); err != nil {
		return nil, domain.InternalError(fmt.Errorf("upload to S3: %w", err))
	}

	// 5. Extract text from PDF
	extractedText, extractErr := pdf.ExtractText(fileData)
	if extractErr != nil {
		slog.Warn("PDF text extraction failed — will proceed with AI-only parsing",
			"resume_id", resumeID,
			"error", extractErr,
		)
		extractedText = ""
	}

	// 6. Create resume record
	resume := &domain.Resume{
		ID:            resumeID,
		UserID:        userID,
		SlotNumber:    slotNumber,
		FileName:      fileName,
		FileSizeBytes: len(fileData),
		S3Key:         s3Key,
		ExtractedText: strPtr(extractedText),
		Status:        domain.ResumeStatusParsing,
		IsDefault:     false,
		IsArchived:    false,
	}

	if err := s.resumeRepo.Create(ctx, resume); err != nil {
		_ = s.fileStore.Delete(ctx, s3Key) // clean up on failure
		return nil, err
	}

	// 7. Deduct credit
	if err := s.deductResumeCredit(ctx, userID, resumeID); err != nil {
		slog.Error("failed to deduct resume credit after upload",
			"user_id", userID, "resume_id", resumeID, "error", err)
	}

	// 8. Queue worker task
	if err := s.enqueueParseTask(resumeID); err != nil {
		slog.Error("failed to enqueue resume parse task",
			"resume_id", resumeID, "error", err)
	}

	return resume, nil
}

// ListResumes returns all active (non-archived) resumes for a user.
func (s *ResumeService) ListResumes(ctx context.Context, userID uuid.UUID) ([]domain.Resume, error) {
	return s.resumeRepo.ListByUser(ctx, userID)
}

// GetResume returns a single resume by ID, verifying ownership.
func (s *ResumeService) GetResume(ctx context.Context, userID, resumeID uuid.UUID) (*domain.Resume, error) {
	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, domain.NotFound("resume", resumeID)
	}
	return resume, nil
}

// SetDefault marks a resume as the user's default.
func (s *ResumeService) SetDefault(ctx context.Context, userID, resumeID uuid.UUID) error {
	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return err
	}
	if resume.UserID != userID {
		return domain.NotFound("resume", resumeID)
	}
	if resume.IsArchived {
		return domain.ValidationError("Cannot set archived resume as default", nil)
	}

	return s.txr.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.resumeRepo.ClearDefaultForUser(txCtx, userID); err != nil {
			return err
		}

		resume.IsDefault = true
		if err := s.resumeRepo.Update(txCtx, resume); err != nil {
			return err
		}

		user, err := s.userRepo.GetByID(txCtx, userID)
		if err != nil {
			return err
		}
		user.DefaultResumeID = &resumeID
		return s.userRepo.Update(txCtx, user)
	})
}

// ArchiveResume archives a resume, removing it from active slots.
func (s *ResumeService) ArchiveResume(ctx context.Context, userID, resumeID uuid.UUID) error {
	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return err
	}
	if resume.UserID != userID {
		return domain.NotFound("resume", resumeID)
	}

	if err := s.resumeRepo.Archive(ctx, resumeID); err != nil {
		return err
	}

	if resume.IsDefault {
		if err := s.clearUserDefaultResume(ctx, userID); err != nil {
			slog.Warn("failed to clear default resume on archive", "user_id", userID, "error", err)
		}
	}

	return nil
}

// GetDownloadURL generates a pre-signed S3 URL for downloading a resume.
func (s *ResumeService) GetDownloadURL(ctx context.Context, userID, resumeID uuid.UUID) (string, error) {
	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return "", err
	}
	if resume.UserID != userID {
		return "", domain.NotFound("resume", resumeID)
	}

	url, err := s.fileStore.GenerateSignedURL(ctx, resume.S3Key, 15*time.Minute)
	if err != nil {
		return "", domain.InternalError(fmt.Errorf("generate download URL: %w", err))
	}

	return url, nil
}

// --- Internal helpers ---

func (s *ResumeService) validateUpload(fileName string, slotNumber int, fileData []byte) *domain.AppError {
	if slotNumber < 1 || slotNumber > maxResumeSlots {
		return domain.ValidationError(
			fmt.Sprintf("Invalid slot number. Must be 1-%d", maxResumeSlots),
			map[string]any{"slot_number": slotNumber},
		)
	}

	if len(fileData) == 0 {
		return domain.ValidationError("File is empty", nil)
	}

	if len(fileData) > maxResumeSizeBytes {
		return domain.ValidationError(
			fmt.Sprintf("File exceeds maximum size of %d MB", maxResumeSizeMB),
			map[string]any{"file_size_bytes": len(fileData), "max_bytes": maxResumeSizeBytes},
		)
	}

	if !pdf.IsPDF(fileData) {
		return domain.ValidationError("File must be a PDF", map[string]any{"file_name": fileName})
	}

	return nil
}

func (s *ResumeService) deductResumeCredit(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) error {
	return s.txr.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.creditRepo.Deduct(txCtx, userID, domain.CreditResumeUpload, 1); err != nil {
			return err
		}

		balance, err := s.creditRepo.GetBalance(txCtx, userID, domain.CreditResumeUpload)
		if err != nil {
			return err
		}

		return s.creditRepo.LogTransaction(txCtx, &domain.CreditTransaction{
			ID:           uuid.Must(uuid.NewV7()),
			UserID:       userID,
			CreditType:   domain.CreditResumeUpload,
			Amount:       -1,
			BalanceAfter: balance,
			Reason:       "Resume upload",
			ReferenceID:  &resumeID,
			CreatedAt:    time.Now(),
		})
	})
}

func (s *ResumeService) enqueueParseTask(resumeID uuid.UUID) error {
	payload := []byte(resumeID.String())
	task := asynq.NewTask("resume:parse_and_score", payload)
	_, err := s.asynq.Enqueue(task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
	)
	return err
}

func (s *ResumeService) clearUserDefaultResume(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.DefaultResumeID = nil
	return s.userRepo.Update(ctx, user)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
