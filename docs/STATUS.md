# Project Status

> **Last updated:** 2026-03-13

## Design Phases (Complete)
- Phase 1: ✅ Complete (PRD.md)
- Phase 2: ✅ Complete (ARCHITECTURE.md)
- Phase 3: ✅ Complete (LLD/database.md, api.md, payments.md, ai-service.md, frontend.md)
- Phase 4: ✅ Complete (CODE-STRUCTURE.md)
- Phase 5: ✅ Complete (SECURITY.md)
- Phase 6: ✅ Complete (DEPLOYMENT.md)
- Phase 7: ✅ Complete (ADMIN-PANEL.md)
- Phase 8: ✅ Complete (MONITORING.md)
- Phase 9: ✅ Complete (BUILD-PLAN.md)

## Implementation Sprints

- Sprint 0 (Foundation): ✅ Complete — PR #13 (merged)
- Sprint 1 (Company Directory): 🔄 Up next
- Sprint 2 (Lists & Tracking): ⬜ Not started
- Sprint 3 (Payments & Resume): ⬜ Not started
- Sprint 4 (AI Features): ⬜ Not started
- Sprint 5 (Admin & Polish): ⬜ Not started
- Sprint 6 (Launch Prep): ⬜ Not started

---

## Sprint 0 — Foundation (Tasks 0.1–0.31)

**Branch:** `feature/sprint-0-project-scaffold`
**PR:** #13
**CI:** ✅ All jobs passing (backend-lint, backend-test, backend-build, frontend-lint, frontend-build)

### Task Checklist

| # | Task | Status |
|:-:|------|:------:|
| 0.1 | Initialise Go module (`backend/`) | ✅ |
| 0.2 | Initialise Next.js project (`frontend/`) | ✅ |
| 0.3 | Docker Compose (Postgres, Redis, MinIO, Mailhog) | ✅ |
| 0.4 | Makefile (all targets) | ✅ |
| 0.5 | Air hot-reload config (`.air.api.toml`, `.air.worker.toml`) | ✅ |
| 0.6 | CI pipeline (`.github/workflows/ci.yml`) | ✅ |
| 0.7 | Pre-commit hooks (`.pre-commit-config.yaml`, `.golangci.yml`) | ✅ |
| 0.8 | `.env.example` + `.gitignore` | ✅ |
| 0.9 | PR template | ✅ |
| 0.10 | Backend config module (`internal/config/`) | ✅ |
| 0.11 | Domain layer (`internal/domain/` — entities, interfaces, errors, enums) | ✅ |
| 0.12 | Database migrations (18 tables across 16 migration files) | ✅ |
| 0.13 | Migration runner (`cmd/migrate/main.go`) | ✅ |
| 0.14 | Repository layer — user + transactor | ✅ |
| 0.15 | Auth service (register, login, refresh, logout, password reset) | ✅ |
| 0.16 | Auth middleware (JWT verify, role check, premium gate) | ✅ |
| 0.17 | Request ID + structured logging middleware | ✅ |
| 0.18 | CORS middleware | ✅ |
| 0.19 | Response helpers (`respondJSON`, `respondError`) | ✅ |
| 0.20 | Auth handlers (register, login, refresh, logout, me) | ✅ |
| 0.21 | Route mounting (`handler/routes.go`) | ✅ |
| 0.22 | API server entry point (`cmd/api/main.go` — DI wiring, graceful shutdown) | ✅ |
| 0.23 | Health check endpoints (`/api/health`, `/api/health/ready`) | ✅ |
| 0.24 | Worker entry point (`cmd/worker/main.go` — Asynq, 3 priority queues) | ✅ |
| 0.25 | Email service (Resend integration + templates) | ✅ |
| 0.26 | Frontend: App shell (providers, route group layouts) | ✅ |
| 0.27 | Frontend: Auth pages (login, register, forgot-password, reset-password, verify-email) | ✅ |
| 0.28 | Frontend: API client (`lib/api.ts`, `lib/query-keys.ts`) | ✅ |
| 0.29 | Frontend: Auth store + hook (`store/auth-store.ts`, `hooks/use-auth.ts`) | ✅ |
| 0.30 | Seed script skeleton (`cmd/seed/main.go`) | ✅ |
| 0.31 | Initial ADRs (`docs/ADR/001-006`) | ✅ |

