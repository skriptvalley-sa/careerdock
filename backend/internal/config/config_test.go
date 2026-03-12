package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMustLoad_PanicsOnMissingRequiredValues(t *testing.T) {
	// Clear any env vars that might satisfy validation.
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("JWT_SECRET")
	viper.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing required config")
		}
	}()

	MustLoad()
}

func TestMustLoad_SucceedsWithRequiredValues(t *testing.T) {
	// Set env vars before MustLoad — Viper's AutomaticEnv reads os.Getenv at access time.
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test?sslmode=disable")
	t.Setenv("REDIS_URL", "localhost:6379")
	t.Setenv("JWT_SECRET", "test-secret")
	// Explicitly set ENVIRONMENT so CI's ENVIRONMENT=test doesn't override the default.
	t.Setenv("ENVIRONMENT", "development")
	// Reset Viper state to avoid cross-test pollution.
	viper.Reset()

	cfg := MustLoad()

	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test?sslmode=disable" {
		t.Errorf("unexpected DATABASE_URL: %s", cfg.DatabaseURL)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default PORT 8080, got %s", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected ENVIRONMENT development, got %s", cfg.Environment)
	}
}

func TestLogLevel(t *testing.T) {
	cfg := &Config{Environment: "development"}
	if cfg.LogLevel().String() != "DEBUG" {
		t.Errorf("expected DEBUG for development, got %s", cfg.LogLevel().String())
	}

	cfg.Environment = "production"
	if cfg.LogLevel().String() != "INFO" {
		t.Errorf("expected INFO for production, got %s", cfg.LogLevel().String())
	}
}

func TestAllowedOriginsParsing(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test?sslmode=disable")
	t.Setenv("REDIS_URL", "localhost:6379")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:3000, https://careerdock.skriptvalley.com")
	viper.Reset()

	cfg := MustLoad()

	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[1] != "https://careerdock.skriptvalley.com" {
		t.Errorf("unexpected second origin: %s", cfg.AllowedOrigins[1])
	}
}
