# Project Status

> **Last updated:** 2026-03-23

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
- Sprint 1 (Company Directory): ✅ Complete — PR #15 (merged), CI ✅
- Sprint 2 (Lists & Tracking): ✅ Complete — PR #17 (merged), PR #18 (merged)
- Sprint 3 (Payments & Resume): 🔨 In progress (tasks 3.1–3.23 complete)
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
**PR:** #15 (merged 2026-03-13)
**CI:** ✅ All jobs passing (backend-lint, backend-test, backend-build, frontend-lint, frontend-build)
**Est. hours:** ~70

### Task Checklist

| # | Task | Status |
|:-:|------|:------:|
| 1.1 | Company repository (`repository/company_repo.go` — CRUD, FTS, cursor pagination) | ✅ |
| 1.2 | Company service (`service/company_service.go` — list, search, getBySlug, filter) | ✅ |
| 1.3 | Company handlers (`handler/company.go` — public endpoints) | ✅ |
| 1.4 | Company routes (mount in `routes.go`) | ✅ |
| 1.5 | Seed data file (`seeds/companies.json` — 50-100 Indian tech companies) | ✅ |
| 1.6 | Seed runner (upgrade `cmd/seed/main.go` — typed parsing, upsert via repo) | ✅ |
| 1.7 | Frontend: Company list page (`app/(public)/companies/page.tsx` — grid/list, search, filters, infinite scroll) | ✅ |
| 1.8 | Frontend: Company profile page (`app/(public)/companies/[slug]/page.tsx` — SSR, SEO meta) | ✅ |
| 1.9 | Frontend: Company components (`CompanyCard`, `CompanyFilters`, `CompanySearchBar`, `TechStackTags`) | ✅ |
| 1.10 | Frontend: Service Worker + IndexedDB (offline caching for company directory) | ✅ |
| 1.11 | S3/MinIO integration (`internal/storage/` — upload, download, signed URL, delete) | ✅ |
| 1.12 | Company logo upload (admin — logo stored in S3 logos bucket) | ⏭️ Deferred to Sprint 5 (Admin) |
| 1.13 | Landing page (`app/page.tsx` — hero, features, CTA) | ✅ |
| 1.14 | Pricing page (`app/(public)/pricing/page.tsx` — Starter Pack, à la carte) | ✅ |
| 1.15 | Frontend: Header + Footer (`components/layout/Header.tsx`, `Footer.tsx`) | ✅ |
| 1.16 | ETags / cache headers (company list + profile for CDN) | ✅ |

### Definition of Done

| Criterion | Status |
|-----------|:------:|
| `GET /api/companies` returns paginated company list | ✅ |
| `GET /api/companies/search?q=tata` returns FTS results | ✅ |
| `GET /api/companies/{slug}` returns company profile | ✅ |
| `make seed` populates 50+ companies | ✅ |
| Company list page renders with search, filter, pagination | ✅ |
| Company profile page is SSR with correct meta tags (OG, Twitter) | ⏭️ Deferred — needs `generateMetadata` in App Router |
| Offline: previously viewed companies available without network | ✅ |

---

## Sprint 2 — User Lists & Tracking (Tasks 2.1–2.15)

**Branch:** `feature/sprint-2-lists-tracking`
**PR:** #17 (merged 2026-03-14), #18 (merged 2026-03-14)
**CI:** ✅ All jobs passing
**Est. hours:** ~63

### Task Checklist

