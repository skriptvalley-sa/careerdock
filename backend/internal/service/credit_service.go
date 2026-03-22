package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// CreditService handles credit balance queries and deductions.
type CreditService struct {
	credits domain.CreditRepository
	tx      domain.Transactor
}

// NewCreditService creates a new CreditService.
func NewCreditService(credits domain.CreditRepository, tx domain.Transactor) *CreditService {
	return &CreditService{credits: credits, tx: tx}
}

// GetAllBalances retrieves all credit balances for a user.
func (s *CreditService) GetAllBalances(ctx context.Context, userID uuid.UUID) (map[domain.CreditType]int, error) {
	return s.credits.GetAllBalances(ctx, userID)
}

// DeductCreditInput holds input for deducting a credit.
type DeductCreditInput struct {
	UserID      uuid.UUID
	CreditType  domain.CreditType
	Amount      int
	Reason      string
	ReferenceID *uuid.UUID
}

// DeductCredit atomically deducts credits and logs the transaction.
func (s *CreditService) DeductCredit(ctx context.Context, input DeductCreditInput) error {
	if input.Amount <= 0 {
		return domain.ValidationError("amount must be positive", map[string]any{
			"field": "amount",
		})
	}

	return s.tx.WithTx(ctx, func(txCtx context.Context) error {
		// Deduct (fails with INSUFFICIENT_CREDITS if balance too low)
		if err := s.credits.Deduct(txCtx, input.UserID, input.CreditType, input.Amount); err != nil {
			return err
		}

		// Get new balance for audit trail
		newBalance, err := s.credits.GetBalance(txCtx, input.UserID, input.CreditType)
		if err != nil {
			return err
		}

		txn := &domain.CreditTransaction{
			ID:           uuid.Must(uuid.NewV7()),
			UserID:       input.UserID,
			CreditType:   input.CreditType,
			Amount:       -input.Amount,
			BalanceAfter: newBalance,
			Reason:       input.Reason,
			ReferenceID:  input.ReferenceID,
			CreatedAt:    time.Now().UTC(),
		}
		return s.credits.LogTransaction(txCtx, txn)
	})
}

// ListTransactions retrieves credit transactions for a user.
func (s *CreditService) ListTransactions(ctx context.Context, userID uuid.UUID, creditType *domain.CreditType, limit int) ([]domain.CreditTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.credits.ListTransactionsByUser(ctx, userID, creditType, limit)
}

// HasCredits checks if a user has at least one credit of the given type.
func (s *CreditService) HasCredits(ctx context.Context, userID uuid.UUID, creditType domain.CreditType, required int) (bool, error) {
	balance, err := s.credits.GetBalance(ctx, userID, creditType)
	if err != nil {
		return false, err
	}
	return balance >= required, nil
}

// ValidateCreditType checks if the given string is a valid credit type.
func ValidateCreditType(ct string) (domain.CreditType, error) {
	switch domain.CreditType(ct) {
	case domain.CreditResumeUpload, domain.CreditATSCheck,
		domain.CreditCuratedList, domain.CreditCVGeneration:
		return domain.CreditType(ct), nil
	default:
		return "", domain.ValidationError(fmt.Sprintf("invalid credit_type: %s", ct), map[string]any{
			"field": "credit_type",
		})
	}
}
