// Package repository implements domain.Transactor and all repository
// interfaces using pgx/v5 against PostgreSQL.
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ctxKey is an unexported type for context keys in this package.
type ctxKey string

const txKey ctxKey = "pg_tx"

// PgTransactor implements domain.Transactor using pgxpool.
type PgTransactor struct {
	pool *pgxpool.Pool
}

// NewTransactor creates a new PgTransactor.
func NewTransactor(pool *pgxpool.Pool) *PgTransactor {
	return &PgTransactor{pool: pool}
}

// WithTx executes fn inside a database transaction. If fn returns an error
// the transaction is rolled back; otherwise it is committed.
// Nested calls to WithTx share the outer transaction (no savepoints).
func (t *PgTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// If already inside a transaction, just run fn.
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// DBTX is the minimal interface that both *pgxpool.Pool and pgx.Tx implement,
// used by repository methods.
type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// getDBTX extracts the transaction from context, or falls back to the pool.
func getDBTX(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}
