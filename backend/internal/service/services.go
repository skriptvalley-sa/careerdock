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
	Auth        *AuthService
	Company     *CompanyService
	List        *ListService
	User        *UserService
	FeatureFlag *FeatureFlagService
	Payment     *PaymentService
	Credit      *CreditService
	Resume      *ResumeService

	// Infrastructure references for health checks
	db    *pgxpool.Pool
	redis *redis.Client

	// Runtime metadata
	Version                string
	IsProduction           bool
	RazorpayKeyID          string
	VerifyWebhookSignature func(payload []byte, signature string) bool
}

// NewServices creates a Services container.
//
//nolint:revive // parameter list is necessarily long — this is the DI constructor
func NewServices(
	auth *AuthService,
	company *CompanyService,
	list *ListService,
	user *UserService,
	featureFlag *FeatureFlagService,
	payment *PaymentService,
	credit *CreditService,
	resume *ResumeService,
	db *pgxpool.Pool,
	redisClient *redis.Client,
	version string,
	isProduction bool,
	razorpayKeyID string,
	verifyWebhookSig func(payload []byte, signature string) bool,
) *Services {
	return &Services{
		Auth:                   auth,
		Company:                company,
		List:                   list,
		User:                   user,
		FeatureFlag:            featureFlag,
		Payment:                payment,
		Credit:                 credit,
		Resume:                 resume,
		db:                     db,
		redis:                  redisClient,
		Version:                version,
		IsProduction:           isProduction,
		RazorpayKeyID:          razorpayKeyID,
		VerifyWebhookSignature: verifyWebhookSig,
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
