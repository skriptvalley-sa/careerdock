# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**CareerDock** is a SaaS career intelligence platform for tech job seekers in India. It's a solo founder project being built design-first, then implemented incrementally.

**Current state:** All 9 design phases complete. Implementation begins with Sprint 0. See `docs/ai/SESSION-GUIDE.md` for session and branch management.

## Working Agreements

1. **Discussion-first**: For each design phase, present recommendations with trade-offs and wait for approval before proceeding.
2. **Document everything**: Create/update markdown files in `docs/` for every finalised decision.
3. **Incremental delivery**: Design for MVP-first. Flag features that can be deferred to v2.
4. **Cost-conscious**: Bootstrapped project — optimize for low operational cost.
5. **India-first context**: Default currency is INR. Target market is India. Pricing in ₹.

## Development Phases

The project follows 9 design phases before implementation:

| Phase | Output |
|-------|--------|
| 1 | `docs/PRD.md` |
| 2 | `docs/ARCHITECTURE.md` |
| 3 | `docs/LLD/` (database, api, ai-service, frontend, payments) |
| 4 | `docs/CODE-STRUCTURE.md` |
| 5 | `docs/SECURITY.md` |
| 6 | `docs/DEPLOYMENT.md` |
| 7 | `docs/ADMIN-PANEL.md` |
| 8 | `docs/MONITORING.md` |
| 9 | `docs/BUILD-PLAN.md` |

After all phases complete, implementation follows 6 sprints (Sprint 0 = Foundation, Sprint 5 = Admin & Polish). See `docs/CLAUDE-PROMPT.md` for full sprint breakdown.

## Planned Tech Stack

### Backend
- **Language:** Go
- **HTTP Framework:** Chi or Echo (TBD during architecture phase)
- **Database:** PostgreSQL (primary), Redis (caching/sessions)
- **Search:** MeiliSearch or Typesense
- **File Storage:** AWS S3 (MinIO for local dev)
- **Job Queue:** Asynq (Redis-backed) or AWS SQS
- **API Style:** REST (GraphQL considered for complex queries)

### Frontend
- **Framework:** Next.js (App Router, SSR for SEO on public company pages)
- **Offline support:** Service Worker + IndexedDB for company directory browsing

### Infrastructure
- **Hosting:** AWS ECS Fargate (Fly.io/Railway for MVP)
- **Database hosting:** AWS RDS or Supabase
- **CDN/DNS:** CloudFront + Route 53
- **CI/CD:** GitHub Actions

### Third-Party Services
- **LLM:** Claude API (primary), OpenAI (fallback) — abstract behind an interface
- **Payments:** Razorpay
- **Email:** Resend or AWS SES
- **Monitoring:** Prometheus + Grafana
- **Error tracking:** Sentry

## Planned Repository Structure

Once implementation begins (Sprint 0), the repository will be structured as:

```
careerdock/
├── docs/                    # All design documents
├── backend/
│   ├── cmd/                 # Entry points: api, worker, migrate, seed
│   ├── internal/
│   │   ├── domain/          # Business entities and interfaces
│   │   ├── handler/         # HTTP handlers
│   │   ├── service/         # Business logic
│   │   ├── repository/      # Database access layer
│   │   ├── ai/              # LLM provider abstraction
│   │   ├── payment/         # Razorpay integration
│   │   ├── middleware/      # Auth, rate limiting, logging
│   │   └── config/          # Configuration management
│   ├── migrations/          # SQL migrations
│   └── seeds/               # Seed data (companies)
├── frontend/
│   └── src/
│       ├── pages/
│       ├── components/
│       ├── hooks/
│       ├── services/        # API client
│       └── store/           # State management
├── infra/
│   ├── terraform/
│   └── docker/
├── scripts/
├── docker-compose.yml       # Local dev: Postgres, Redis, MinIO, Mailhog, MeiliSearch
└── Makefile                 # dev, test, lint, migrate, seed, build
```

## Key Architectural Decisions (Pending Finalization)

- **Go backend layering:** handler → service → repository (no direct DB access from handlers)
- **Auth:** JWT + refresh tokens in httpOnly cookies only (no localStorage)
- **LLM abstraction:** Single interface supporting Claude and OpenAI backends
- **Async AI operations:** Resume analysis and ATS scoring run via job queue, not synchronously
- **AI result caching:** Results are deterministic — cache aggressively to reduce LLM costs
- **Feature flags:** Database-backed, admin-controllable, no redeploy needed

## Core Product Features

1. **Public company directory** (~200-500 Indian tech companies) — browsable without auth, SEO-optimized
2. **Free tier:** Up to 3 custom company lists + application tracking (status, dates, notes)
3. **Paid tier (one-time "Starter Pack", ~₹299-499):**
   - Upload up to 3 resumes (PDF) with AI extraction
   - AI-curated company matching based on resume profile
   - ATS scoring: General score, company-specific score, job-specific score
   - Additional credits purchasable à la carte

## Commands

```bash
make dev        # Start local dev environment (Docker Compose)
make test       # Run all tests
make lint       # Run all linters (backend + frontend)
make lint-backend   # Run Go linter only (golangci-lint v2)
make lint-frontend  # Run frontend linter only (ESLint)
make migrate    # Run database migrations
make seed       # Seed company data
make build      # Build backend + frontend
```

For Go tests:
```bash
cd backend && go test ./...                      # All tests
cd backend && go test ./internal/service/...     # Specific package
cd backend && go test -run TestFunctionName ./...  # Single test
```

## Pre-Commit Checks

**IMPORTANT:** Always run `make lint` locally before committing Go or frontend changes. This catches gofmt, revive, errcheck, and ESLint issues before they hit CI.

```bash
# Run before every commit that touches backend or frontend code:
make lint

# If only backend changed:
make lint-backend

# If only frontend changed:
make lint-frontend
```

The CI pipeline runs 5 checks: backend-lint, backend-test, backend-build, frontend-lint, frontend-build. Running `make lint` locally mirrors the lint checks. Running `make build` mirrors the build checks.

**Requirements:**
- `golangci-lint` v2 (installed via Homebrew: `brew install golangci-lint`)
- Node.js + npm (for frontend ESLint)

## Reference Documents

- `docs/CLAUDE-PROMPT.md` — Full master prompt with all requirements, open questions, and constraints
- `docs/USAGE-GUIDE.md` — How to structure Claude Code sessions for this project
- `ai/SESSION-GUIDE.md` — Session, branch, and PR workflow for the build phase
- `docs/STATUS.md` — Track which phases and sprints are complete