| # | Task | Status |
|:-:|------|:------:|
| 2.1 | List repository (`repository/list_repo.go` — CRUD, entries, status history, interview rounds) | ✅ |
| 2.2 | List entry repository (CRUD for entries with application status, notes, dates) | ✅ |
| 2.3 | List service (`service/list_service.go` — enforce 3-list free / 5-list premium limit) | ✅ |
| 2.4 | List handlers (`handler/list.go` — all list + entry + round endpoints + dashboard) | ✅ |
| 2.5 | List routes (mount authenticated routes in `routes.go`, wire DI in `main.go`) | ✅ |
| 2.6 | Frontend: Lists page (`app/(dashboard)/lists/page.tsx` — list cards, create modal) | ✅ |
| 2.7 | Frontend: List detail page (`app/(dashboard)/lists/[id]/page.tsx` — entry table, add company, status tracking) | ✅ |
| 2.8 | Frontend: Application tracker (status pipeline with inline status editing) | ✅ |
| 2.9 | Frontend: Dashboard (free) (`app/(dashboard)/dashboard/page.tsx` — funnel view, list overview) | ✅ |
| 2.10 | Frontend: Dashboard layout (`app/(dashboard)/layout.tsx` — sidebar nav, top bar, mobile nav) | ✅ |
| 2.11 | Frontend: Settings page (`app/(dashboard)/settings/page.tsx` — profile edit, password change, account deletion) | ✅ |
| 2.12 | User service (`service/user_service.go` — profile update, password change, soft delete) | ✅ |
| 2.13 | User handlers (`handler/user.go` — PUT /users/me, PUT /users/me/password, DELETE /users/me) | ✅ |
| 2.14 | Notification model skeleton (`repository/notification_repo.go` — create, list, mark read, count unread) | ✅ |
| 2.15 | SSE endpoint skeleton (`GET /api/events` — authenticated, heartbeat, no events yet) | ✅ |

### Definition of Done

| Criterion | Status |
|-----------|:------:|
| User can create up to 3 lists (free tier) | ✅ |
| User can add companies to lists with application status | ✅ |
| Status transitions work (not_applied → applied → ... → accepted/rejected) | ✅ |
| Dashboard shows funnel view with correct counts | ✅ |
| Settings: profile edit, password change, account deletion work | ✅ |
| SSE: endpoint accepts connections (no events emitted yet) | ✅ |

### Additional Work (Feedback Sessions 02–04)

- 4 feedback sessions implemented across UI polish, UX improvements, and bug fixes
- Key additions: sidebar nav, quick-add-to-list modal, company browser panel, unsaved changes guard, fuzzy search, office mode filter, applications company filter
- Bug fixes: delete account request body, company detail page application filtering
- See `ai/feedback/FEEDBACK-STATUS.md` for full details

---

## Sprint 3 — Payments & Resume Foundation (Tasks 3.1–3.23)

**Branch 1:** `feature/sprint-3-payments` — PR #19 (merged)
**Branch 2:** `feature/sprint-3-resume-and-ai` — PR #20 (merged)
**Branch 3:** `feature/sprint-3-frontend` — PR TBD
**CI:** ✅ All jobs passing
**Est. hours:** ~95

### Task Checklist

| # | Task | Status |
|:-:|------|:------:|
| 3.1 | Payment repository (`repository/payment_repo.go` — create order, update status, list) | ✅ |
| 3.2 | Credit repository (`repository/credit_repo.go` — allocate, deduct, balance, log) | ✅ |
| 3.3 | Payment service (`service/payment_service.go` — Razorpay order, webhook, refund) | ✅ |
| 3.4 | Razorpay adapter (`payment/razorpay.go` — create order, verify signature, refund) | ✅ |
| 3.5 | Payment handlers (`handler/payment.go` — create order, webhook, history) | ✅ |
| 3.6 | Webhook signature verification (HMAC-SHA256 for `/api/webhooks/razorpay`) | ✅ |
| 3.7 | Credit service (`service/credit_service.go` — balance, deduction, premium gating) | ✅ |
| 3.8 | Premium middleware (`auth.RequirePremium` — checks `premium_since`) | ✅ |
| 3.9 | Resume repository (`repository/resume_repo.go` — CRUD, list by user, slot lookup) | ✅ |
| 3.10 | Resume service (`service/resume_service.go` — upload, validate, S3 store, credit deduct) | ✅ |
| 3.11 | Resume handlers (`handler/resume.go` — upload multipart, list, get, set default, archive, download URL) | ✅ |
| 3.12 | PDF extraction (`pdf/extractor.go` — extract text from PDF using pdfcpu) | ✅ |
| 3.13 | AI provider interface (`ai/provider.go` — LLMProvider interface + all types) | ✅ |
| 3.14 | Claude provider (`ai/claude.go` — Messages API, callWithPDF, response parsing) | ✅ |
| 3.15 | OpenAI provider (`ai/openai.go` — Chat Completions API, text-only calls) | ✅ |
| 3.16 | Fallback provider (`ai/fallback.go` — Claude → OpenAI fallback) | ✅ |
| 3.17 | Prompt templates (`ai/prompts/` — security preamble, resume parse, ATS general) | ✅ |
| 3.18 | AI result cache (`ai/cache.go` — Redis GET/SET with SHA256 keys, per-op TTL) | ✅ |
| 3.19 | Worker: resume parse+score task (`worker/task_resume_parse.go` — parse + general ATS) | ✅ |
| 3.20 | Worker: email send task (`worker/task_email_send.go` — Resend integration) | ✅ |
| 3.21 | Frontend: Pricing + checkout (Razorpay Checkout.js, order, confirmation) | ✅ |
| 3.22 | Frontend: Resume management (`app/(dashboard)/resumes/page.tsx`) | ✅ |
| 3.23 | Frontend: Credit balance display (dashboard sidebar/header) | ✅ |

