// Package main provides the database migration runner.
//
// Usage:
//
//	go run ./cmd/migrate up        # Apply all pending migrations
//	go run ./cmd/migrate down [N]  # Roll back N migrations (default: 1)
//	go run ./cmd/migrate version   # Print current migration version
//	go run ./cmd/migrate force V   # Force set version V (for recovery)
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/skriptvalley/careerdock/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.MustLoad()
	command := os.Args[1]

	// Resolve migrations directory relative to the binary or working directory.
	migrationsDir := findMigrationsDir()
	sourceURL := "file://" + migrationsDir

	m, err := migrate.New(sourceURL, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("warning: failed to close source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("warning: failed to close database: %v", dbErr)
		}
	}()

	switch command {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration up failed: %v", err) //nolint:gocritic // acceptable: deferred close runs on normal exit paths
		}
		if err == migrate.ErrNoChange {
			fmt.Println("No pending migrations.")
		} else {
			fmt.Println("Migrations applied successfully.")
		}

	case "down":
		steps := 1
		if len(os.Args) >= 3 {
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil || steps < 1 {
				log.Fatalf("invalid step count: %s", os.Args[2])
			}
		}
		err = m.Steps(-steps)
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration down failed: %v", err)
		}
		fmt.Printf("Rolled back %d migration(s).\n", steps)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("failed to get version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: migrate force VERSION")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid version: %s", os.Args[2])
		}
		err = m.Force(version)
		if err != nil {
			log.Fatalf("force version failed: %v", err)
		}
		fmt.Printf("Forced version to %d.\n", version)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// findMigrationsDir looks for the migrations directory in common locations.
func findMigrationsDir() string {
	candidates := []string{
		"migrations",         // when run from backend/
		"backend/migrations", // when run from project root
		"../migrations",      // when binary is in backend/bin/
	}

	for _, dir := range candidates {
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}

	log.Fatal("migrations directory not found — run from backend/ or project root")
	return ""
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `CareerDock Migration Runner

Usage:
  go run ./cmd/migrate <command> [args]

Commands:
  up              Apply all pending migrations
  down [N]        Roll back N migrations (default: 1)
  version         Print current migration version
  force VERSION   Force set version (for recovery from dirty state)
`)
}
