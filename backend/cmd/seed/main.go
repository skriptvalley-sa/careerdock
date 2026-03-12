// Package main is the entry point for the CareerDock data seeder.
//
// Usage:
//
//	go run ./cmd/seed/                  # Seed all data
//	go run ./cmd/seed/ companies        # Seed companies only
//	go run ./cmd/seed/ flags            # Seed feature flags only
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Load config
	cfg := config.MustLoad()

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel(),
	})
	slog.SetDefault(slog.New(logHandler))

	// 2. Connect to database
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()
	slog.Info("connected to database")

	// 3. Determine what to seed
	target := "all"
	if len(os.Args) >= 2 {
		target = os.Args[1]
	}

	switch target {
	case "all":
		if err := seedCompanies(ctx, db); err != nil {
			return fmt.Errorf("seed companies: %w", err)
		}
		slog.Info("all seeds applied successfully")

	case "companies":
		if err := seedCompanies(ctx, db); err != nil {
			return fmt.Errorf("seed companies: %w", err)
		}

	case "flags":
		slog.Warn("feature flag seeding not yet implemented")

	default:
		return fmt.Errorf("unknown seed target: %s (usage: seed [all|companies|flags])", target)
	}

	return nil
}

// seedCompanies reads seeds/companies.json and inserts companies.
func seedCompanies(ctx context.Context, db *pgxpool.Pool) error {
	seedFile := findSeedFile("companies.json")
	if seedFile == "" {
		slog.Warn("seeds/companies.json not found — skipping company seed")
		return nil
	}

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read seed file: %w", err)
	}

	// Parse as generic JSON array for now — will be typed in Sprint 1
	var companies []map[string]any
	if err := json.Unmarshal(data, &companies); err != nil {
		return fmt.Errorf("parse seed file: %w", err)
	}

	slog.Info("seeding companies", "count", len(companies))

	// TODO (Sprint 1): Use CompanyRepository.Create for proper upsert
	_ = ctx
	_ = db
	slog.Warn("company seed insert not yet implemented — seed file parsed successfully")

	return nil
}

// findSeedFile looks for a seed file in common locations.
func findSeedFile(name string) string {
	candidates := []string{
		filepath.Join("seeds", name),
		filepath.Join("backend", "seeds", name),
		filepath.Join("..", "seeds", name),
	}
	for _, path := range candidates {
		abs, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return ""
}
