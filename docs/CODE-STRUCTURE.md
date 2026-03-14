# CareerDock — Code Structure & Developer Workflow

> **Version:** 1.0
> **Status:** Draft (Phase 4)
> **Last updated:** 2026-03-12
> **Depends on:** [PRD.md](./PRD.md), [ARCHITECTURE.md](./ARCHITECTURE.md), [LLD/](./LLD/)

---

## 1. Repository Layout

```
careerdock/
├── docs/                        # Design documents (this folder)
│   ├── PRD.md
│   ├── ARCHITECTURE.md
│   ├── CODE-STRUCTURE.md        # ← this file
│   ├── LLD/
│   │   ├── database.md
│   │   ├── api.md
│   │   ├── ai-service.md
│   │   ├── frontend.md
│   │   └── payments.md
│   ├── ADR/                     # Architecture Decision Records
│   ├── STATUS.md
│   ├── CLAUDE-PROMPT.md
│   └── USAGE-GUIDE.md
│
├── backend/                     # Go backend (API + Worker + CLI tools)
│   ├── cmd/                     # Entry points
│   │   ├── api/                 # HTTP server
│   │   │   └── main.go
│   │   ├── worker/              # Asynq job processor
│   │   │   └── main.go
│   │   ├── migrate/             # DB migration runner
│   │   │   └── main.go
│   │   └── seed/                # Seed data loader
│   │       └── main.go
│   ├── internal/                # Private application code
│   │   ├── config/              # Viper configuration
│   │   ├── domain/              # Business entities, interfaces, errors
│   │   ├── handler/             # HTTP handlers (thin)
│   │   ├── service/             # Business logic
│   │   ├── repository/          # Database access (pgx)
│   │   ├── ai/                  # LLM provider abstraction
│   │   │   ├── provider.go      # Interface definition
│   │   │   ├── claude.go        # Claude implementation
│   │   │   ├── openai.go        # OpenAI implementation
│   │   │   ├── cache.go         # Redis AI result cache
│   │   │   └── prompts/         # Prompt templates
│   │   ├── payment/             # Razorpay integration
│   │   ├── email/               # Resend / SES integration
│   │   ├── storage/             # S3 / MinIO file operations
│   │   ├── pdf/                 # PDF text extraction
│   │   ├── middleware/          # Auth, rate limiting, CORS, logging, request ID
│   │   ├── worker/              # Asynq task definitions and handlers
│   │   └── testutil/            # Shared test helpers, fixtures, factories
│   ├── migrations/              # SQL migration files (golang-migrate)
│   │   ├── 000001_initial_schema.up.sql
│   │   ├── 000001_initial_schema.down.sql
│   │   └── ...
│   ├── seeds/                   # Seed data (JSON files)
│   │   └── companies.json
│   ├── go.mod
│   └── go.sum
│
├── frontend/                    # Next.js frontend
│   ├── src/
│   │   ├── app/                 # App Router pages and layouts
│   │   │   ├── (public)/        # Route group: public pages (SSR)
│   │   │   │   ├── companies/
│   │   │   │   │   ├── page.tsx
│   │   │   │   │   └── [slug]/
│   │   │   │   │       └── page.tsx
│   │   │   │   └── pricing/
│   │   │   │       └── page.tsx
│   │   │   ├── (auth)/          # Route group: login, register, etc.
│   │   │   │   ├── login/
│   │   │   │   ├── register/
│   │   │   │   ├── forgot-password/
│   │   │   │   └── reset-password/
│   │   │   ├── (dashboard)/     # Route group: authenticated pages
│   │   │   │   ├── dashboard/
│   │   │   │   ├── lists/
│   │   │   │   ├── resumes/
│   │   │   │   ├── ats/
│   │   │   │   ├── curated-lists/
│   │   │   │   └── settings/
│   │   │   ├── (admin)/         # Route group: admin pages
│   │   │   │   └── admin/
│   │   │   │       ├── users/
│   │   │   │       ├── companies/
│   │   │   │       ├── payments/
│   │   │   │       ├── features/
│   │   │   │       └── ai/
│   │   │   ├── layout.tsx       # Root layout
│   │   │   └── page.tsx         # Landing page
│   │   ├── components/          # Reusable UI components
│   │   │   ├── ui/              # Primitives (Button, Input, Modal, etc.)
│   │   │   ├── layout/          # Header, Footer, Sidebar, Navigation
│   │   │   ├── company/         # CompanyCard, CompanyFilters, etc.
│   │   │   ├── list/            # ListCard, ApplicationTracker, etc.
│   │   │   ├── resume/          # ResumeUpload, ATSScoreCard, etc.
│   │   │   └── admin/           # Admin-specific components
│   │   ├── hooks/               # Custom React hooks
│   │   │   ├── use-auth.ts
│   │   │   ├── use-companies.ts
│   │   │   └── use-sse.ts
│   │   ├── lib/                 # Framework utilities
│   │   │   ├── api-client.ts    # Typed fetch wrapper
│   │   │   ├── query-keys.ts    # TanStack Query key factory
│   │   │   └── utils.ts         # General helpers
│   │   ├── store/               # Zustand stores
│   │   │   ├── auth-store.ts
│   │   │   └── ui-store.ts
│   │   ├── types/               # Shared TypeScript types
│   │   │   ├── api.ts           # API response/request shapes
│   │   │   ├── company.ts
│   │   │   ├── resume.ts
│   │   │   └── user.ts
│   │   └── styles/              # Global styles
│   │       └── globals.css      # Tailwind directives + custom CSS
│   ├── public/                  # Static assets
│   │   ├── favicon.ico
│   │   └── sw.js                # Service Worker for offline company cache
│   ├── next.config.js
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   └── package-lock.json
│
├── infra/                       # Infrastructure as Code
│   ├── docker/
│   │   ├── Dockerfile.api       # Go API server
│   │   ├── Dockerfile.worker    # Go Asynq worker
│   │   └── nginx.conf           # Reverse proxy config
│   └── terraform/               # AWS infra (Phase 6)
│       └── .gitkeep
│
├── scripts/                     # Developer scripts
│   ├── dev.sh                   # Local environment manager
│   ├── setup.sh                 # Compatibility wrapper for dev.sh setup
│   ├── gen-migration.sh         # Create next sequential migration file pair
│   └── GUIDE.md                 # Manual for script usage
│
├── .github/
│   ├── workflows/
│   │   ├── ci.yml               # PR checks: lint, test, build
│   │   └── deploy.yml           # Tag-triggered deploy (Phase 6)
│   └── pull_request_template.md
│
├── docker-compose.yml           # Local dev environment
├── docker-compose.prod.yml      # Production services
├── Makefile                     # Developer commands
├── .env.example                 # Environment variable template
├── .gitignore
├── .golangci.yml                # Go linter configuration
├── .prettierrc                  # Frontend formatter config
├── .eslintrc.json               # Frontend linter config
└── README.md
```

### 1.1 Why No Separate `admin/` Directory

The CLAUDE-PROMPT.md mentions an `admin/` directory. We integrate admin pages under `frontend/src/app/(admin)/admin/` instead:

- Admin shares the same component library, auth hooks, and API client as the user-facing app.
- Route-group-based organisation (`(admin)/`) gives admin pages their own layout without duplicating shared code.
- Role-based middleware in the backend controls access — the frontend just renders what the API returns.

### 1.2 Why No `backend/pkg/`

