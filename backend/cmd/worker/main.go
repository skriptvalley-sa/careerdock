// Package main is the entry point for the CareerDock background worker.
//
// It uses hibiken/asynq backed by Redis to process async tasks such as
// email delivery, resume parsing, and ATS scoring. Task handlers are
// registered here and implemented in internal/worker/.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/skriptvalley/careerdock/internal/ai"
	"github.com/skriptvalley/careerdock/internal/config"
	"github.com/skriptvalley/careerdock/internal/email"
	"github.com/skriptvalley/careerdock/internal/repository"
	"github.com/skriptvalley/careerdock/internal/storage"
	"github.com/skriptvalley/careerdock/internal/worker"
)

// Task type constants — kept here as the canonical registry.
// Handlers are in internal/worker/task_*.go.
const (
	TaskSendEmail           = "email:send"
	TaskResumeParseAndScore = "resume:parse_and_score"
	TaskATSCompanyCheck     = "ats:company_check"
	TaskATSJobCheck         = "ats:job_check"
	TaskCurateCompanyList   = "ai:curate_company_list"
	TaskUserCleanup         = "admin:user_cleanup"
	TaskCompanyEnrich       = "admin:company_enrich"
	TaskCompanyRefresh      = "admin:company_refresh"
)

func main() {
	// 1. Load config
	cfg := config.MustLoad()

	// 2. Set up structured logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel(),
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting CareerDock Worker",
		"environment", cfg.Environment,
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

	// Build repositories and services needed by worker tasks
	resumeRepo := repository.NewResumeRepo(db)
	companyRepo := repository.NewCompanyRepo(db)
	atsCheckRepo := repository.NewATSCheckRepo(db)
	curatedListRepo := repository.NewCuratedListRepo(db)
	resumeStore, err := storage.NewS3Store(ctx, cfg.S3, cfg.S3.ResumeBucket)
	if err != nil {
		logger.Error("failed to create S3 resume store", "error", err)
		return
	}

	// Build AI provider (fallback: Claude → OpenAI)
	var aiProvider ai.LLMProvider
	claudeProvider := ai.NewClaudeProvider(cfg.ClaudeAPIKey, "", 0)
	openaiProvider := ai.NewOpenAIProvider(cfg.OpenAIAPIKey, "", 0)
	aiProvider = ai.NewFallbackProvider(claudeProvider, openaiProvider)

	// Asynq client for enqueueing sub-tasks (e.g., company refresh → enrich)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisURL})
	defer func() { _ = asynqClient.Close() }()

	// AI result cache
	aiCache := ai.NewResultCache(redisClient)

	// Email sender
	emailSender := email.NewResendSender(cfg.ResendAPIKey, cfg.FromEmail)

	// 4. Create Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisURL},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: newAsynqLogger(logger),
		},
	)

	// 5. Register task handlers
	mux := asynq.NewServeMux()

	// Sprint 3: Email send + Resume parse/score
	emailHandler := worker.NewEmailSendHandler(emailSender)
	mux.HandleFunc(TaskSendEmail, emailHandler.Handle)

	resumeParseHandler := worker.NewResumeParseHandler(resumeRepo, resumeStore, aiProvider, aiCache, redisClient)
	mux.HandleFunc(TaskResumeParseAndScore, resumeParseHandler.Handle)

	// Sprint 4: ATS company + job check workers
	atsCompanyHandler := worker.NewATSCompanyHandler(atsCheckRepo, resumeRepo, companyRepo, resumeStore, aiProvider, aiCache, redisClient)
	mux.HandleFunc(TaskATSCompanyCheck, atsCompanyHandler.Handle)

	atsJobHandler := worker.NewATSJobHandler(atsCheckRepo, resumeRepo, resumeStore, aiProvider, aiCache, redisClient)
	mux.HandleFunc(TaskATSJobCheck, atsJobHandler.Handle)

	// Sprint 4: Curated company list worker
	curateListHandler := worker.NewCurateListHandler(curatedListRepo, resumeRepo, companyRepo, aiProvider, aiCache, redisClient)
	mux.HandleFunc(TaskCurateCompanyList, curateListHandler.Handle)

	// Sprint 4: Periodic user hard-delete cleanup
	userRepo := repository.NewUserRepo(db)
	userCleanupHandler := worker.NewUserCleanupHandler(userRepo)
	mux.HandleFunc(TaskUserCleanup, userCleanupHandler.Handle)

	// Sprint 5: Company enrichment workers
	companyEnrichHandler := worker.NewCompanyEnrichHandler(companyRepo, aiProvider)
	mux.HandleFunc(TaskCompanyEnrich, companyEnrichHandler.Handle)

	companyRefreshHandler := worker.NewCompanyRefreshHandler(companyRepo, asynqClient)
	mux.HandleFunc(TaskCompanyRefresh, companyRefreshHandler.Handle)

	// 6. Set up Asynq scheduler for periodic tasks
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: cfg.RedisURL},
		&asynq.SchedulerOpts{
			Logger: newAsynqLogger(logger),
		},
	)

	// Run user hard-delete sweep daily at 02:00 UTC.
	if _, err := scheduler.Register("0 2 * * *", asynq.NewTask(TaskUserCleanup, nil),
		asynq.Queue("low"),
	); err != nil {
		logger.Error("failed to register scheduler task", "task", TaskUserCleanup, "error", err)
		return
	}

	// Run company refresh weekly on Sundays at 03:00 UTC.
	if _, err := scheduler.Register("0 3 * * 0", asynq.NewTask(TaskCompanyRefresh, nil),
		asynq.Queue("low"),
	); err != nil {
		logger.Error("failed to register scheduler task", "task", TaskCompanyRefresh, "error", err)
		return
	}

	// 7. Start server + scheduler with graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		logger.Info("Asynq worker starting", "concurrency", 5)
		errCh <- srv.Run(mux)
	}()

	schedErrCh := make(chan error, 1)
	go func() {
		logger.Info("Asynq scheduler starting")
		schedErrCh <- scheduler.Run()
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		if err != nil {
			logger.Error("worker error", "error", err)
		}
	case err := <-schedErrCh:
		if err != nil {
			logger.Error("scheduler error", "error", err)
		}
	}

	scheduler.Shutdown()
	srv.Shutdown()
	logger.Info("worker stopped gracefully")
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
		opts = &redis.Options{Addr: redisURL}
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return client, nil
}

// asynqLogger adapts slog to the asynq.Logger interface.
type asynqLogger struct {
	logger *slog.Logger
}

func newAsynqLogger(l *slog.Logger) *asynqLogger {
	return &asynqLogger{logger: l.With("component", "asynq")}
}

func (l *asynqLogger) Debug(args ...any) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...any) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...any) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...any) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...any) {
	l.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}
