package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/skriptvalley/careerdock/internal/domain"
)

const taskCurateCompanyList = "ai:curate_company_list"

// CuratedListService handles curated company list generation requests.
type CuratedListService struct {
	curatedListRepo domain.CuratedListRepository
	resumeRepo      domain.ResumeRepository
	creditRepo      domain.CreditRepository
	txr             domain.Transactor
	asynq           *asynq.Client
}

// NewCuratedListService creates a new CuratedListService.
func NewCuratedListService(
	curatedListRepo domain.CuratedListRepository,
	resumeRepo domain.ResumeRepository,
	creditRepo domain.CreditRepository,
	txr domain.Transactor,
	asynqClient *asynq.Client,
) *CuratedListService {
	return &CuratedListService{
		curatedListRepo: curatedListRepo,
		resumeRepo:      resumeRepo,
		creditRepo:      creditRepo,
		txr:             txr,
		asynq:           asynqClient,
	}
}

// GenerateList creates a curated company list for a user's resume, deducts a credit,
// and enqueues the AI worker task.
//
// Pipeline:
//  1. Verify resume exists, belongs to user, and is ready
//  2. Compute preferences_hash (sha256 of resumeID) for deduplication
//  3. Return existing list if hash matches (no credit charge)
//  4. Check curated_list credit balance
//  5. Create CuratedList record (result = {})
//  6. Deduct credit
//  7. Enqueue ai:curate_company_list task
func (s *CuratedListService) GenerateList(ctx context.Context, userID, resumeID uuid.UUID) (*domain.CuratedList, error) {
	resume, err := s.resumeRepo.GetByID(ctx, resumeID)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, domain.NotFound("resume", resumeID)
	}
	if resume.Status != domain.ResumeStatusReady {
		return nil, domain.ValidationError("Resume is not ready for curation", map[string]any{
			"resume_id": resumeID,
			"status":    resume.Status,
		})
	}
	if len(resume.ParsedData) == 0 {
		return nil, domain.ValidationError("Resume has not been parsed yet", nil)
	}

	preferencesHash := hashCacheKey(resumeID.String())

	existing, err := s.curatedListRepo.GetByPreferencesHash(ctx, preferencesHash)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("preferences hash lookup: %w", err))
	}
	if existing != nil {
		slog.Info("returning cached curated list", "list_id", existing.ID)
		return existing, nil
	}

	balance, err := s.creditRepo.GetBalance(ctx, userID, domain.CreditCuratedList)
	if err != nil {
		return nil, domain.InternalError(fmt.Errorf("check credit balance: %w", err))
	}
	if balance < 1 {
		return nil, domain.InsufficientCredits(domain.CreditCuratedList)
	}

	listID := uuid.Must(uuid.NewV7())
	list := &domain.CuratedList{
		ID:              listID,
		UserID:          userID,
		ResumeID:        resumeID,
		PreferencesHash: preferencesHash,
		Result:          json.RawMessage("{}"),
	}

	if err := s.curatedListRepo.Create(ctx, list); err != nil {
		return nil, err
	}

	if err := s.deductCuratedListCredit(ctx, userID, listID); err != nil {
		slog.Error("failed to deduct curated list credit after creation",
			"user_id", userID, "list_id", listID, "error", err)
	}

	if err := s.enqueueTask(taskCurateCompanyList, listID); err != nil {
		slog.Error("failed to enqueue curate list task",
			"list_id", listID, "error", err)
	}

	return list, nil
}

// GetList returns a single curated list by ID, verifying ownership.
func (s *CuratedListService) GetList(ctx context.Context, userID, listID uuid.UUID) (*domain.CuratedList, error) {
	list, err := s.curatedListRepo.GetByID(ctx, listID)
	if err != nil {
		return nil, err
	}
	if list.UserID != userID {
		return nil, domain.NotFound("curated_list", listID)
	}
	return list, nil
}

// ListByUser returns all curated lists for a user, newest first.
func (s *CuratedListService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.CuratedList, error) {
	return s.curatedListRepo.ListByUser(ctx, userID)
}

// Rename updates the name of a curated list, verifying ownership.
func (s *CuratedListService) Rename(ctx context.Context, userID, listID uuid.UUID, name string) error {
	list, err := s.curatedListRepo.GetByID(ctx, listID)
	if err != nil {
		return err
	}
	if list.UserID != userID {
		return domain.NotFound("curated_list", listID)
	}
	return s.curatedListRepo.Rename(ctx, listID, name)
}

// Delete removes a curated list, verifying ownership.
func (s *CuratedListService) Delete(ctx context.Context, userID, listID uuid.UUID) error {
	list, err := s.curatedListRepo.GetByID(ctx, listID)
	if err != nil {
		return err
	}
	if list.UserID != userID {
		return domain.NotFound("curated_list", listID)
	}
	return s.curatedListRepo.Delete(ctx, listID)
}

// --- Internal helpers ---

func (s *CuratedListService) deductCuratedListCredit(ctx context.Context, userID, listID uuid.UUID) error {
	return s.txr.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.creditRepo.Deduct(txCtx, userID, domain.CreditCuratedList, 1); err != nil {
			return err
		}

		balance, err := s.creditRepo.GetBalance(txCtx, userID, domain.CreditCuratedList)
		if err != nil {
			return err
		}

		return s.creditRepo.LogTransaction(txCtx, &domain.CreditTransaction{
			ID:           uuid.Must(uuid.NewV7()),
			UserID:       userID,
			CreditType:   domain.CreditCuratedList,
			Amount:       -1,
			BalanceAfter: balance,
			Reason:       "AI curated company list",
			ReferenceID:  &listID,
			CreatedAt:    time.Now(),
		})
	})
}

func (s *CuratedListService) enqueueTask(taskType string, listID uuid.UUID) error {
	payload := []byte(listID.String())
	task := asynq.NewTask(taskType, payload)
	_, err := s.asynq.Enqueue(task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Minute), // longer timeout — large company list
	)
	return err
}
