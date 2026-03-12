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

	"github.com/skriptvalley/careerdock/internal/config"
	"github.com/skriptvalley/careerdock/internal/handler"
	"github.com/skriptvalley/careerdock/internal/middleware"
	"github.com/skriptvalley/careerdock/internal/repository"
	"github.com/skriptvalley/careerdock/internal/service"
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

	// 5. Build service layer
	authSvc := service.NewAuthService(userRepo, tokenRepo, txr, redisClient, cfg.JWTSecret)
	svc := service.NewServices(authSvc, db, redisClient, version, cfg.IsProduction())

	// 6. Build handler layer + mount routes
	r := chi.NewRouter()

	// Global middleware (order matters: outermost first)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(cfg.AllowedOrigins))

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
