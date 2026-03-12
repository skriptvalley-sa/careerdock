package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Services is the top-level dependency container that holds all service
// instances. It is constructed once in cmd/api/main.go and passed to the
// handler layer via MountRoutes.
type Services struct {
	Auth *AuthService

	// Infrastructure references for health checks
	db    *pgxpool.Pool
	redis *redis.Client

	// Runtime metadata
	Version      string
	IsProduction bool
}

// NewServices creates a Services container.
func NewServices(
	auth *AuthService,
	db *pgxpool.Pool,
	redisClient *redis.Client,
	version string,
	isProduction bool,
) *Services {
	return &Services{
		Auth:         auth,
		db:           db,
		redis:        redisClient,
		Version:      version,
		IsProduction: isProduction,
	}
}

// PingDB checks database connectivity.
func (s *Services) PingDB(ctx context.Context) error {
	return s.db.Ping(ctx)
}

// PingRedis checks Redis connectivity.
func (s *Services) PingRedis(ctx context.Context) error {
	return s.redis.Ping(ctx).Err()
}