`pkg/` is for code intended to be imported by external consumers. CareerDock is a single application — nothing needs to be externally importable. All Go code lives under `internal/` which the compiler enforces as private.

If a shared utility is truly needed across `cmd/` entry points (e.g., a custom logger wrapper), it goes in `internal/` since all entry points are within the same module.

---

## 2. Backend Architecture

### 2.1 Layering Rules

```
┌─────────────────────────────────────────────────────┐
│                    cmd/api/main.go                    │
│              (wires everything, starts server)        │
└────────────────────────┬────────────────────────────┘
                         │ constructs
┌────────────────────────▼────────────────────────────┐
│                     middleware/                       │
│         Auth, RateLimit, CORS, RequestID, Logger     │
└────────────────────────┬────────────────────────────┘
                         │ wraps
┌────────────────────────▼────────────────────────────┐
│                      handler/                        │
│          Validate input → call service → respond     │
│          NO business logic. NO direct DB access.     │
└────────────────────────┬────────────────────────────┘
                         │ calls
┌────────────────────────▼────────────────────────────┐
│                      service/                        │
│       Business logic. Orchestrates repositories,     │
│       AI providers, payment gateway, email sender.   │
│       Owns transaction boundaries.                   │
└──────┬─────────┬──────────┬────────────┬────────────┘
       │         │          │            │
       ▼         ▼          ▼            ▼
  repository/   ai/     payment/     email/
   (pgx)     (Claude/   (Razorpay)  (Resend)
              OpenAI)
```

**Hard rules:**
1. Handlers never import `repository`, `ai`, `payment`, or `email`.
2. Repositories never import `service` or `handler`.
3. Services may depend on multiple repositories but never on handlers.
4. All cross-layer communication uses interfaces defined in `domain/`.

### 2.2 The `domain/` Package

`domain/` is the **dependency inversion hub** — it defines all business entities and interfaces. Every other package depends on `domain/`, but `domain/` depends on nothing (except stdlib).

```
internal/domain/
├── entities.go          # User, Company, Resume, List, etc.
├── errors.go            # Application-level error types
├── interfaces.go        # Repository, AI provider, Payment, Email interfaces
└── enums.go             # Role, ApplicationStatus, CreditType constants
```

**Entity example:**

```go
// domain/entities.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type User struct {
    ID            uuid.UUID
    Email         string
    PasswordHash  string
    Name          string
    Role          Role
    PremiumSince  *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

type Company struct {
    ID              uuid.UUID
    Name            string
    Slug            string
    Website         *string
    Description     *string
    LogoURL         *string
    Size            *string
    Founded         *int
    Headquarters    *string
    TechStack       []string
    Domains         []string
    CareersPageURL  *string
    IsVerified      bool
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Interface example:**

```go
// domain/interfaces.go
package domain

import "context"

// --- Repository interfaces ---

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
}

type CompanyRepository interface {
    List(ctx context.Context, filter CompanyFilter) ([]Company, string, error) // entities, next cursor, error
    GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
    GetBySlug(ctx context.Context, slug string) (*Company, error)
    Search(ctx context.Context, query string, filter CompanyFilter) ([]Company, string, error)
    Create(ctx context.Context, company *Company) error
    Update(ctx context.Context, company *Company) error
}

// ... ResumeRepository, ListRepository, ATSCheckRepository, etc.

// --- External service interfaces ---

type AIProvider interface {
    ParseResume(ctx context.Context, req *ParseResumeRequest) (*ParsedResume, error)
    ScoreATSGeneral(ctx context.Context, req *ATSGeneralRequest) (*ATSResult, error)
    ScoreATSCompany(ctx context.Context, req *ATSCompanyRequest) (*ATSResult, error)
    ScoreATSJob(ctx context.Context, req *ATSJobRequest) (*ATSResult, error)
    CurateCompanyList(ctx context.Context, req *CurateListRequest) (*CuratedList, error)
    EnrichCompanyProfile(ctx context.Context, req *EnrichRequest) (*CompanyProfile, error)
}

type PaymentGateway interface {
    CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)
    VerifyPayment(ctx context.Context, req *VerifyPaymentRequest) (*PaymentVerification, error)
}

type EmailSender interface {
    Send(ctx context.Context, msg *EmailMessage) error
}

type FileStore interface {
    Upload(ctx context.Context, key string, data []byte, contentType string) error
    Download(ctx context.Context, key string) ([]byte, error)
    GenerateSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
    Delete(ctx context.Context, key string) error
}
```

**Error types:**

```go
// domain/errors.go
package domain

import "fmt"

type ErrorCode string

const (
    ErrCodeNotFound       ErrorCode = "NOT_FOUND"
    ErrCodeConflict       ErrorCode = "CONFLICT"
    ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
    ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
    ErrCodeForbidden      ErrorCode = "FORBIDDEN"
    ErrCodeRateLimited    ErrorCode = "RATE_LIMITED"
    ErrCodeInsufficientCredits ErrorCode = "INSUFFICIENT_CREDITS"
    ErrCodePaymentFailed  ErrorCode = "PAYMENT_FAILED"
    ErrCodeAIUnavailable  ErrorCode = "AI_UNAVAILABLE"
    ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
)