### Definition of Done

| Criterion | Status | Notes |
|-----------|:------:|-------|
| `make dev` starts Postgres, Redis, MinIO, Mailhog | ✅ | `docker-compose.yml` with all 4 services + healthchecks |
| `make dev-api` starts API with hot reload, `GET /api/health` returns 200 | ✅ | `.air.api.toml` + `cmd/api/main.go` + `handler/health.go` |
| `make dev-worker` starts worker (no tasks yet) | ✅ | `.air.worker.toml` + `cmd/worker/main.go` (Asynq placeholder) |
| `make dev-frontend` starts Next.js on :3000 | ✅ | Next.js 15 App Router with Tailwind |
| User can register, verify email, login, refresh token, logout | ✅ | Full auth flow: service + handlers + middleware + frontend pages |
| CI pipeline passes (lint + test + build) on PR | ✅ | 5 jobs: backend-lint, backend-test, backend-build, frontend-lint, frontend-build |
| `make migrate` applies all 18 tables | ✅ | 16 migration files, 18 `CREATE TABLE` statements |
| `make migrate-down` rolls back cleanly | ✅ | All 16 `.down.sql` files present |

---

## Sprint 1 — Company Directory (Tasks 1.1–1.16)

**Branch:** `feature/sprint-1-company-directory`
**PR:** TBD
**CI:** ⬜ Not started
**Est. hours:** ~70

### Task Checklist

| # | Task | Status |
|:-:|------|:------:|
| 1.1 | Company repository (`repository/company_repo.go` — CRUD, FTS, cursor pagination) | ⬜ |
| 1.2 | Company service (`service/company_service.go` — list, search, getBySlug, filter) | ⬜ |
| 1.3 | Company handlers (`handler/company.go` — public endpoints) | ⬜ |
| 1.4 | Company routes (mount in `routes.go`) | ⬜ |
| 1.5 | Seed data file (`seeds/companies.json` — 50-100 Indian tech companies) | ⬜ |
| 1.6 | Seed runner (upgrade `cmd/seed/main.go` — typed parsing, upsert via repo) | ⬜ |
| 1.7 | Frontend: Company list page (`app/(public)/companies/page.tsx` — grid/list, search, filters, infinite scroll) | ⬜ |
| 1.8 | Frontend: Company profile page (`app/(public)/companies/[slug]/page.tsx` — SSR, SEO meta) | ⬜ |
| 1.9 | Frontend: Company components (`CompanyCard`, `CompanyFilters`, `CompanySearchBar`, `TechStackTags`) | ⬜ |
| 1.10 | Frontend: Service Worker + IndexedDB (offline caching for company directory) | ⬜ |
| 1.11 | S3/MinIO integration (`internal/storage/` — upload, download, signed URL, delete) | ⬜ |
| 1.12 | Company logo upload (admin — logo stored in S3 logos bucket) | ⬜ |
| 1.13 | Landing page (`app/page.tsx` — hero, features, CTA) | ⬜ |
| 1.14 | Pricing page (`app/(public)/pricing/page.tsx` — Starter Pack, à la carte) | ⬜ |
| 1.15 | Frontend: Header + Footer (`components/layout/Header.tsx`, `Footer.tsx`) | ⬜ |
| 1.16 | ETags / cache headers (company list + profile for CDN) | ⬜ |

### Definition of Done

| Criterion | Status |
|-----------|:------:|
| `GET /api/companies` returns paginated company list | ⬜ |
| `GET /api/companies/search?q=tata` returns FTS results | ⬜ |
| `GET /api/companies/{slug}` returns company profile | ⬜ |
| `make seed` populates 50+ companies | ⬜ |
| Company list page renders with search, filter, pagination | ⬜ |
| Company profile page is SSR with correct meta tags (OG, Twitter) | ⬜ |
| Offline: previously viewed companies available without network | ⬜ |
