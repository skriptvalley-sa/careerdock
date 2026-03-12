// Package config loads and validates application configuration from
// environment variables (12-factor). Viper reads .env in development
// and real env vars in production.
package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	// Server
	Port        string
	Environment string // "development", "test", "staging", "production"

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Auth
	JWTSecret string

	// AI Providers
	ClaudeAPIKey string
	OpenAIAPIKey string

	// Payments
	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string

	// Email
	ResendAPIKey string
	FromEmail    string

	// S3 / MinIO
	S3 S3Config

	// CORS
	AllowedOrigins []string

	// Sentry
	SentryDSN string
}

// S3Config holds S3/MinIO configuration.
type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	ResumeBucket    string
	LogoBucket      string
	UsePathStyle    bool
}

// MustLoad loads configuration and panics on failure.
// This is called once at startup — a missing required value is fatal.
func MustLoad() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("FROM_EMAIL", "noreply@careerdock.skriptvalley.com")
	viper.SetDefault("S3_USE_PATH_STYLE", true)
	viper.SetDefault("S3_RESUME_BUCKET", "careerdock-resumes")
	viper.SetDefault("S3_LOGO_BUCKET", "careerdock-logos")
	viper.SetDefault("S3_REGION", "us-east-1")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000")

	// Ignore error — env vars are sufficient in production where no .env exists.
	_ = viper.ReadInConfig()

	// Build config from explicit Get calls — this correctly respects
	// AutomaticEnv, defaults, and .env file values in priority order.
	cfg := &Config{
		Port:        viper.GetString("PORT"),
		Environment: viper.GetString("ENVIRONMENT"),

		DatabaseURL: viper.GetString("DATABASE_URL"),
		RedisURL:    viper.GetString("REDIS_URL"),
		JWTSecret:   viper.GetString("JWT_SECRET"),

		ClaudeAPIKey: viper.GetString("CLAUDE_API_KEY"),
		OpenAIAPIKey: viper.GetString("OPENAI_API_KEY"),

		RazorpayKeyID:         viper.GetString("RAZORPAY_KEY_ID"),
		RazorpayKeySecret:     viper.GetString("RAZORPAY_KEY_SECRET"),
		RazorpayWebhookSecret: viper.GetString("RAZORPAY_WEBHOOK_SECRET"),

		ResendAPIKey: viper.GetString("RESEND_API_KEY"),
		FromEmail:    viper.GetString("FROM_EMAIL"),

		S3: S3Config{
			Endpoint:        viper.GetString("S3_ENDPOINT"),
			Region:          viper.GetString("S3_REGION"),
			AccessKeyID:     viper.GetString("S3_ACCESS_KEY_ID"),
			SecretAccessKey: viper.GetString("S3_SECRET_ACCESS_KEY"),
			ResumeBucket:    viper.GetString("S3_RESUME_BUCKET"),
			LogoBucket:      viper.GetString("S3_LOGO_BUCKET"),
			UsePathStyle:    viper.GetBool("S3_USE_PATH_STYLE"),
		},

		SentryDSN: viper.GetString("SENTRY_DSN"),
	}

	// Parse comma-separated origins
	if origins := viper.GetString("ALLOWED_ORIGINS"); origins != "" {
		cfg.AllowedOrigins = strings.Split(origins, ",")
		for i := range cfg.AllowedOrigins {
			cfg.AllowedOrigins[i] = strings.TrimSpace(cfg.AllowedOrigins[i])
		}
	}

	cfg.validate()
	return cfg
}

// LogLevel returns the slog level appropriate for the current environment.
func (c *Config) LogLevel() slog.Level {
	switch c.Environment {
	case "development", "test":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// validate checks that required configuration values are present.
// Panics with the name of the first missing value.
func (c *Config) validate() {
	required := map[string]string{
		"DATABASE_URL": c.DatabaseURL,
		"REDIS_URL":    c.RedisURL,
		"JWT_SECRET":   c.JWTSecret,
	}
	for name, val := range required {
		if val == "" {
			panic("required config missing: " + name)
		}
	}
}
