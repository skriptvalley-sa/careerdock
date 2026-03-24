package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/skriptvalley/careerdock/internal/domain"
)

// UserCleanupHandler hard-deletes users whose soft-delete grace period has
// elapsed (30 days). It is invoked by the Asynq scheduler once per day.
type UserCleanupHandler struct {
	userRepo domain.UserRepository
}

// NewUserCleanupHandler creates a new UserCleanupHandler.
func NewUserCleanupHandler(userRepo domain.UserRepository) *UserCleanupHandler {
	return &UserCleanupHandler{userRepo: userRepo}
}

// Handle performs the hard-delete sweep.
func (h *UserCleanupHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -30)

	n, err := h.userRepo.HardDeleteExpired(ctx, cutoff)
	if err != nil {
		slog.Error("user cleanup: hard-delete failed", "error", err)
		return err
	}

	if n > 0 {
		slog.Info("user cleanup: hard-deleted expired accounts", "count", n)
	}
	return nil
}
