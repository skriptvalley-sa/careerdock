package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// NotificationRepo implements domain.NotificationRepository using pgx.
type NotificationRepo struct {
	pool *pgxpool.Pool
}

// NewNotificationRepo creates a new NotificationRepo.
func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

// Create inserts a new notification.
func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	q := getDBTX(ctx, r.pool)

	err := q.QueryRow(ctx, `
		INSERT INTO notifications (id, user_id, type, title, message, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		n.ID, n.UserID, n.Type, n.Title, n.Message, n.Data, n.CreatedAt,
	).Scan(&n.ID)

	if err != nil {
		return domain.InternalError(err)
	}
	return nil
}

// ListByUser returns the most recent notifications for a user.
func (r *NotificationRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Notification, error) {
	q := getDBTX(ctx, r.pool)

	if limit < 1 || limit > 100 {
		limit = 20
	}

	rows, err := q.Query(ctx, `
		SELECT id, user_id, type, title, message, data, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, domain.InternalError(err)
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title,
			&n.Message, &n.Data, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, domain.InternalError(err)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError(err)
	}

	return notifications, nil
}

// MarkRead marks a notification as read.
func (r *NotificationRepo) MarkRead(ctx context.Context, id uuid.UUID) error {
	q := getDBTX(ctx, r.pool)
	now := time.Now().UTC()

	var returnedID uuid.UUID
	err := q.QueryRow(ctx, `
		UPDATE notifications SET read_at = $2
		WHERE id = $1 AND read_at IS NULL
		RETURNING id`, id, now).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFound("notification", id)
		}
		return domain.InternalError(err)
	}
	return nil
}

// CountUnread returns the number of unread notifications for a user.
func (r *NotificationRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	q := getDBTX(ctx, r.pool)

	var count int
	err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&count)
	if err != nil {
		return 0, domain.InternalError(err)
	}
	return count, nil
}
