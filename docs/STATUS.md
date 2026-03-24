# Project Status

> **Last updated:** 2026-03-25 — Sprint 4 complete ✅

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
- Sprint 3 (Payments & Resume): ✅ Complete — PR #19, #20, #21, #22, #23, #24, #25, #26 (all merged)
- Sprint 4 (AI Features): ✅ Complete — PR #28, #29, #30, #31 (all merged)
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
**Branch 3:** `feature/sprint-3-frontend` — PR #21 (merged)
**SSE & Hotfixes:** PR #22 (SSE wiring), #23 (Flusher unwrap), #24 (reconnect backoff), #25 (logger Unwrap), #26 (WriteTimeout fix) — all merged
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

### Post-Sprint Hotfixes (PRs #22–#26)

- **PR #22**: Wired SSE real-time updates — worker publishes `resume_ready` via Redis pub/sub, frontend `useSSE` hook invalidates queries
- **PR #23**: Fixed SSE 500 — added `getFlusher()` with unwrapper interface to reach `http.Flusher` through Chi middleware
- **PR #24**: Fixed SSE reconnect — replaced one-shot error handler with exponential backoff (1s → 30s cap)
- **PR #25**: Added `Unwrap()` to logger middleware's `statusWriter` so `getFlusher()` and `ResponseController` can reach the base writer
- **PR #26**: Used `http.ResponseController.SetWriteDeadline()` to extend WriteTimeout before each SSE write, preventing 30s connection drops

---

## Sprint 4 — AI Features (Tasks 4.1–4.17)

**Branch (ATS backend):** `feature/sprint-4-ats-backend` — **PR #28 (merged)**
**Branch (Curated lists):** `feature/sprint-4-curated-lists` — **PR #29 (merged)**
**Branch (Frontend):** `feature/sprint-4-frontend` — **PR #30 (merged)**
**Branch (Scheduler):** `feature/sprint-4-scheduler` — **PR #31 (merged)**
**CI:** ✅ All jobs passing
**Est. hours:** ~90

### Task Checklist

| # | Task | Status |
|:-:|------|:------:|
| 4.1 | ATS check repository (`repository/ats_repo.go` — create, get, list by user/resume) | ✅ |
| 4.2 | ATS service (`service/ats_service.go` — company check, job check, credit deduction) | ✅ |
| 4.3 | ATS handlers (`handler/ats.go` — POST /ats/company, POST /ats/job, GET /ats/:id, GET /ats/) | ✅ |
| 4.4 | Worker: company ATS task (`worker/task_ats_company.go` — download PDF, score with AI, store result) | ✅ |
| 4.5 | Worker: job ATS task (`worker/task_ats_job.go` — download PDF, score with AI + JD, store result) | ✅ |
| 4.6 | Curated list repository (`repository/curated_list_repo.go` — create, get, list) | ✅ |
| 4.7 | Curated list service (`service/curated_list_service.go` — trigger generation, credit deduction) | ✅ |
| 4.8 | Curated list handler (`handler/curated_list.go` — POST /curated-lists, GET /curated-lists/:id) | ✅ |
| 4.9 | Worker: curate company list task (`worker/task_curate_list.go` — build profile, query companies, AI curation) | ✅ |
| 4.10 | AI prompt templates — Company ATS, Job ATS, Curated List prompts | ✅ |
| 4.11 | Output validation (`ai/validation.go` — schema validation for AI responses, score bounds, retry) | ✅ |
| 4.12 | Frontend: ATS check page (`app/(dashboard)/ats/page.tsx` — select resume, choose company/paste JD) | ✅ |
| 4.13 | Frontend: ATS result page (`app/(dashboard)/ats/[id]/page.tsx` — score display, breakdown, recommendations) | ✅ |
| 4.14 | Frontend: Curated lists page (`app/(dashboard)/curated-lists/page.tsx` — generate new, view results) | ✅ |
| 4.15 | Frontend: Premium dashboard (resume health, credit tracker, recent ATS scores, quick actions) | ✅ |
| 4.16 | Frontend: SSE integration for ATS/curated list completion events | ✅ |
| 4.17 | Asynq scheduler setup (periodic tasks — user hard-delete cleanup) | ✅ |

### Implementation Notes (Tasks 4.1–4.11)

- **ATS DB dedup**: `ats_checks.cache_key = sha256(resumeID + ":" + companyID)` prevents double-charging for identical requests
- **Curated list DB dedup**: `curated_lists.preferences_hash = sha256(resumeID)` — per-resume dedup with cache miss check before credit deduction
- **Two-level caching**: DB-level dedup (no re-charge) + Redis AI cache (7-day TTL for company/curated, 24h for job) prevents redundant LLM calls
- **Async result pattern**: Workers create records with `result = '{}'` placeholder; frontend detects pending by checking `result == {}`; workers call `UpdateResult` on completion
- **SSE events published**: `ats_company_complete`, `ats_job_complete`, `curated_list_complete` — all sent to `sse:user:{userID}` Redis channel
- **AI prompt categories**:
  - Company ATS: `tech_stack_match`, `domain_fit`, `seniority_fit`, `keyword_density`, `impact_metrics`
  - Job ATS: `required_skills`, `preferred_skills`, `experience_level`, `domain_relevance`, `keyword_density`
