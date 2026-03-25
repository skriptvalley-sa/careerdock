package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// AuditLogRepo implements domain.AuditLogRepository using pgx.
type AuditLogRepo struct {
	pool *pgxpool.Pool
}

// NewAuditLogRepo creates a new AuditLogRepo.
func NewAuditLogRepo(pool *pgxpool.Pool) *AuditLogRepo {
	return &AuditLogRepo{pool: pool}
}

// Create inserts an audit log entry.
func (r *AuditLogRepo) Create(ctx context.Context, entry *domain.AuditLogEntry) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO admin_audit_log (
			id, admin_id, action, entity_type, entity_id,
			details, ip_address, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		entry.ID, entry.AdminID, entry.Action, entry.EntityType, entry.EntityID,
		entry.Details, entry.IPAddress, entry.CreatedAt,
	).Scan(&entry.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// List returns audit log entries matching the filter, newest first.
func (r *AuditLogRepo) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLogEntry, error) {
	q := getDBTX(ctx, r.pool)

	clauses := []string{}
	args := []any{}
	argIdx := 1

	if filter.AdminID != nil {
		args = append(args, *filter.AdminID)
		clauses = append(clauses, fmt.Sprintf("admin_id = $%d", argIdx))
		argIdx++
	}
	if filter.EntityType != nil {
		args = append(args, *filter.EntityType)
		clauses = append(clauses, fmt.Sprintf("entity_type = $%d", argIdx))
		argIdx++
	}
	if filter.EntityID != nil {
		args = append(args, *filter.EntityID)
		clauses = append(clauses, fmt.Sprintf("entity_id = $%d", argIdx))
		argIdx++
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
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
		SELECT id, admin_id, action, entity_type, entity_id,
		       details, ip_address, created_at
		FROM admin_audit_log
		%s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s`, where, limitParam, offsetParam)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var entries []domain.AuditLogEntry
	for rows.Next() {
		var e domain.AuditLogEntry
		if err := rows.Scan(
			&e.ID, &e.AdminID, &e.Action, &e.EntityType, &e.EntityID,
			&e.Details, &e.IPAddress, &e.CreatedAt,
		); err != nil {
			return nil, domain.InternalError(err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}
	return entries, nil
}
