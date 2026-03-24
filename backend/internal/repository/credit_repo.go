package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// CreditRepo implements domain.CreditRepository using pgx.
type CreditRepo struct {
	pool *pgxpool.Pool
}

// NewCreditRepo creates a new CreditRepo.
func NewCreditRepo(pool *pgxpool.Pool) *CreditRepo {
	return &CreditRepo{pool: pool}
}

// GetBalance retrieves the current balance for a credit type.
// Returns 0 if no row exists (user has never received this credit type).
func (r *CreditRepo) GetBalance(ctx context.Context, userID uuid.UUID, creditType domain.CreditType) (int, error) {
	q := getDBTX(ctx, r.pool)

	var balance int
	err := q.QueryRow(ctx, `
		SELECT balance FROM user_credits
		WHERE user_id = $1 AND credit_type = $2`,
		userID, string(creditType),
	).Scan(&balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, domain.InternalError(err)
	}
	return balance, nil
}

// GetAllBalances retrieves all credit balances for a user.
func (r *CreditRepo) GetAllBalances(ctx context.Context, userID uuid.UUID) (map[domain.CreditType]int, error) {
	q := getDBTX(ctx, r.pool)

	rows, err := q.Query(ctx, `
		SELECT credit_type, balance FROM user_credits
		WHERE user_id = $1`, userID)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	balances := make(map[domain.CreditType]int)
	for rows.Next() {
		var ct domain.CreditType
		var bal int
		if err := rows.Scan(&ct, &bal); err != nil {
			return nil, domain.InternalError(err)
		}
		balances[ct] = bal
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return balances, nil
}

// Allocate adds credits to a user's balance using an upsert.
// Safe to call concurrently — ON CONFLICT atomically increments the balance.
func (r *CreditRepo) Allocate(ctx context.Context, userID uuid.UUID, creditType domain.CreditType, amount int) error {
	q := getDBTX(ctx, r.pool)

	_, err := q.Exec(ctx, `
		INSERT INTO user_credits (id, user_id, credit_type, balance, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, credit_type) DO UPDATE SET
			balance = user_credits.balance + EXCLUDED.balance,
			updated_at = NOW()`,
		uuid.Must(uuid.NewV7()), userID, string(creditType), amount,
	)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// Deduct subtracts credits from a user's balance.
// Returns INSUFFICIENT_CREDITS if the balance is too low.
// The CHECK (balance >= 0) constraint also guards against races.
func (r *CreditRepo) Deduct(ctx context.Context, userID uuid.UUID, creditType domain.CreditType, amount int) error {
	q := getDBTX(ctx, r.pool)

	var newBalance int
	err := q.QueryRow(ctx, `
		UPDATE user_credits
		SET balance = balance - $3, updated_at = NOW()
		WHERE user_id = $1 AND credit_type = $2 AND balance >= $3
		RETURNING balance`,
		userID, string(creditType), amount,
	).Scan(&newBalance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InsufficientCredits(creditType)
		}
		return domain.InternalError(err)
	}
	return nil
}

// LogTransaction inserts an immutable credit transaction audit record.
func (r *CreditRepo) LogTransaction(ctx context.Context, txn *domain.CreditTransaction) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO credit_transactions (
			id, user_id, credit_type, amount, balance_after,
			reason, reference_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		txn.ID, txn.UserID, string(txn.CreditType), txn.Amount, txn.BalanceAfter,
		txn.Reason, txn.ReferenceID, txn.CreatedAt,
	).Scan(&txn.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// ListTransactionsByUser retrieves credit transactions for a user, newest first.
func (r *CreditRepo) ListTransactionsByUser(ctx context.Context, userID uuid.UUID, creditType *domain.CreditType, limit int) ([]domain.CreditTransaction, error) {
	q := getDBTX(ctx, r.pool)

	query := `
		SELECT id, user_id, credit_type, amount, balance_after,
		       reason, reference_id, created_at
		FROM credit_transactions
		WHERE user_id = $1`
	args := []any{userID}

	if creditType != nil {
		query += ` AND credit_type = $2`
		args = append(args, string(*creditType))
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		if creditType != nil {
			query += ` LIMIT $3`
		} else {
			query += ` LIMIT $2`
		}
		args = append(args, limit)
	}

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var txns []domain.CreditTransaction
	for rows.Next() {
		var t domain.CreditTransaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.CreditType, &t.Amount, &t.BalanceAfter,
			&t.Reason, &t.ReferenceID, &t.CreatedAt,
		); err != nil {
			return nil, domain.InternalError(err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return txns, nil
}

// ListAllTransactions returns credit transactions matching the filter with total count (admin use).
func (r *CreditRepo) ListAllTransactions(ctx context.Context, filter domain.CreditTransactionFilter) ([]domain.CreditTransaction, int, error) {
	q := getDBTX(ctx, r.pool)

	clauses := []string{}
	args := []any{}
	argIdx := 1

	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", argIdx))
		argIdx++
	}
	if filter.CreditType != nil {
		args = append(args, string(*filter.CreditType))
		clauses = append(clauses, fmt.Sprintf("credit_type = $%d", argIdx))
		argIdx++
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	var total int
	if err := q.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM credit_transactions %s", where), args...).Scan(&total); err != nil {
		return nil, 0, domain.InternalError(err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit)
	limitParam := fmt.Sprintf("$%d", argIdx)
	argIdx++
	args = append(args, offset)
	offsetParam := fmt.Sprintf("$%d", argIdx)

	query := fmt.Sprintf(`
		SELECT id, user_id, credit_type, amount, balance_after,
		       reason, reference_id, created_at
		FROM credit_transactions
		%s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s`, where, limitParam, offsetParam)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, domain.InternalError(err)
	}
	defer rows.Close()

	var txns []domain.CreditTransaction
	for rows.Next() {
		var t domain.CreditTransaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.CreditType, &t.Amount, &t.BalanceAfter,
			&t.Reason, &t.ReferenceID, &t.CreatedAt,
		); err != nil {
			return nil, 0, domain.InternalError(err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, domain.InternalError(err)
	}
	return txns, total, nil
}