// AppError is the standard application error type.
// Handlers map these to HTTP status codes.
type AppError struct {
    Code    ErrorCode
    Message string
    Details map[string]any
    Err     error // wrapped original error (never sent to client)
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructors for common errors
func NotFound(resource string, id any) *AppError {
    return &AppError{
        Code:    ErrCodeNotFound,
        Message: fmt.Sprintf("%s not found", resource),
        Details: map[string]any{"id": id},
    }
}

func ValidationError(message string, details map[string]any) *AppError {
    return &AppError{
        Code:    ErrCodeValidation,
        Message: message,
        Details: details,
    }
}
```

**Enum constants:**

```go
// domain/enums.go
package domain

type Role string

const (
    RoleUser      Role = "user"
    RoleModerator Role = "moderator"
    RoleAdmin     Role = "admin"
)

type ApplicationStatus string

const (
    StatusWishlist   ApplicationStatus = "wishlist"
    StatusApplied    ApplicationStatus = "applied"
    StatusScreening  ApplicationStatus = "screening"
    StatusInterviewing ApplicationStatus = "interviewing"
    StatusOffer      ApplicationStatus = "offer"
    StatusAccepted   ApplicationStatus = "accepted"
    StatusRejected   ApplicationStatus = "rejected"
    StatusWithdrawn  ApplicationStatus = "withdrawn"
)

type CreditType string

const (
    CreditResumeUpload CreditType = "resume_upload"
    CreditATSCheck     CreditType = "ats_check"
    CreditCuratedList  CreditType = "curated_list"
    CreditCVGeneration CreditType = "cv_generation"
)
```

### 2.3 Dependency Injection — Constructor Injection with `App` Struct

No DI framework. Each layer accepts its dependencies as constructor arguments. A central `App` struct wires everything together in `cmd/api/main.go`.

```go
// cmd/api/main.go
package main

import (
    "context"
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
    // ... other imports
)

func main() {
    // 1. Load config
    cfg := config.MustLoad()

    // 2. Set up logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: cfg.LogLevel(),
    }))
    slog.SetDefault(logger)

    // 3. Connect infrastructure
    dbPool := mustConnectDB(cfg.DatabaseURL)
    defer dbPool.Close()

    redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
    defer redisClient.Close()

    // 4. Build repository layer
    repos := &repository.Repositories{
        User:    repository.NewUserRepo(dbPool),
        Company: repository.NewCompanyRepo(dbPool),
        Resume:  repository.NewResumeRepo(dbPool),
        List:    repository.NewListRepo(dbPool),
        ATS:     repository.NewATSCheckRepo(dbPool),
        Payment: repository.NewPaymentRepo(dbPool),
        Credit:  repository.NewCreditRepo(dbPool),
        // ...
    }

    // 5. Build external service adapters
    aiProvider := ai.NewFallbackProvider(
        ai.NewClaudeProvider(cfg.ClaudeAPIKey),
        ai.NewOpenAIProvider(cfg.OpenAIAPIKey),
        redisClient,
    )
    paymentGW := payment.NewRazorpay(cfg.RazorpayKey, cfg.RazorpaySecret)
    emailSender := email.NewResend(cfg.ResendAPIKey)
    fileStore := storage.NewS3(cfg.S3Config)

    // 6. Build service layer
    services := &service.Services{
        Auth:    service.NewAuthService(repos.User, redisClient, cfg.JWTSecret),
        User:    service.NewUserService(repos.User),
        Company: service.NewCompanyService(repos.Company),
        Resume:  service.NewResumeService(repos.Resume, fileStore, aiProvider),
        List:    service.NewListService(repos.List, repos.User),
        ATS:     service.NewATSService(repos.ATS, repos.Resume, aiProvider),
        Payment: service.NewPaymentService(repos.Payment, repos.Credit, paymentGW),
        Credit:  service.NewCreditService(repos.Credit),
        Admin:   service.NewAdminService(repos, aiProvider),
        // ...
    }

    // 7. Build handler layer and mount routes
    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(middleware.Logger(logger))
    r.Use(middleware.Recoverer)
    r.Use(middleware.CORS(cfg.AllowedOrigins))

    handler.MountRoutes(r, services, middleware.NewAuth(services.Auth))

    // 8. Start server with graceful shutdown
    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        logger.Info("server starting", "port", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("server failed", "error", err)
            os.Exit(1)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    logger.Info("server shutting down")
    srv.Shutdown(ctx)
}
```

**Worker entry point follows the same pattern:**

```go
// cmd/worker/main.go — wires repos + AI + storage, registers Asynq handlers, starts processing.
```

### 2.4 Handler Pattern

Handlers are thin — validate, delegate, respond. Each handler struct takes a service interface.

```go
// handler/company.go
package handler

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/skriptvalley/careerdock/internal/domain"
)

type CompanyHandler struct {
    service domain.CompanyService // interface, not concrete type
}

func NewCompanyHandler(svc domain.CompanyService) *CompanyHandler {
    return &CompanyHandler{service: svc}
}

func (h *CompanyHandler) List(w http.ResponseWriter, r *http.Request) {
    filter, err := decodeCompanyFilter(r)
    if err != nil {
        respondError(w, domain.ValidationError("invalid filter parameters", nil))
        return
    }

    companies, cursor, err := h.service.List(r.Context(), filter)
    if err != nil {
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, PaginatedResponse{
        Data:       companies,
        Pagination: &Pagination{NextCursor: cursor, HasMore: cursor != ""},
    })
}

func (h *CompanyHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, "slug")

    company, err := h.service.GetBySlug(r.Context(), slug)
    if err != nil {
        respondError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, DataResponse{Data: company})
}
```

**Shared response helpers** in `handler/response.go`:

```go
// handler/response.go
package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "github.com/skriptvalley/careerdock/internal/domain"
)

type DataResponse struct {
    Data any `json:"data"`
}

type PaginatedResponse struct {
    Data       any         `json:"data"`
    Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
    NextCursor string `json:"next_cursor,omitempty"`
    HasMore    bool   `json:"has_more"`
}

type ErrorResponse struct {
    Error ErrorBody `json:"error"`
}

