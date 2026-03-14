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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skriptvalley/careerdock/internal/config"
	"github.com/skriptvalley/careerdock/internal/domain"
	"github.com/skriptvalley/careerdock/internal/repository"
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

	companyRepo := repository.NewCompanyRepo(db)

	switch target {
	case "all":
		if err := seedCompanies(ctx, companyRepo); err != nil {
			return fmt.Errorf("seed companies: %w", err)
		}
		slog.Info("all seeds applied successfully")

	case "companies":
		if err := seedCompanies(ctx, companyRepo); err != nil {
			return fmt.Errorf("seed companies: %w", err)
		}

	case "flags":
		slog.Warn("feature flag seeding not yet implemented")

	default:
		return fmt.Errorf("unknown seed target: %s (usage: seed [all|companies|flags])", target)
	}

	return nil
}

// seedCompany is the typed JSON representation in seeds/companies.json.
type seedCompany struct {
	Slug              string          `json:"slug"`
	Name              string          `json:"name"`
	LogoURL           *string         `json:"logo_url,omitempty"`
	Description       *string         `json:"description,omitempty"`
	Size              *string         `json:"size,omitempty"`
	Headquarters      *string         `json:"headquarters,omitempty"`
	FoundedYear       *int            `json:"founded_year,omitempty"`
	CareersPageURL    *string         `json:"careers_page_url,omitempty"`
	GlassdoorURL      *string         `json:"glassdoor_url,omitempty"`
	AmbitionboxURL    *string         `json:"ambitionbox_url,omitempty"`
	LinkedinURL       *string         `json:"linkedin_url,omitempty"`
	TechStack         []string        `json:"tech_stack"`
	Domains           []string        `json:"domains"`
	HiringStatus      string          `json:"hiring_status"`
	InterviewPatterns json.RawMessage `json:"interview_patterns,omitempty"`
	CompensationTier  *string         `json:"compensation_tier,omitempty"`
	HasRSU            bool            `json:"has_rsu"`
	HasRSURefresher   bool            `json:"has_rsu_refresher"`
	OfficeModes       []string        `json:"office_modes"`
	CompensationBands json.RawMessage `json:"compensation_bands,omitempty"`
}

func (s seedCompany) toDomain() domain.Company {
	c := domain.Company{
		ID:                uuid.Must(uuid.NewV7()),
		Slug:              s.Slug,
		Name:              s.Name,
		LogoURL:           s.LogoURL,
		Description:       s.Description,
		Headquarters:      s.Headquarters,
		FoundedYear:       s.FoundedYear,
		CareersPageURL:    s.CareersPageURL,
		GlassdoorURL:      s.GlassdoorURL,
		AmbitionboxURL:    s.AmbitionboxURL,
		LinkedinURL:       s.LinkedinURL,
		TechStack:         s.TechStack,
		Domains:           s.Domains,
		HiringStatus:      domain.HiringStatus(s.HiringStatus),
		InterviewPatterns: s.InterviewPatterns,
		CompensationTier:  s.CompensationTier,
		HasRSU:            s.HasRSU,
		HasRSURefresher:   s.HasRSURefresher,
		OfficeModes:       s.OfficeModes,
		CompensationBands: s.CompensationBands,
	}
	if s.Size != nil {
		sz := domain.CompanySize(*s.Size)
		c.Size = &sz
	}
	if c.TechStack == nil {
		c.TechStack = []string{}
	}
	if c.Domains == nil {
		c.Domains = []string{}
	}
	if c.OfficeModes == nil {
		c.OfficeModes = []string{}
	}
	return c
}

// seedCompanies reads seeds/companies.json and upserts companies via the repository.
func seedCompanies(ctx context.Context, repo *repository.CompanyRepo) error {
	seedFile := findSeedFile("companies.json")
	if seedFile == "" {
		slog.Warn("seeds/companies.json not found — skipping company seed")
		return nil
	}

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read seed file: %w", err)
	}

	var seeds []seedCompany
	if err := json.Unmarshal(data, &seeds); err != nil {
		return fmt.Errorf("parse seed file: %w", err)
	}

	slog.Info("seeding companies", "count", len(seeds))

	var upserted, errCount int
	for _, s := range seeds {
		if s.Slug == "" || s.Name == "" {
			slog.Warn("skipping company with missing slug or name", "slug", s.Slug, "name", s.Name)
			errCount++
			continue
		}

		company := s.toDomain()
		if err := repo.Upsert(ctx, &company); err != nil {
			slog.Error("failed to upsert company", "slug", s.Slug, "error", err)
			errCount++
			continue
		}
		upserted++
	}

	slog.Info("company seed complete",
		"total", len(seeds),
		"upserted", upserted,
		"errors", errCount,
	)

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