- **Curated list**: AI selects top-20 companies from `CompanyRepository.ListAll()` (capped at 500); returns ranked list with `match_score`, `match_reasons`, `recommendation`
- **Output validation + retry**: `ValidateATSResultRetry` / `ValidateCuratedListResultRetry` retry up to 2 times on schema/bounds violations before returning error

### Implementation Notes (Tasks 4.12–4.17)

- **ATS check page** (`/ats`): Resume selector (ready resumes only), company/JD mode toggle, `CompanyCombobox` for company mode, JD textarea with 100–10,000 char counter; submits → navigates to `/ats/{id}`
- **ATS result page** (`/ats/[id]`): Score ring (color-coded: ≥80 neon green, ≥60 amber, <60 red), match label, breakdown grid with progress bars + feedback text, suggestions list; polls via `useATSCheck` every 5s while pending; SSE invalidates cache on completion
- **Curated lists page** (`/curated-lists`): Inline generate form with resume selector; `CuratedListCard` shows 5 companies collapsed / expand-all; `PendingListPoller` polls `useCuratedList(id)` every 8s while pending
- **Premium dashboard section**: Only rendered for `user.premium_since != null`; shows resume health card (default resume + ATS score), credit tracker (resume_upload/ats_check/curated_list), quick action links, recent ATS checks mini-list (last 3)
- **SSE hooks** (`use-ats.ts`, `use-curated-lists.ts`): `isATSComplete(result)` and `isCuratedListComplete(result)` are TypeScript type guards checking for `score`/`companies` key presence; `use-sse.ts` invalidates query caches on `ats_company_complete`, `ats_job_complete`, `curated_list_complete`
- **Scheduler** (task 4.17): `asynq.Scheduler` in `cmd/worker/main.go`; cron `0 2 * * *` enqueues `admin:user_cleanup`; handler calls `UserRepo.HardDeleteExpired(ctx, now-30d)`; scheduler shuts down alongside the task server on SIGTERM

### End-to-End Test Results (VPS — 2026-03-25)

| Test | Result |
|------|--------|
| Upload resume (slot 1) | ✅ — file stored in MinIO |
| Worker parse: PDF with no text | ✅ — correctly marked `status=failed` |
| ATS company check (Go resume vs company) | ✅ — score=62, 5 breakdown categories, 5 suggestions, cached in Redis 7d |
| ATS job check (Go resume vs JD) | ✅ — score=78, 5 breakdown categories, 5 suggestions, cached 24h |
| Curated list generation | ✅ — 20 companies ranked (Google 95%, Uber 94%, Cloudflare 93%, Razorpay 92%, PhonePe 91%) |
| SSE `ats_company_complete` | ✅ — `{check_id, resume_id, score: 72}` received live |
| SSE `curated_list_complete` | ✅ — `{list_id, resume_id, companies_ranked: 20}` received live |
| Heartbeat | ✅ — `: heartbeat` every 25s |

### Definition of Done

| Criterion | Status |
|-----------|:------:|
| Company ATS check: user selects resume + company, gets score with breakdown | ✅ |
| Job ATS check: user selects resume + pastes JD, gets score with breakdown | ✅ |
| Curated list: AI-curated company list based on resume profile | ✅ |
| All AI operations async — loading state → SSE notification on completion | ✅ |
| Credit deduction works correctly for each operation | ✅ |
| Results are cached — repeat requests return instantly | ✅ |
| AI fallback works end-to-end (Claude → OpenAI) | ✅ |
| Premium dashboard shows resume health, credits, recent activity | ✅ |
| Scheduler: daily user hard-delete cleanup at 02:00 UTC | ✅ |
| VPS deployed + full end-to-end test passing | ✅ |
| SSE events verified: `ats_company_complete`, `ats_job_complete`, `curated_list_complete` | ✅ |

---

## Sprint 5 — Admin & Polish (Tasks 5.1–5.x)

**Branch:** `feature/sprint-5-admin-polish` *(not started)*
**Status:** ⬜ Not started

### Planned Tasks

| # | Task | Status |
|:-:|------|:------:|
| 5.1 | Admin panel — company CRUD (create, edit, logo upload) | ⬜ |
| 5.2 | Admin panel — user management (list, ban, grant premium) | ⬜ |
| 5.3 | Admin panel — credit management (manual allocation) | ⬜ |
| 5.4 | Admin panel — payment & transaction logs | ⬜ |
| 5.5 | Company enrichment worker (`admin:company_enrich` — auto-populate tech stack from web) | ⬜ |
| 5.6 | Company refresh worker (`admin:company_refresh` — periodic re-enrichment) | ⬜ |
| 5.7 | SSR + `generateMetadata` for company profile page (OG tags, Twitter card) | ⬜ |
| 5.8 | Rate limiting middleware (per-IP + per-user) | ⬜ |
| 5.9 | Notification system (in-app notifications for events) | ⬜ |
| 5.10 | UI polish — loading skeletons, error boundaries, empty states | ⬜ |
| 5.11 | Onboarding flow — guided setup for new users | ⬜ |
| 5.12 | Mobile nav improvements — bottom tab bar | ⬜ |
