// Package main is the entry point for the CareerDock API server.
//
// Boot sequence:
//  1. Load config (Viper)
//  2. Set up structured logger (slog/JSON)
//  3. Connect infrastructure (Postgres, Redis)
//  4. Build repository layer
//  5. Build service layer
//  6. Build handler layer + mount routes
//  7. Start HTTP server with graceful shutdown
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/hibiken/asynq"

	"github.com/skriptvalley/careerdock/internal/ai"
	"github.com/skriptvalley/careerdock/internal/config"
	"github.com/skriptvalley/careerdock/internal/handler"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/payment"
	"github.com/skriptvalley/careerdock/internal/repository"
	"github.com/skriptvalley/careerdock/internal/service"
	"github.com/skriptvalley/careerdock/internal/storage"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	// 1. Load config
	cfg := config.MustLoad()

	// 2. Set up structured logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel(),
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting CareerDock API",
		"version", version,
		"environment", cfg.Environment,
		"port", cfg.Port,
	)

	// 3. Connect infrastructure
	ctx := context.Background()

	db, err := connectDB(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return
	}
	defer db.Close()
	logger.Info("connected to database")

	redisClient, err := connectRedis(ctx, cfg.RedisURL)
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		return
	}
	defer func() { _ = redisClient.Close() }()
	logger.Info("connected to Redis")

	// 4. Build repository layer
	txr := repository.NewTransactor(db)
	userRepo := repository.NewUserRepo(db)
	tokenRepo := repository.NewTokenRepo(db)
	companyRepo := repository.NewCompanyRepo(db)
	listRepo := repository.NewListRepo(db)
	featureFlagRepo := repository.NewFeatureFlagRepo(db)
	paymentRepo := repository.NewPaymentRepo(db)
	creditRepo := repository.NewCreditRepo(db)
	resumeRepo := repository.NewResumeRepo(db)
	atsCheckRepo := repository.NewATSCheckRepo(db)
	curatedListRepo := repository.NewCuratedListRepo(db)
	auditLogRepo := repository.NewAuditLogRepo(db)

	// S3 storage for resume files
	resumeStore, err := storage.NewS3Store(ctx, cfg.S3, cfg.S3.ResumeBucket)
	if err != nil {
		logger.Error("failed to create S3 resume store", "error", err)
		return
	}
	if cfg.IsDevelopment() {
		if err := resumeStore.EnsureBucket(ctx); err != nil {
			logger.Warn("failed to ensure resume bucket (MinIO may not be running)", "error", err)
		}
	}

	// S3 storage for company logos
	logoStore, err := storage.NewS3Store(ctx, cfg.S3, cfg.S3.LogoBucket)
	if err != nil {
		logger.Error("failed to create S3 logo store", "error", err)
		return
	}
	if cfg.IsDevelopment() {
		if err := logoStore.EnsureBucket(ctx); err != nil {
			logger.Warn("failed to ensure logo bucket (MinIO may not be running)", "error", err)
		}
	}

	// Asynq client for queuing background tasks
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisURL})
	defer func() { _ = asynqClient.Close() }()

	// 5. Build service layer
	authSvc := service.NewAuthService(userRepo, tokenRepo, txr, redisClient, cfg.JWTSecret)
	companySvc := service.NewCompanyService(companyRepo)
	listSvc := service.NewListService(listRepo, userRepo, txr)
	userSvc := service.NewUserService(userRepo, txr)
	featureFlagSvc := service.NewFeatureFlagService(featureFlagRepo)
	razorpayGateway := payment.NewRazorpayGateway(cfg.RazorpayKeyID, cfg.RazorpayKeySecret, cfg.RazorpayWebhookSecret)
	paymentSvc := service.NewPaymentService(paymentRepo, creditRepo, userRepo, razorpayGateway, txr)
	creditSvc := service.NewCreditService(creditRepo, txr)
	resumeSvc := service.NewResumeService(resumeRepo, userRepo, creditRepo, resumeStore, txr, asynqClient)
	atsSvc := service.NewATSService(atsCheckRepo, resumeRepo, companyRepo, creditRepo, resumeStore, txr, asynqClient)
	curatedListSvc := service.NewCuratedListService(curatedListRepo, resumeRepo, creditRepo, txr, asynqClient)
	adminSvc := service.NewAdminService(companyRepo, userRepo, creditRepo, paymentRepo, auditLogRepo, logoStore, txr, redisClient)
	notificationRepo := repository.NewNotificationRepo(db)
	notificationSvc := service.NewNotificationService(notificationRepo)

	// AI provider for moderator company generation (fallback: Claude → OpenAI)
	claudeProvider := ai.NewClaudeProvider(cfg.ClaudeAPIKey, "", 0)
	openaiProvider := ai.NewOpenAIProvider(cfg.OpenAIAPIKey, "", 0)
	aiProvider := ai.NewFallbackProvider(claudeProvider, openaiProvider)

	editLockRepo := repository.NewCompanyEditLockRepo(db)
	moderatorSvc := service.NewModeratorService(companyRepo, editLockRepo, aiProvider)

	applicationRepo := repository.NewApplicationRepo(db)
	applicationSvc := service.NewApplicationService(applicationRepo, txr)

	svc := service.NewServices(authSvc, companySvc, listSvc, userSvc, featureFlagSvc, paymentSvc, creditSvc, resumeSvc, atsSvc, curatedListSvc, adminSvc, notificationSvc, moderatorSvc, applicationSvc, db, redisClient, version, cfg.IsProduction(), cfg.RazorpayKeyID, razorpayGateway.VerifyWebhookSignature)

	// 6. Build handler layer + mount routes
	r := chi.NewRouter()

	// Global middleware (order matters: outermost first)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// Rate limiting: configurable via RATE_LIMIT_IP_PER_MIN / RATE_LIMIT_USER_PER_MIN env vars
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.RateLimitIPPerMin, cfg.RateLimitUserPerMin, time.Minute)
	r.Use(rateLimiter.Middleware)

	auth := middleware.NewAuth(authSvc)
	handler.MountRoutes(r, svc, auth)

	// 7. Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("server stopped gracefully")
}

// connectDB creates a pgx connection pool.
func connectDB(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// connectRedis creates a Redis client and verifies connectivity.
func connectRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL("redis://" + redisURL)
	if err != nil {
		// Fallback: treat as host:port
		opts = &redis.Options{Addr: redisURL}
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return client, nil
}