### Implementation Notes (Tasks 3.1–3.8, 3.9–3.20)

- **Product catalog** in code: starter_pack (₹399), resume_upload (₹49), ats_bundle (₹99), rebuy_pack (₹399)
- **Idempotent webhooks**: duplicate `payment.captured` events detected and safely ignored
- **Atomic credit allocation**: payment capture + credits + audit log + premium_since in single DB transaction
- **Business rules**: starter_pack only for non-premium users; rebuy_pack only for premium users
- **Endpoints added**: `POST /api/payments/orders`, `GET /api/payments`, `POST /api/webhooks/razorpay`, `GET /api/credits`, `GET /api/credits/transactions`
- **Resume endpoints**: `POST /api/resumes` (multipart upload), `GET /api/resumes`, `GET /api/resumes/:id`, `PUT /api/resumes/:id/default`, `DELETE /api/resumes/:id`, `GET /api/resumes/:id/download`
- **PDF extraction** via pdfcpu (pure Go); text stored in DB, raw PDF sent to Claude for ATS scoring
- **AI providers**: Claude (primary, supports native PDF), OpenAI (fallback, text-only), FallbackProvider wraps both
- **Prompt templates**: anti-injection preamble, XML-delimited user content, resume parse + ATS general system prompts
- **AI result cache**: Redis-backed, SHA256 cache keys, per-operation TTL (24h for parse/ATS general)
- **Worker tasks**: `resume:parse_and_score` (parse + general ATS, cache-aware), `email:send` (Resend integration)
- **DI wiring**: Asynq client in API server, full AI provider chain in worker, S3 resume store with auto-bucket in dev

### Implementation Notes (Tasks 3.21–3.23)

- **Razorpay Checkout.js**: Dynamic script loader (`lib/razorpay.ts`) with TypeScript declarations, no npm package needed
- **Pricing page**: Upgraded from static to interactive — authenticated users see buy buttons, Razorpay modal opens on click, payment confirmation updates credits + premium status
- **Product flow**: Non-premium users buy Starter Pack; premium users can buy a la carte credits (resume_upload, ats_bundle, rebuy_pack)
- **Resume management page**: Slot-based UI (3 slots), drag-and-drop PDF upload, status badges (ready/processing/failed), ATS score + parsed skills display, set default / archive / download actions
- **Credit balance**: Sidebar widget shows total + per-type breakdown for premium users; collapsed mode shows compact badge
- **Sidebar**: Added "Resumes" nav link (FileText icon) between Applications and Companies
- **Hooks**: `use-payments.ts` (credit balance, transactions, order creation, payment confirmation), `use-resumes.ts` (list, upload, set default, archive, download URL)
- **API types**: Added PaymentOrder, CreditBalances, CreditTransaction, ResumeListItem, ResumeDetail, ParsedSummary types
- **SSE real-time updates**: Redis pub/sub backend — worker publishes `resume_ready` events, SSE handler subscribes per-user and forwards to connected clients; frontend `useSSE` hook auto-invalidates TanStack Query caches on events

### Definition of Done

| Criterion | Status |
|-----------|:------:|
| User can purchase Starter Pack via Razorpay (test mode) | ✅ |
| Webhook correctly allocates credits and sets `premium_since` | ✅ |
| À la carte purchases work (resume upload credit, ATS bundle) | ✅ |
| Credit balance shown in UI, deducted on premium actions | ✅ |
| User can upload PDF resume (validated, stored in S3/MinIO) | ✅ |
| Resume parse + general ATS runs async, results stored in DB | ✅ |
| SSE notifies user when resume processing completes | ✅ |
| AI fallback: if Claude fails, OpenAI is used | ✅ |
| AI result cache: repeated requests return cached result | ✅ |