type ErrorBody struct {
    Code    string         `json:"code"`
    Message string         `json:"message"`
    Details map[string]any `json:"details,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, err error) {
    var appErr *domain.AppError
    if errors.As(err, &appErr) {
        status := mapErrorCodeToHTTP(appErr.Code)
        respondJSON(w, status, ErrorResponse{
            Error: ErrorBody{
                Code:    string(appErr.Code),
                Message: appErr.Message,
                Details: appErr.Details,
            },
        })
        return
    }
    // Unexpected error — don't leak internals
    respondJSON(w, http.StatusInternalServerError, ErrorResponse{
        Error: ErrorBody{
            Code:    string(domain.ErrCodeInternal),
            Message: "An unexpected error occurred",
        },
    })
}

func mapErrorCodeToHTTP(code domain.ErrorCode) int {
    switch code {
    case domain.ErrCodeNotFound:
        return http.StatusNotFound
    case domain.ErrCodeConflict:
        return http.StatusConflict
    case domain.ErrCodeValidation:
        return http.StatusUnprocessableEntity
    case domain.ErrCodeUnauthorized:
        return http.StatusUnauthorized
    case domain.ErrCodeForbidden:
        return http.StatusForbidden
    case domain.ErrCodeRateLimited:
        return http.StatusTooManyRequests
    case domain.ErrCodeInsufficientCredits:
        return http.StatusPaymentRequired
    default:
        return http.StatusInternalServerError
    }
}
```

### 2.5 Route Mounting

All routes mounted from a single function. Each handler registers its own sub-routes.

```go
// handler/routes.go
package handler

import (
    "github.com/go-chi/chi/v5"
    "github.com/skriptvalley/careerdock/internal/middleware"
    "github.com/skriptvalley/careerdock/internal/service"
)

func MountRoutes(r chi.Router, svc *service.Services, auth *middleware.Auth) {
    // --- Public ---
    companyH := NewCompanyHandler(svc.Company)
    r.Route("/api/companies", func(r chi.Router) {
        r.Get("/", companyH.List)
        r.Get("/search", companyH.Search)
        r.Get("/{slug}", companyH.GetBySlug)
    })

    // --- Auth (no token required) ---
    authH := NewAuthHandler(svc.Auth)
    r.Route("/api/auth", func(r chi.Router) {
        r.Post("/register", authH.Register)
        r.Post("/login", authH.Login)
        r.Post("/refresh", authH.Refresh)
        r.Post("/forgot-password", authH.ForgotPassword)
        r.Post("/reset-password", authH.ResetPassword)
        r.Get("/verify-email/{token}", authH.VerifyEmail)
    })

    // --- Authenticated ---
    r.Group(func(r chi.Router) {
        r.Use(auth.RequireAuthenticated)

        r.Post("/api/auth/logout", authH.Logout)
        r.Get("/api/auth/me", authH.Me)

        listH := NewListHandler(svc.List)
        r.Route("/api/lists", func(r chi.Router) {
            r.Get("/", listH.List)
            r.Post("/", listH.Create)
            r.Get("/{id}", listH.Get)
            r.Put("/{id}", listH.Update)
            r.Delete("/{id}", listH.Delete)
            r.Post("/{id}/entries", listH.AddEntry)
            // ...
        })

        // --- Premium ---
        r.Group(func(r chi.Router) {
            r.Use(auth.RequirePremium)

            resumeH := NewResumeHandler(svc.Resume)
            r.Route("/api/resumes", func(r chi.Router) {
                r.Get("/", resumeH.List)
                r.Post("/", resumeH.Upload)
                r.Get("/{id}", resumeH.Get)
                r.Delete("/{id}", resumeH.Archive)
            })

            atsH := NewATSHandler(svc.ATS)
            r.Route("/api/ats", func(r chi.Router) {
                r.Post("/company", atsH.CheckCompany)
                r.Post("/job", atsH.CheckJob)
                r.Get("/{id}", atsH.GetResult)
                r.Get("/", atsH.ListResults)
            })
        })

        paymentH := NewPaymentHandler(svc.Payment)
        r.Route("/api/payments", func(r chi.Router) {
            r.Post("/orders", paymentH.CreateOrder)
            r.Get("/history", paymentH.History)
        })
    })

    // --- Webhooks (signature-verified, no JWT) ---
    r.Post("/api/webhooks/razorpay", NewPaymentHandler(svc.Payment).Webhook)

    // --- Admin ---
    r.Group(func(r chi.Router) {
        r.Use(auth.RequireAuthenticated)
        r.Use(auth.RequireAdmin)

        adminH := NewAdminHandler(svc.Admin)
        r.Route("/api/admin", func(r chi.Router) {
            r.Get("/users", adminH.ListUsers)
            r.Get("/companies", adminH.ListCompanies)
            r.Post("/companies", adminH.CreateCompany)
            r.Put("/companies/{id}", adminH.UpdateCompany)
            r.Get("/feature-flags", adminH.ListFeatureFlags)
            r.Put("/feature-flags/{id}", adminH.UpdateFeatureFlag)
            // ...
        })
    })

    // --- SSE (authenticated) ---
    r.Group(func(r chi.Router) {
        r.Use(auth.RequireAuthenticated)
        r.Get("/api/events", NewSSEHandler(svc).Stream)
    })

    // --- Health ---
    r.Get("/api/health", HealthCheck)
}
```

### 2.6 Service Layer — Transaction Handling

Services own transaction boundaries. A `Transactor` interface allows services to wrap multiple repository calls in a single DB transaction without leaking `pgx` into the domain layer.

```go
// domain/interfaces.go (addition)
type Transactor interface {
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

```go
// repository/transactor.go
package repository

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type ctxKeyTx struct{}

type PgxTransactor struct {
    pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *PgxTransactor {
    return &PgxTransactor{pool: pool}
}

func (t *PgxTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := t.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Store tx in context so repos can pick it up
    txCtx := context.WithValue(ctx, ctxKeyTx{}, tx)
    if err := fn(txCtx); err != nil {
        return err
    }
    return tx.Commit(ctx)
}

// connFromCtx returns the tx from context if present, otherwise the pool.
func connFromCtx(ctx context.Context, pool *pgxpool.Pool) pgxQuerier {
    if tx, ok := ctx.Value(ctxKeyTx{}).(pgxQuerier); ok {
        return tx
    }
    return pool
}
```

**Usage in a service:**

```go
// service/payment.go
func (s *PaymentService) HandleWebhook(ctx context.Context, event *domain.RazorpayEvent) error {
    return s.tx.WithTx(ctx, func(ctx context.Context) error {
        // 1. Update payment status
        if err := s.paymentRepo.UpdateStatus(ctx, event.OrderID, domain.PaymentCaptured); err != nil {
            return err
        }
        // 2. Allocate credits
        if err := s.creditRepo.Allocate(ctx, event.UserID, event.Credits); err != nil {
            return err
        }
        // 3. Log transaction
        return s.creditRepo.LogTransaction(ctx, event.UserID, event.Credits, "payment", event.OrderID)
    })
}
```

### 2.7 Repository Pattern

Repositories use raw SQL via `pgx`. No ORM. Each repo accepts the pool and uses `connFromCtx` to transparently support transactions.

```go
// repository/user.go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/skriptvalley/careerdock/internal/domain"
)

type UserRepo struct {
    pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
    return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    conn := connFromCtx(ctx, r.pool)

    var u domain.User
    err := conn.QueryRow(ctx,
        `SELECT id, email, password_hash, name, role, premium_since,
                created_at, updated_at, deleted_at
         FROM users
         WHERE email = $1 AND deleted_at IS NULL`,
        email,
    ).Scan(
        &u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.PremiumSince,
        &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
    )
    if err != nil {
        return nil, mapPgError(err, "user")
    }
    return &u, nil
}
```

---

## 3. Frontend Architecture

### 3.1 Route Groups

Next.js App Router route groups (parenthesised directories) provide layout separation without affecting URL structure:

| Route Group | Layout | Auth Required | Purpose |
|-------------|--------|:---:|---------|
| `(public)/` | Minimal header + footer | No | Company directory, pricing, landing |
| `(auth)/` | Centred card layout | No | Login, register, password reset |
| `(dashboard)/` | Sidebar + top nav | Yes | Lists, resumes, ATS, settings |
| `(admin)/` | Admin sidebar + top nav | Yes (admin role) | User mgmt, companies, flags |

Each route group has its own `layout.tsx` that wraps pages with the appropriate shell.

### 3.2 State Management Split

| Concern | Tool | Why |
|---------|------|-----|
| Server state (API data) | TanStack Query | Auto caching, revalidation, optimistic updates, loading/error states |
| Client-only UI state | Zustand | Sidebar open/closed, theme, active filters (not persisted) |
| Auth state | Zustand + cookies | Zustand tracks `isAuthenticated` + `user` from `/api/auth/me` on app load; tokens stay in httpOnly cookies |
| Form state | React Hook Form | Validation, error messages, submission handling |

### 3.3 API Client

A thin typed wrapper around `fetch` that handles cookies, errors, and JSON automatically:

```typescript
// lib/api-client.ts

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

type ApiResponse<T> = { data: T };
type PaginatedResponse<T> = { data: T[]; pagination?: { next_cursor: string; has_more: boolean } };
type ApiError = { error: { code: string; message: string; details?: Record<string, unknown> } };

class ApiClient {
  private async request<T>(path: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`${API_BASE}${path}`, {
      ...options,
      credentials: 'include',  // send httpOnly cookies
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!res.ok) {
      const body: ApiError = await res.json();
      throw new AppError(body.error.code, body.error.message, body.error.details);
    }

    return res.json();
  }

  // Typed methods
  companies = {
    list: (cursor?: string) =>
      this.request<PaginatedResponse<Company>>(`/api/companies?cursor=${cursor || ''}`),
    getBySlug: (slug: string) =>
      this.request<ApiResponse<Company>>(`/api/companies/${slug}`),
    search: (query: string, filters?: CompanyFilters) =>
      this.request<PaginatedResponse<Company>>(`/api/companies/search?q=${query}`),
  };

  auth = {
    login: (email: string, password: string) =>
      this.request<ApiResponse<User>>('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    register: (data: RegisterPayload) =>
      this.request<ApiResponse<User>>('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    me: () => this.request<ApiResponse<User>>('/api/auth/me'),
    logout: () => this.request<void>('/api/auth/logout', { method: 'POST' }),
  };

  // ... resumes, lists, ats, payments, admin
}

export const api = new ApiClient();
```

### 3.4 TanStack Query Key Convention

All query keys live in one file to prevent collisions and enable targeted invalidation:

```typescript
// lib/query-keys.ts

export const queryKeys = {
  companies: {
    all:      ['companies'] as const,
    list:     (cursor?: string) => ['companies', 'list', cursor] as const,
    detail:   (slug: string) => ['companies', 'detail', slug] as const,
    search:   (q: string) => ['companies', 'search', q] as const,
  },
  lists: {
    all:      ['lists'] as const,
    detail:   (id: string) => ['lists', 'detail', id] as const,
  },
  resumes: {
    all:      ['resumes'] as const,
    detail:   (id: string) => ['resumes', 'detail', id] as const,
  },
  ats: {
    all:      ['ats'] as const,
    detail:   (id: string) => ['ats', 'detail', id] as const,
  },
  auth: {
    me:       ['auth', 'me'] as const,
  },
  credits: {
    balance:  ['credits', 'balance'] as const,
  },
} as const;
```

### 3.5 Component Naming Conventions

| Category | Pattern | Example |
|----------|---------|---------|
| Pages | `page.tsx` (Next.js convention) | `app/(dashboard)/lists/page.tsx` |
| Layouts | `layout.tsx` | `app/(dashboard)/layout.tsx` |
| UI primitives | PascalCase, single file | `components/ui/Button.tsx` |
| Feature components | PascalCase, feature folder | `components/company/CompanyCard.tsx` |
| Hooks | `use-<name>.ts` (kebab-case file) | `hooks/use-auth.ts` |
| Stores | `<name>-store.ts` | `store/auth-store.ts` |
| Types | `<name>.ts` | `types/company.ts` |
| Utilities | `<name>.ts` | `lib/utils.ts` |

---

## 4. Configuration

### 4.1 Environment Variables (Viper)

All config loaded from environment variables (12-factor). Viper reads `.env` in development and real env vars in production.

```go
// internal/config/config.go
package config

import (
    "log/slog"
    "github.com/spf13/viper"
)

type Config struct {
    // Server
    Port            string `mapstructure:"PORT"`
    Environment     string `mapstructure:"ENVIRONMENT"` // "development", "staging", "production"

    // Database
    DatabaseURL     string `mapstructure:"DATABASE_URL"`

    // Redis
    RedisURL        string `mapstructure:"REDIS_URL"`

    // Auth
    JWTSecret       string `mapstructure:"JWT_SECRET"`

    // AI Providers
    ClaudeAPIKey    string `mapstructure:"CLAUDE_API_KEY"`
    OpenAIAPIKey    string `mapstructure:"OPENAI_API_KEY"`

    // Payments
    RazorpayKey     string `mapstructure:"RAZORPAY_KEY_ID"`
    RazorpaySecret  string `mapstructure:"RAZORPAY_KEY_SECRET"`
    RazorpayWebhookSecret string `mapstructure:"RAZORPAY_WEBHOOK_SECRET"`

    // Email
    ResendAPIKey    string `mapstructure:"RESEND_API_KEY"`
    FromEmail       string `mapstructure:"FROM_EMAIL"`

    // S3
    S3Config        S3Config

    // CORS
    AllowedOrigins  []string `mapstructure:"ALLOWED_ORIGINS"`

    // Sentry
    SentryDSN       string `mapstructure:"SENTRY_DSN"`
}

type S3Config struct {
    Endpoint        string `mapstructure:"S3_ENDPOINT"`   // MinIO in dev
    Region          string `mapstructure:"S3_REGION"`
    AccessKeyID     string `mapstructure:"S3_ACCESS_KEY_ID"`
    SecretAccessKey  string `mapstructure:"S3_SECRET_ACCESS_KEY"`
    ResumeBucket    string `mapstructure:"S3_RESUME_BUCKET"`
    LogoBucket      string `mapstructure:"S3_LOGO_BUCKET"`
    UsePathStyle    bool   `mapstructure:"S3_USE_PATH_STYLE"` // true for MinIO
}

func MustLoad() *Config {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()

    // Defaults
    viper.SetDefault("PORT", "8080")
    viper.SetDefault("ENVIRONMENT", "development")
    viper.SetDefault("S3_USE_PATH_STYLE", true)
    viper.SetDefault("S3_RESUME_BUCKET", "careerdock-resumes")
    viper.SetDefault("S3_LOGO_BUCKET", "careerdock-logos")

    viper.ReadInConfig() // ignore error — env vars sufficient in production

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        panic("failed to unmarshal config: " + err.Error())
    }
    cfg.validate()
    return &cfg
}

func (c *Config) LogLevel() slog.Level {
    if c.Environment == "development" {
        return slog.LevelDebug
    }
    return slog.LevelInfo
}

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
```

### 4.2 `.env.example`

```env
# Server
PORT=8080
ENVIRONMENT=development

# Database
DATABASE_URL=postgres://careerdock:careerdock@localhost:5432/careerdock?sslmode=disable

# Redis
REDIS_URL=localhost:6379

# Auth
JWT_SECRET=change-me-in-production

# AI Providers
CLAUDE_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...

# Payments
RAZORPAY_KEY_ID=rzp_test_...
RAZORPAY_KEY_SECRET=...
RAZORPAY_WEBHOOK_SECRET=...

# Email
RESEND_API_KEY=re_...
FROM_EMAIL=noreply@careerdock.skriptvalley.com

# S3 / MinIO (local dev)
S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_RESUME_BUCKET=careerdock-resumes
S3_LOGO_BUCKET=careerdock-logos
S3_USE_PATH_STYLE=true

# CORS
ALLOWED_ORIGINS=http://localhost:3000

# Sentry
SENTRY_DSN=
```

---

## 5. Version Control Strategy

### 5.1 Branching Model — Trunk-Based Development

```
main ──●──●──●──●──●──●──●──●──●──●──●──●──●──●── (always deployable)
        \   /    \   /       \   /       \   /
         feat/   fix/        feat/       fix/
         login   slug-404   ats-flow    credit-race
```

**Rules:**

| Rule | Detail |
|------|--------|
| `main` is always deployable | Every merge passes CI (lint + test + build) |
| Short-lived feature branches | `feat/<name>`, `fix/<name>` — merged within 1-3 days |
| No long-lived branches | No `develop`, `staging`, or `release/*` branches |
| Squash merge via PR | Each feature = one clean commit on `main` |
| Feature flags for WIP | Incomplete features hidden behind DB-backed feature flags |

### 5.2 Release Strategy — Semantic Versioning via Git Tags

Releases are **git tags on `main`**, not branches. Tags are immutable snapshots — no maintenance overhead.

```
main ──●──●──●──●──●──●──●──●──●──●──●──●──●──●──
              │              │              │
           v1.0.0        v1.1.0        v1.1.1
           (launch)    (new feature)   (bug fix)
```

**Semver rules:**

| Version Bump | When | Example |
|-------------|------|---------|
| `MAJOR` (v2.0.0) | Breaking API changes, DB schema overhauls | Complete auth rewrite |
| `MINOR` (v1.1.0) | New features, backward-compatible changes | Add CV generation feature |
| `PATCH` (v1.0.1) | Bug fixes, security patches, dependency updates | Fix credit allocation race condition |

**Creating a release:**

```bash
# Tag the current main
git tag -a v1.2.0 -m "Release v1.2.0: Add job ATS check feature"
git push origin v1.2.0

# CI sees the tag → builds → deploys (details in DEPLOYMENT.md)
```

**Hotfix on an older version** (rare — only when main has moved significantly):

```bash
# 1. Create temporary branch from the tag
git checkout -b hotfix/v1.0.x v1.0.0

# 2. Cherry-pick the fix
git cherry-pick <commit-sha>

# 3. Tag the patch
git tag -a v1.0.1 -m "Hotfix: Fix critical credit race condition"
git push origin v1.0.1

# 4. Merge fix back to main (if not already there)
git checkout main
git merge hotfix/v1.0.x

# 5. Delete the branch — it served its purpose
git branch -d hotfix/v1.0.x
```

**Deployable artifacts are built from git tags matching `v*`. The tag uniquely identifies the code running in production.** Detailed CI/CD pipeline and environment promotion are defined in [DEPLOYMENT.md](./DEPLOYMENT.md) (Phase 6).

### 5.3 Commit Convention — Conventional Commits

```
<type>(<scope>): <short summary>

<optional body>

<optional footer>
```

**Types:**

| Type | Usage |
|------|-------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `docs` | Documentation only |
| `test` | Adding or correcting tests |
| `chore` | Tooling, CI, dependencies, build scripts |
| `perf` | Performance improvement |
| `style` | Code formatting (no logic change) |

**Scopes** (optional, match package names):

`api`, `worker`, `auth`, `company`, `resume`, `ats`, `payment`, `ai`, `admin`, `frontend`, `infra`, `db`

**Examples:**

```
feat(ats): add job-specific ATS scoring endpoint
fix(payment): prevent duplicate credit allocation on webhook retry
refactor(ai): extract prompt templates into separate files
chore(infra): add Redis health check to docker-compose
docs: update CODE-STRUCTURE with transaction handling pattern
```

### 5.4 PR Template

```markdown
<!-- .github/pull_request_template.md -->

## Summary

<!-- What does this PR do? 1-3 sentences. -->

## Changes

- [ ] Change 1
- [ ] Change 2

## Type

- [ ] feat
- [ ] fix
- [ ] refactor
- [ ] chore
- [ ] docs

## Checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] No new secrets or credentials in code
- [ ] Migration is reversible (if applicable)
- [ ] API changes documented (if applicable)
```

### 5.5 Protected Branch — `main`

GitHub branch protection rules for `main`:

| Rule | Setting |
|------|---------|
| Require PR before merging | Yes |
| Require status checks (CI) | Yes — `lint`, `test`, `build` |
| Require conversation resolution | Yes |
| Allow force pushes | No |
| Allow deletions | No |

---

## 6. Development Workflow

### 6.1 Local Development Environment

**`docker-compose.yml`** provides all infrastructure dependencies:

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: careerdock
      POSTGRES_PASSWORD: careerdock
      POSTGRES_DB: careerdock
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U careerdock"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - miniodata:/data

  minio-init:
    image: minio/mc:latest
    depends_on:
      minio:
        condition: service_started
    entrypoint: >
      /bin/sh -c "
        sleep 2;
        mc alias set local http://minio:9000 minioadmin minioadmin;
        mc mb --ignore-existing local/careerdock-resumes;
        mc mb --ignore-existing local/careerdock-logos;
      "

  mailhog:
    image: mailhog/mailhog:latest
    ports:
      - "1025:1025"   # SMTP
      - "8025:8025"   # Web UI

volumes:
  pgdata:
  miniodata:
```

**What runs locally vs in Docker:**

| Component | Runs In | Why |
|-----------|---------|-----|
| Postgres, Redis, MinIO, Mailhog | Docker Compose | Infrastructure — no local install needed |
| Go API server | Host (with hot reload) | Faster iteration, debugger access |
| Go worker | Host (with hot reload) | Same as above |
| Next.js frontend | Host (`npm run dev`) | HMR is instant, debugger access |

### 6.2 Makefile

```makefile
# Makefile

.PHONY: dev dev-infra dev-api dev-worker dev-frontend \
        test test-backend test-frontend lint build \
        migrate migrate-down migrate-new seed \
        docker-build clean

# ─── Development ─────────────────────────────────────────

## Start all infrastructure (Postgres, Redis, MinIO, Mailhog)
dev-infra:
	docker compose up -d

## Start Go API server with hot reload
dev-api:
	cd backend && air -c .air.api.toml

## Start Asynq worker with hot reload
dev-worker:
	cd backend && air -c .air.worker.toml

## Start Next.js dev server
dev-frontend:
	cd frontend && npm run dev

## Start everything (run in separate terminals or use a process manager)
dev: dev-infra
	@echo "Infrastructure started. Run in separate terminals:"
	@echo "  make dev-api"
	@echo "  make dev-worker"
	@echo "  make dev-frontend"

# ─── Testing ─────────────────────────────────────────────

## Run all backend tests
test-backend:
	cd backend && go test -race -count=1 ./...

## Run frontend tests
test-frontend:
	cd frontend && npm test

## Run all tests
test: test-backend test-frontend

# ─── Linting ─────────────────────────────────────────────

## Run Go linter
lint-backend:
	cd backend && golangci-lint run ./...

## Run frontend linter
lint-frontend:
	cd frontend && npm run lint

## Run all linters
lint: lint-backend lint-frontend

# ─── Database ────────────────────────────────────────────

## Run pending migrations
migrate:
	cd backend && go run ./cmd/migrate/ up

## Rollback last migration
migrate-down:
	cd backend && go run ./cmd/migrate/ down 1

## Create new migration pair (usage: make migrate-new NAME=add_xyz_table)
migrate-new:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-new NAME=description"; exit 1; fi
	./scripts/gen-migration.sh $(NAME)

## Seed database with initial data
seed:
	cd backend && go run ./cmd/seed/

# ─── Build ───────────────────────────────────────────────

## Build Go binaries
build-backend:
	cd backend && CGO_ENABLED=0 go build -o bin/api ./cmd/api/
	cd backend && CGO_ENABLED=0 go build -o bin/worker ./cmd/worker/

## Build frontend
build-frontend:
	cd frontend && npm run build

## Build everything
build: build-backend build-frontend

## Build Docker images
docker-build:
	docker build -f infra/docker/Dockerfile.api -t careerdock-api:latest ./backend
	docker build -f infra/docker/Dockerfile.worker -t careerdock-worker:latest ./backend

# ─── Cleanup ─────────────────────────────────────────────

## Stop all Docker services and remove volumes
clean:
	docker compose down -v
	cd backend && rm -rf bin/
```

### 6.3 Hot Reload — Air

[Air](https://github.com/air-verse/air) watches Go files and restarts on changes.

```toml
# backend/.air.api.toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o tmp/api ./cmd/api/"
bin = "tmp/api"
include_ext = ["go"]
exclude_dir = ["tmp", "migrations", "seeds"]
delay = 1000

[log]
time = false

[misc]
clean_on_exit = true
```

```toml
# backend/.air.worker.toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o tmp/worker ./cmd/worker/"
bin = "tmp/worker"
include_ext = ["go"]
exclude_dir = ["tmp", "migrations", "seeds"]
delay = 1000
```

### 6.4 Pre-Commit Hooks

Using [pre-commit](https://pre-commit.com/) framework:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      # Go: format
      - id: gofmt
        name: gofmt
        entry: gofmt -w
        language: system
        types: [go]

      # Go: vet
      - id: govet
        name: go vet
        entry: bash -c 'cd backend && go vet ./...'
        language: system
        pass_filenames: false
        files: '\.go$'

      # Go: lint (fast — only changed files)
      - id: golangci-lint
        name: golangci-lint
        entry: bash -c 'cd backend && golangci-lint run --new-from-rev=HEAD~1 ./...'
        language: system
        pass_filenames: false
        files: '\.go$'

      # Frontend: lint
      - id: eslint
        name: eslint
        entry: bash -c 'cd frontend && npx eslint --fix'
        language: system
        types: [ts, tsx]

      # Frontend: format
      - id: prettier
        name: prettier
        entry: bash -c 'cd frontend && npx prettier --write'
        language: system
        types_or: [ts, tsx, css, json]

      # Secrets: detect accidentally committed secrets
      - id: detect-secrets
        name: detect-secrets
        entry: detect-secrets-hook --baseline .secrets.baseline
        language: system
```

### 6.5 Go Linter Configuration

```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck        # unchecked errors
    - govet           # suspicious constructs
    - staticcheck     # comprehensive static analysis
    - unused          # unused code
    - gosimple        # simplifications
    - ineffassign     # useless assignments
    - misspell        # common misspellings
    - gofmt           # formatting
    - goimports       # import ordering
    - revive          # opinionated linter (replaces golint)
    - gocritic        # diagnostic + style + performance checks
    - exhaustive      # check switch exhaustiveness on enums
    - nilerr          # returning nil after checking err != nil

linters-settings:
  revive:
    rules:
      - name: exported
        arguments:
          - "checkPrivateReceivers"
      - name: unexported-return
        disabled: true  # we use unexported types in internal/

  exhaustive:
    default-signifies-exhaustive: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck   # test helpers often ignore errors intentionally
```

---

## 7. Testing Strategy

### 7.1 Backend Test Structure

```
backend/internal/
├── service/
│   ├── auth_service.go
│   └── auth_service_test.go      # Unit tests (mocked repos)
├── repository/
│   ├── user_repo.go
│   └── user_repo_test.go         # Integration tests (real DB via testcontainers)
├── handler/
│   ├── auth_handler.go
│   └── auth_handler_test.go      # HTTP tests (httptest + mocked services)
└── testutil/
    ├── factories.go              # Test data builders
    ├── assertions.go             # Custom test helpers
    └── db.go                     # Testcontainers Postgres setup
```

### 7.2 Test Categories

| Category | Package | Dependencies | Speed |
|----------|---------|-------------|-------|
| **Unit tests** | `service/*_test.go` | Mocked interfaces (repos, AI, payment) | Fast (~ms) |
| **HTTP tests** | `handler/*_test.go` | `httptest.Server` + mocked services | Fast (~ms) |
| **Integration tests** | `repository/*_test.go` | Real Postgres via testcontainers-go | Medium (~s) |
| **Integration tests** | `ai/*_test.go` | Real Redis (via testcontainers) for cache tests | Medium (~s) |

### 7.3 Mocking Strategy

Generate mocks from `domain/` interfaces using `go generate`:

```go
// domain/interfaces.go
//go:generate mockgen -source=interfaces.go -destination=../testutil/mocks/mocks.go -package=mocks
```

Alternatively, hand-written test doubles for simpler interfaces (e.g., `FileStore`).

### 7.4 Integration Tests — Testcontainers

```go
// testutil/db.go
package testutil

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    ctx := context.Background()

    pgContainer, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("careerdock_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    if err != nil {
        t.Fatalf("start postgres container: %v", err)
    }

    t.Cleanup(func() {
        pgContainer.Terminate(ctx)
    })

    connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        t.Fatalf("connect to test db: %v", err)
    }
    t.Cleanup(pool.Close)

    // Run migrations
    runMigrations(t, connStr)

    return pool
}
```

### 7.5 Test Factories

```go
// testutil/factories.go
package testutil

import (
    "time"
    "github.com/google/uuid"
    "github.com/skriptvalley/careerdock/internal/domain"
)

func NewUser(overrides ...func(*domain.User)) *domain.User {
    u := &domain.User{
        ID:           uuid.Must(uuid.NewV7()),
        Email:        "test@example.com",
        PasswordHash: "$2a$10$hashedpassword",
        Name:         "Test User",
        Role:         domain.RoleUser,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    for _, fn := range overrides {
        fn(u)
    }
    return u
}

func PremiumUser(u *domain.User) {
    now := time.Now()
    u.PremiumSince = &now
}

func AdminUser(u *domain.User) {
    u.Role = domain.RoleAdmin
}

// Usage: testutil.NewUser(testutil.PremiumUser, testutil.AdminUser)
```

### 7.6 Frontend Tests

```
frontend/src/
├── components/
│   ├── company/
│   │   ├── CompanyCard.tsx
│   │   └── CompanyCard.test.tsx       # Component unit test
├── hooks/
│   ├── use-auth.ts
│   └── use-auth.test.ts               # Hook unit test
└── __tests__/
    └── integration/
        └── login-flow.test.tsx         # Integration test
```

**Tools:** Jest + React Testing Library. E2E tests (Playwright) deferred to post-MVP.

---

## 8. CI Pipeline

### 8.1 PR Checks — `.github/workflows/ci.yml`

```yaml
# .github/workflows/ci.yml
name: CI

on:
  pull_request:
    branches: [main]

jobs:
  backend:
    name: Backend (lint + test + build)
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: careerdock_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U test"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache-dependency-path: backend/go.sum

      - name: Lint
        working-directory: backend
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
          golangci-lint run ./...

      - name: Test
        working-directory: backend
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/careerdock_test?sslmode=disable
          REDIS_URL: localhost:6379
        run: go test -race -count=1 -coverprofile=coverage.out ./...

      - name: Build
        working-directory: backend
        run: |
          CGO_ENABLED=0 go build -o bin/api ./cmd/api/
          CGO_ENABLED=0 go build -o bin/worker ./cmd/worker/

  frontend:
    name: Frontend (lint + test + build)
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install
        working-directory: frontend
        run: npm ci

      - name: Lint
        working-directory: frontend
        run: npm run lint

      - name: Test
        working-directory: frontend
        run: npm test -- --ci --coverage

      - name: Build
        working-directory: frontend
        run: npm run build
```

### 8.2 Deploy Workflow (Stub for Phase 6)

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    tags:
      - 'v*'  # Triggered by version tags

jobs:
  deploy:
    name: Build & Deploy
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      # Build Docker images, push to registry, deploy to EC2
      # Details in DEPLOYMENT.md (Phase 6)
      - run: echo "Deploy pipeline — see docs/DEPLOYMENT.md"
```

---

## 9. Migration Workflow

### 9.1 Tooling

- **Library:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Format:** `NNNNNN_description.up.sql` / `NNNNNN_description.down.sql`
- **Numbering:** Sequential (000001, 000002, ...) — simpler than timestamps for a solo project
- **Runner:** `cmd/migrate/main.go` wraps golang-migrate with the project's config

### 9.2 Creating a Migration

```bash
make migrate-new NAME=add_interview_rounds_table
# Creates:
#   backend/migrations/000017_add_interview_rounds_table.up.sql
#   backend/migrations/000017_add_interview_rounds_table.down.sql
```

The `gen-migration.sh` script scans existing migrations and creates the next sequential number.

### 9.3 Migration Rules

1. **Every migration must be reversible.** `down.sql` undoes `up.sql` exactly.
2. **One concern per migration.** Don't mix table creation with data backfill.
3. **No data-destructive changes without a plan.** Dropping a column? First: stop writing to it, then: deploy code that doesn't read it, then: drop it in the next migration.
4. **Test migrations locally.** `make migrate` → verify → `make migrate-down` → verify.
5. **Never edit a committed migration.** Create a new one instead.

### 9.4 Migration Entry Point

```go
// cmd/migrate/main.go
package main

import (
    "flag"
    "log"
    "os"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"

    "github.com/skriptvalley/careerdock/internal/config"
)

func main() {
    cfg := config.MustLoad()

    m, err := migrate.New("file://migrations", cfg.DatabaseURL)
    if err != nil {
        log.Fatal(err)
    }

    cmd := "up"
    if len(os.Args) > 1 {
        cmd = os.Args[1]
    }

    switch cmd {
    case "up":
        if err := m.Up(); err != nil && err != migrate.ErrNoChange {
            log.Fatal(err)
        }
        log.Println("migrations applied successfully")
    case "down":
        steps := 1
        if len(os.Args) > 2 {
            flag.IntVar(&steps, "", 1, "")
            steps, _ = parseSteps(os.Args[2])
        }
        if err := m.Steps(-steps); err != nil {
            log.Fatal(err)
        }
        log.Printf("rolled back %d migration(s)\n", steps)
    case "version":
        v, dirty, _ := m.Version()
        log.Printf("version: %d, dirty: %v\n", v, dirty)
    default:
        log.Fatalf("unknown command: %s (use: up, down, version)", cmd)
    }
}
```

---

## 10. Docker — Production Images

### 10.1 API Server

```dockerfile
# infra/docker/Dockerfile.api
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api/

# ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /api /usr/local/bin/api
COPY migrations/ /app/migrations/

EXPOSE 8080
ENTRYPOINT ["api"]
```

### 10.2 Worker

```dockerfile
# infra/docker/Dockerfile.worker
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker/

# ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /worker /usr/local/bin/worker

ENTRYPOINT ["worker"]
```

### 10.3 Image Tagging

Images are tagged with the git tag that triggered the build:

```
careerdock-api:v1.2.0
careerdock-worker:v1.2.0
```

This ensures you always know exactly which code is running in production. Details on registry, push, and deployment are in [DEPLOYMENT.md](./DEPLOYMENT.md) (Phase 6).

---

## 11. Coding Standards

### 11.1 Go Conventions

| Convention | Rule |
|-----------|------|
| Formatting | `gofmt` + `goimports` (enforced by linter) |
| Naming | stdlib conventions — `MixedCaps`, not `snake_case`. Acronyms uppercase: `ID`, `URL`, `HTTP` |
| Errors | Always check. Return `error` as last value. Wrap with `fmt.Errorf("...: %w", err)` for context |
| Panics | Never in handlers/services/repos. Only acceptable in `main()` for unrecoverable startup failures |
| Context | First parameter everywhere: `func (s *Service) Foo(ctx context.Context, ...)` |
| Interfaces | Accept interfaces, return structs. Define interfaces where they're consumed (`domain/`) |
| Packages | One package per directory. Package name = directory name. No `util` grab-bag packages |
| Comments | Exported functions: `// FunctionName does ...` format. Comment the "why", not the "what" |
| File size | Split files by entity/concept, not by layer. If a file exceeds ~300 lines, consider splitting |

### 11.2 Frontend Conventions

| Convention | Rule |
|-----------|------|
| Components | Functional only. No class components |
| Files | kebab-case (`company-card.tsx`). PascalCase exports (`CompanyCard`) |
| Types | Strict TypeScript. No `any` — use `unknown` + type guards |
| Imports | Absolute imports via `@/` alias (e.g., `@/components/ui/Button`) |
| Props | Explicit interface per component (no inline types for public components) |
| Side effects | In `useEffect` with proper deps. Data fetching in TanStack Query hooks |
| Error boundaries | Wrap each route group with an error boundary |
| Loading states | Skeleton components, not spinners (better perceived performance) |

### 11.3 SQL Conventions

| Convention | Rule |
|-----------|------|
| Keywords | UPPERCASE (`SELECT`, `WHERE`, `JOIN`) |
| Table names | snake_case, plural (`users`, `ats_checks`) |
| Column names | snake_case (`created_at`, `premium_since`) |
| Index names | `idx_{table}_{columns}` |
| Constraint names | `chk_{table}_{column}`, `uq_{table}_{columns}`, `fk_{table}_{ref}` |
| Parameters | Positional (`$1`, `$2`) via pgx — never string interpolation |

---

## 12. Architecture Decision Records (ADR)

Major architectural decisions are recorded in `docs/ADR/` to preserve context for future reference.

**Format:**

```markdown
# ADR NNN — Title

**Status:** Accepted | Superseded | Deprecated
**Date:** YYYY-MM-DD
**Context:** Why this decision was needed
**Decision:** What was decided
**Consequences:** Trade-offs and implications
```

**Initial ADRs to create during Sprint 0:**

| ADR | Topic |
|-----|-------|
| 001 | UUID v7 for primary keys |
| 002 | VARCHAR + CHECK over PostgreSQL enums |
| 003 | Trunk-based development with tag releases |
| 004 | Constructor injection over DI framework |
| 005 | Raw SQL (pgx) over ORM |
| 006 | Zustand + TanStack Query over Redux |

---

## 13. Quick Reference — Daily Workflow

```bash
# 1. Start infrastructure
make dev-infra

# 2. Start services (each in its own terminal)
make dev-api
make dev-worker
make dev-frontend

# 3. Open in browser
#    Frontend:  http://localhost:3000
#    API:       http://localhost:8080/api/health
#    MinIO:     http://localhost:9001
#    Mailhog:   http://localhost:8025

# 4. Create feature branch
git checkout -b feat/my-feature

# 5. Make changes (Air auto-reloads Go, Next.js HMR for frontend)

# 6. Run checks before committing
make lint
make test

# 7. Commit
git add -p
git commit -m "feat(scope): short description"

# 8. Push and create PR
git push -u origin feat/my-feature
gh pr create

# 9. After merge, clean up
git checkout main
git pull
git branch -d feat/my-feature
```

---

## Appendix A — File Naming Cheat Sheet

| Layer | Go file | Test file | Convention |
|-------|---------|-----------|------------|
| Domain entities | `entities.go` | — | Single file, all entities |
| Domain interfaces | `interfaces.go` | — | Single file, all interfaces |
| Handler | `{resource}.go` | `{resource}_test.go` | `company.go`, `resume.go` |
| Service | `{resource}_service.go` | `{resource}_service_test.go` | `auth_service.go` |
| Repository | `{resource}_repo.go` | `{resource}_repo_test.go` | `user_repo.go` |
| Middleware | `{name}.go` | `{name}_test.go` | `auth.go`, `rate_limit.go` |
| AI provider | `{provider}.go` | `{provider}_test.go` | `claude.go`, `openai.go` |
| Asynq tasks | `task_{name}.go` | `task_{name}_test.go` | `task_resume_parse.go` |
| Migrations | `NNNNNN_{desc}.up.sql` | — | `000001_initial_schema.up.sql` |

| Layer | Frontend file | Test file | Convention |
|-------|---------------|-----------|------------|
| Page | `page.tsx` | — | Next.js convention |
| Layout | `layout.tsx` | — | Next.js convention |
| Component | `ComponentName.tsx` | `ComponentName.test.tsx` | PascalCase |
| Hook | `use-{name}.ts` | `use-{name}.test.ts` | kebab-case |
| Store | `{name}-store.ts` | `{name}-store.test.ts` | kebab-case |
| Type | `{name}.ts` | — | kebab-case |
| Utility | `{name}.ts` | `{name}.test.ts` | kebab-case |
