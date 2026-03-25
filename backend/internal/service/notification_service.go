package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// NotificationService handles notification operations.
type NotificationService struct {
	notifications domain.NotificationRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(notifications domain.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

// List returns the most recent notifications for a user.
func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Notification, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.notifications.ListByUser(ctx, userID, limit)
}

// MarkRead marks a notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, id uuid.UUID) error {
	return s.notifications.MarkRead(ctx, id)
}

// CountUnread returns the number of unread notifications for a user.
func (s *NotificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.notifications.CountUnread(ctx, userID)
}
