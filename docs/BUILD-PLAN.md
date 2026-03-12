# CareerDock — Build Plan

> **Version:** 1.0
> **Status:** Draft (Phase 9)
> **Last updated:** 2026-03-12
> **Depends on:** All preceding design documents

---

## 1. Overview

This is the final design document. It maps every feature and infrastructure component from Phases 1-8 into implementation sprints. After this document is approved, implementation begins.

**Timeline:** 7 sprints over ~14 weeks (solo developer).

**"Done" criteria for launch:**
1. Users can browse companies, sign up, create lists, track applications (free tier).
2. Users can pay, upload resumes, get ATS scores, see AI-curated suggestions (premium tier).
3. Admin can manage companies, users, features, and see financials.
4. System deployed, monitored, handling ~1,000 concurrent users.
5. All design documents complete and up-to-date in `docs/`.

---

## 2. Sprint Overview

| Sprint | Theme | Duration | Deliverables |
|:---:|-------|----------|-------------|
| **0** | Foundation | 2 weeks | Repo, CI, Docker, DB, auth, skeletons |
| **1** | Company Directory | 2 weeks | Company APIs, seed data, public pages, search |
| **2** | User Lists & Tracking | 2 weeks | List CRUD, application tracking, free dashboard |
| **3** | Payments & Resume | 2 weeks | Razorpay, credits, resume upload, AI foundation |
| **4** | AI Features | 2 weeks | ATS scoring (3 tiers), curated lists, premium dashboard |
| **5** | Admin & Polish | 2 weeks | Admin panel, feature flags, monitoring, security audit |
| **6** | Launch Prep | 2 weeks | Load testing, bug fixes, production deploy, beta |

### 2.1 Dependency Graph

```
Sprint 0 (Foundation)
    │
    ├── Sprint 1 (Company Directory)
    │       │
    │       ├── Sprint 2 (Lists & Tracking)
    │       │       │
    │       │       └── Sprint 3 (Payments & Resume)
    │       │               │
    │       │               └── Sprint 4 (AI Features)
    │       │                       │
    │       │                       └── Sprint 5 (Admin & Polish)
    │       │                               │
    │       │                               └── Sprint 6 (Launch Prep)
    │       │
    │       └── (Seed data needed for Sprint 2+)
    │
    └── (Auth needed for Sprint 2+)
```

Each sprint builds on the previous. No sprint can start until its predecessor is complete.

---

## 3. Sprint 0 — Foundation (Weeks 1-2)

**Goal:** Deployable skeleton with auth, CI, and local dev environment.

### 3.1 Tasks

| # | Task | Files/Components | Est. Hours |
|:-:|------|-----------------|:---:|
| 0.1 | Initialise Go module (`backend/`) | `go.mod`, `go.sum` | 1 |
| 0.2 | Initialise Next.js project (`frontend/`) | `package.json`, `tsconfig.json`, `tailwind.config.ts` | 1 |
| 0.3 | Docker Compose (local dev) | `docker-compose.yml` — Postgres, Redis, MinIO, Mailhog | 2 |
| 0.4 | Makefile | All targets from [CODE-STRUCTURE.md §6.2](./CODE-STRUCTURE.md) | 1 |
| 0.5 | Air hot-reload config | `.air.api.toml`, `.air.worker.toml` | 0.5 |
| 0.6 | CI pipeline | `.github/workflows/ci.yml` — lint, test, build for both backend + frontend | 3 |
| 0.7 | Pre-commit hooks | `.pre-commit-config.yaml`, `.golangci.yml`, `.eslintrc.json`, `.prettierrc` | 2 |
| 0.8 | `.env.example` + `.gitignore` | Root config files | 0.5 |
| 0.9 | PR template | `.github/pull_request_template.md` | 0.5 |
| 0.10 | Backend config module | `internal/config/` — Viper setup, validation | 3 |
| 0.11 | Domain layer | `internal/domain/` — entities, interfaces, errors, enums | 4 |
| 0.12 | Database migrations (initial schema) | `migrations/000001_initial_schema.up.sql` — all 18 tables from [database.md](./LLD/database.md) | 6 |
| 0.13 | Migration runner | `cmd/migrate/main.go` | 2 |
| 0.14 | Repository layer (user + session) | `internal/repository/user_repo.go`, `transactor.go` | 4 |
| 0.15 | Auth service | `internal/service/auth_service.go` — register, login, refresh, logout, password reset | 8 |
| 0.16 | Auth middleware | `internal/middleware/auth.go` — JWT verify, role check, rate limit | 4 |
| 0.17 | Request ID + logging middleware | `internal/middleware/request_id.go`, `logger.go` | 2 |
| 0.18 | CORS middleware | `internal/middleware/cors.go` | 1 |
| 0.19 | Response helpers | `internal/handler/response.go` — respondJSON, respondError | 2 |
| 0.20 | Auth handlers | `internal/handler/auth.go` — register, login, refresh, logout, me | 4 |
| 0.21 | Route mounting | `internal/handler/routes.go` | 2 |
| 0.22 | API server entry point | `cmd/api/main.go` — DI wiring, graceful shutdown | 4 |
| 0.23 | Health check endpoint | `GET /api/health` | 1 |
| 0.24 | Worker entry point (skeleton) | `cmd/worker/main.go` — Asynq server, no tasks yet | 2 |
| 0.25 | Email service (skeleton) | `internal/email/` — Resend integration, email verification template | 3 |
| 0.26 | Frontend: App shell | `app/layout.tsx`, route groups, Tailwind setup, basic components | 4 |
| 0.27 | Frontend: Auth pages | Login, register, forgot-password, reset-password pages | 6 |
| 0.28 | Frontend: API client | `lib/api-client.ts`, `lib/query-keys.ts` | 3 |
| 0.29 | Frontend: Auth store + hook | `store/auth-store.ts`, `hooks/use-auth.ts` | 3 |
| 0.30 | Seed script (skeleton) | `cmd/seed/main.go` — reads JSON, inserts companies | 2 |
| 0.31 | Initial ADRs | `docs/ADR/001-006` (UUID v7, VARCHAR enums, trunk-based dev, DI, raw SQL, Zustand) | 2 |

**Total estimated: ~81 hours (~2 weeks at 40h/week)**

### 3.2 Definition of Done

- [ ] `make dev` starts Postgres, Redis, MinIO, Mailhog
- [ ] `make dev-api` starts API with hot reload, `GET /api/health` returns 200
- [ ] `make dev-worker` starts worker (no tasks yet)
- [ ] `make dev-frontend` starts Next.js on :3000
- [ ] User can register, verify email, login, refresh token, logout
- [ ] CI pipeline passes (lint + test + build) on PR
- [ ] `make migrate` applies all 18 tables
- [ ] `make migrate-down` rolls back cleanly

---

## 4. Sprint 1 — Company Directory (Weeks 3-4)

**Goal:** Public browsable company directory with search and SEO-optimised profile pages.

### 4.1 Tasks

| # | Task | Files/Components | Est. Hours |
|:-:|------|-----------------|:---:|
| 1.1 | Company repository | `internal/repository/company_repo.go` — CRUD, search (FTS), list with cursor pagination | 6 |
| 1.2 | Company service | `internal/service/company_service.go` — list, search, getBySlug, filter | 4 |
| 1.3 | Company handlers | `internal/handler/company.go` — public endpoints | 4 |
| 1.4 | Company routes | Mount in `routes.go` | 1 |
| 1.5 | Seed data file | `seeds/companies.json` — 50-100 Indian tech companies | 8 |
| 1.6 | Seed runner | `cmd/seed/main.go` — reads JSON, upserts companies | 3 |
| 1.7 | Frontend: Company list page | `app/(public)/companies/page.tsx` — grid/list view, search bar, filters (size, domain, tech stack), infinite scroll | 8 |
| 1.8 | Frontend: Company profile page | `app/(public)/companies/[slug]/page.tsx` — SSR, SEO meta tags, all company fields | 6 |
| 1.9 | Frontend: Company components | `CompanyCard`, `CompanyFilters`, `CompanySearchBar`, `TechStackTags` | 6 |
| 1.10 | Frontend: Service Worker + IndexedDB | Offline caching for company directory | 6 |
| 1.11 | S3/MinIO integration | `internal/storage/` — upload, download, signed URL, delete | 4 |
| 1.12 | Company logo upload (admin) | Logo stored in S3 logos bucket | 2 |
| 1.13 | Landing page | `app/page.tsx` — hero, features, CTA | 4 |
| 1.14 | Pricing page | `app/(public)/pricing/page.tsx` — Starter Pack, à la carte | 3 |
| 1.15 | Frontend: Header + Footer | `components/layout/Header.tsx`, `Footer.tsx` | 3 |
| 1.16 | ETags / cache headers | Company list + profile cache headers for CDN | 2 |

**Total estimated: ~70 hours**

### 4.2 Definition of Done

- [ ] `GET /api/companies` returns paginated company list
- [ ] `GET /api/companies/search?q=tata` returns FTS results
- [ ] `GET /api/companies/{slug}` returns company profile
- [ ] `make seed` populates 50+ companies
- [ ] Company list page renders with search, filter, pagination
- [ ] Company profile page is SSR with correct meta tags (OG, Twitter)
- [ ] Offline: previously viewed companies available without network

---

## 5. Sprint 2 — User Lists & Tracking (Weeks 5-6)

**Goal:** Users can create lists, add companies, and track applications.

### 5.1 Tasks

| # | Task | Files/Components | Est. Hours |
|:-:|------|-----------------|:---:|
| 2.1 | List repository | `internal/repository/list_repo.go` — CRUD, entries, cursor pagination | 6 |
| 2.2 | List entry repository | CRUD for list entries with application status, notes, dates | 4 |
| 2.3 | List service | `internal/service/list_service.go` — enforce 3-list limit (free) / 5-list limit (premium) | 4 |
| 2.4 | List handlers | `internal/handler/list.go` — all list + entry endpoints | 4 |
| 2.5 | List routes | Mount authenticated routes | 1 |
| 2.6 | Frontend: Lists page | `app/(dashboard)/lists/page.tsx` — list cards, create modal | 4 |
| 2.7 | Frontend: List detail page | `app/(dashboard)/lists/[id]/page.tsx` — entry table, add company, status tracking | 8 |
| 2.8 | Frontend: Application tracker | Status pipeline (wishlist → applied → screening → interviewing → offer), date tracking, notes | 6 |
| 2.9 | Frontend: Dashboard (free) | `app/(dashboard)/dashboard/page.tsx` — funnel view (counts per status), recent activity, quick add | 6 |
| 2.10 | Frontend: Dashboard layout | `app/(dashboard)/layout.tsx` — sidebar with nav links, top bar | 4 |
| 2.11 | Frontend: Settings page | `app/(dashboard)/settings/page.tsx` — profile edit, password change, account deletion | 4 |
| 2.12 | User service | `internal/service/user_service.go` — profile update, password change, soft delete | 3 |
| 2.13 | User handlers | `internal/handler/user.go` — settings endpoints | 2 |
| 2.14 | Notification model (skeleton) | `internal/repository/notification_repo.go` — create, list, mark read | 3 |
| 2.15 | SSE endpoint (skeleton) | `GET /api/events` — authenticated, sends job completion events | 4 |

**Total estimated: ~63 hours**

### 5.2 Definition of Done

- [ ] User can create up to 3 lists (free tier)
- [ ] User can add companies to lists with application status
- [ ] Status transitions work (wishlist → applied → ... → accepted/rejected)
- [ ] Dashboard shows funnel view with correct counts
- [ ] Settings: profile edit, password change, account deletion work
- [ ] SSE: endpoint accepts connections (no events emitted yet)

---

## 6. Sprint 3 — Payments & Resume Foundation (Weeks 7-8)

**Goal:** Razorpay integration, credit system, resume upload, and AI service foundation.

### 6.1 Tasks

| # | Task | Files/Components | Est. Hours |
|:-:|------|-----------------|:---:|
| 3.1 | Payment repository | `internal/repository/payment_repo.go` — create order, update status, list | 4 |
| 3.2 | Credit repository | `internal/repository/credit_repo.go` — allocate, deduct (SELECT FOR UPDATE), balance, log transaction | 5 |
| 3.3 | Payment service | `internal/service/payment_service.go` — create Razorpay order, handle webhook, refund | 8 |
| 3.4 | Razorpay adapter | `internal/payment/razorpay.go` — create order API, verify signature, refund API | 6 |
| 3.5 | Payment handlers | `internal/handler/payment.go` — create order, webhook, history | 4 |
| 3.6 | Webhook signature verification | HMAC-SHA256 middleware for `/api/webhooks/razorpay` | 2 |
| 3.7 | Credit service | `internal/service/credit_service.go` — balance check, deduction, premium gating | 3 |
| 3.8 | Premium middleware | `auth.RequirePremium` — checks `premium_since IS NOT NULL` | 1 |
| 3.9 | Resume repository | `internal/repository/resume_repo.go` — CRUD, list by user | 3 |
| 3.10 | Resume service | `internal/service/resume_service.go` — upload (validate PDF, store S3, create record), archive, download URL | 5 |
| 3.11 | Resume handlers | `internal/handler/resume.go` — upload (multipart), list, get, archive | 4 |
| 3.12 | PDF extraction | `internal/pdf/extractor.go` — extract text from PDF using pdfcpu or unipdf | 4 |
| 3.13 | AI provider interface | `internal/ai/provider.go` — interface definition matching [ai-service.md](./LLD/ai-service.md) | 2 |
| 3.14 | Claude provider | `internal/ai/claude.go` — API client, callWithPDF, response parsing | 6 |
| 3.15 | OpenAI provider | `internal/ai/openai.go` — API client, text-only calls, response parsing | 5 |
| 3.16 | Fallback provider | `internal/ai/fallback.go` — cache check → Claude → OpenAI → cache store | 4 |
| 3.17 | Prompt templates | `internal/ai/prompts/` — resume parse, all system/user prompts from [ai-service.md §4](./LLD/ai-service.md) | 4 |
| 3.18 | AI result cache | `internal/ai/cache.go` — Redis GET/SET with SHA256 keys, TTL per operation | 3 |
| 3.19 | Worker: resume parse+score task | `internal/worker/task_resume_parse.go` — extract text, parse with AI, general ATS score, store results | 6 |
| 3.20 | Worker: email send task | `internal/worker/task_email_send.go` — Resend integration | 2 |
| 3.21 | Frontend: Pricing + checkout | Razorpay Checkout.js integration, order creation, payment confirmation | 6 |
| 3.22 | Frontend: Resume management | `app/(dashboard)/resumes/page.tsx` — upload, list, status, download | 6 |
| 3.23 | Frontend: Credit balance display | Show credits in dashboard sidebar/header | 2 |

**Total estimated: ~95 hours (~2.5 weeks — may need buffer)**

### 6.2 Definition of Done

- [ ] User can purchase Starter Pack via Razorpay (test mode)
- [ ] Webhook correctly allocates credits and sets `premium_since`
- [ ] À la carte purchases work (resume upload credit, ATS bundle)
- [ ] Credit balance shown in UI, deducted on premium actions
- [ ] User can upload PDF resume (validated, stored in S3/MinIO)
- [ ] Resume parse + general ATS runs async, results stored in DB
- [ ] SSE notifies user when resume processing completes
- [ ] AI fallback: if Claude fails, OpenAI is used
- [ ] AI result cache: repeated requests return cached result

---

## 7. Sprint 4 — AI Features (Weeks 9-10)

**Goal:** All three ATS scoring tiers and AI-curated company lists.

### 7.1 Tasks

| # | Task | Files/Components | Est. Hours |
|:-:|------|-----------------|:---:|
| 4.1 | ATS check repository | `internal/repository/ats_repo.go` — create, get, list by user/resume | 4 |
| 4.2 | ATS service | `internal/service/ats_service.go` — company check, job check, credit deduction | 5 |
| 4.3 | ATS handlers | `internal/handler/ats.go` — POST /ats/company, POST /ats/job, GET /ats/:id, GET /ats/ | 4 |
| 4.4 | Worker: company ATS task | `internal/worker/task_ats_company.go` — download PDF, score with AI, store result | 5 |
| 4.5 | Worker: job ATS task | `internal/worker/task_ats_job.go` — download PDF, score with AI (JD text input), store result | 5 |
| 4.6 | Curated list repository | `internal/repository/curated_list_repo.go` — create, get, list | 3 |
| 4.7 | Curated list service | `internal/service/curated_list_service.go` — trigger generation, credit deduction | 3 |
| 4.8 | Curated list handler | `internal/handler/curated_list.go` — POST /curated-lists, GET /curated-lists/:id | 3 |
| 4.9 | Worker: curate company list task | `internal/worker/task_curate_list.go` — build candidate profile, query companies, AI curation | 5 |
| 4.10 | AI prompt templates (remaining) | Company ATS, Job ATS, Curated List prompts | 3 |
| 4.11 | Output validation | `internal/ai/validation.go` — schema validation for all AI responses, score bounds, retry logic | 4 |
| 4.12 | Frontend: ATS check page | `app/(dashboard)/ats/page.tsx` — select resume, choose company or paste JD, submit | 6 |
| 4.13 | Frontend: ATS result page | `app/(dashboard)/ats/[id]/page.tsx` — score display, breakdown, recommendations | 6 |
| 4.14 | Frontend: Curated lists page | `app/(dashboard)/curated-lists/page.tsx` — generate new, view results | 5 |
| 4.15 | Frontend: Premium dashboard | Enhanced dashboard — resume health, credit tracker, recent ATS scores, quick actions | 6 |
| 4.16 | Frontend: SSE integration | `hooks/use-sse.ts` — listen for job completion events, update UI | 4 |
| 4.17 | Asynq scheduler setup | Register periodic tasks (user hard-delete cleanup) | 2 |

**Total estimated: ~78 hours**

### 7.2 Definition of Done

- [ ] Company ATS check: user selects resume + company, gets score with breakdown
- [ ] Job ATS check: user selects resume + pastes JD, gets score with breakdown
- [ ] Curated list: user generates AI-curated company list based on resume profile
- [ ] All AI operations are async — user sees loading state, gets SSE notification on completion
- [ ] Credit deduction works correctly for each operation
- [ ] Results are cached — repeat requests return instantly
- [ ] AI fallback works end-to-end (Claude → OpenAI)
- [ ] Premium dashboard shows resume health, credits, recent activity

---

## 8. Sprint 5 — Admin & Polish (Weeks 11-12)

**Goal:** Admin panel, feature flags, monitoring, security hardening, and UX polish.

### 8.1 Tasks

| # | Task | Files/Components | Est. Hours |
|:-:|------|-----------------|:---:|
| 5.1 | Admin service | `internal/service/admin_service.go` — dashboard stats, user management, company management | 6 |
| 5.2 | Admin handlers | `internal/handler/admin.go` — all 22 admin endpoints from [ADMIN-PANEL.md §11](./ADMIN-PANEL.md) | 8 |
| 5.3 | Audit log service | `internal/service/audit_service.go` — log all admin actions | 3 |
| 5.4 | Feature flag service | `internal/service/feature_flag_service.go` — CRUD, Redis cache, IsEnabled() | 4 |
| 5.5 | Feature flag repository | `internal/repository/feature_flag_repo.go` | 2 |
| 5.6 | Moderation service | `internal/service/moderation_service.go` — list edits, approve, reject | 4 |
| 5.7 | Company edit handlers | Submit edit (moderator), review endpoints (admin) | 3 |
| 5.8 | Worker: company enrich task | `internal/worker/task_company_enrich.go` — AI enrichment | 3 |
| 5.9 | Worker: company refresh task | `internal/worker/task_company_refresh.go` — AI full refresh | 3 |
| 5.10 | Worker: alert check task | `internal/worker/task_alert_check.go` — AI cost, stuck payments, failed jobs | 3 |
| 5.11 | Worker: user cleanup task | `internal/worker/task_user_cleanup.go` — hard-delete expired soft-deleted users + S3 cleanup | 3 |
| 5.12 | Frontend: Admin layout | `app/(admin)/admin/layout.tsx` — sidebar, top bar | 3 |
| 5.13 | Frontend: Admin dashboard | `app/(admin)/admin/page.tsx` — stat cards, revenue chart | 4 |
| 5.14 | Frontend: User management | `app/(admin)/admin/users/` — list, detail, suspend/unsuspend | 5 |
| 5.15 | Frontend: Company management | `app/(admin)/admin/companies/` — list, create, edit, delete | 5 |
| 5.16 | Frontend: Moderation queue | `app/(admin)/admin/moderation/` — list, diff review | 4 |
| 5.17 | Frontend: Payment management | `app/(admin)/admin/payments/` — list, refund flow | 4 |
| 5.18 | Frontend: AI cost dashboard | `app/(admin)/admin/ai/` — cost chart, breakdown table | 4 |
| 5.19 | Frontend: Feature flags | `app/(admin)/admin/features/` — toggle UI | 3 |
| 5.20 | Frontend: Audit log | `app/(admin)/admin/audit-log/` — filterable log | 3 |
| 5.21 | Prometheus metrics middleware | `internal/middleware/metrics.go` — HTTP, DB, Redis, AI, job metrics | 4 |
| 5.22 | Sentry integration | Backend + worker + frontend setup | 3 |
| 5.23 | Nginx security headers | Update `nginx.conf` with all headers from [SECURITY.md §2](./SECURITY.md) | 1 |
| 5.24 | Brute-force lockout | `internal/middleware/brute_force.go` — 5 attempts / 15 min | 2 |
| 5.25 | Log redaction handler | Custom slog handler that redacts sensitive fields | 2 |
| 5.26 | Seed feature flags | Initial 6 flags from [ADMIN-PANEL.md §9.4](./ADMIN-PANEL.md) | 1 |
| 5.27 | Data export endpoint | `GET /api/settings/export` — user data as JSON | 3 |

**Total estimated: ~93 hours (~2.5 weeks — may need buffer)**

### 8.2 Definition of Done

- [ ] Admin dashboard shows live stats (users, revenue, companies, AI costs)
- [ ] Admin can manage users (suspend, unsuspend, change role)
- [ ] Admin can manage companies (CRUD, AI enrich/refresh)
- [ ] Moderation queue works (submit edit as moderator, review as admin)
- [ ] Admin can issue refunds (with pre-condition checks)
- [ ] Feature flags toggleable from admin UI without redeploy
- [ ] AI cost dashboard shows cost breakdown and trend
- [ ] Audit log captures all admin actions
- [ ] Prometheus metrics exposed, Sentry catching errors
- [ ] Brute-force lockout active on login
- [ ] Data export endpoint returns user's data

---

## 9. Sprint 6 — Launch Prep (Weeks 13-14)

**Goal:** Production deployment, testing, and beta.

### 9.1 Tasks

| # | Task | Files/Components | Est. Hours |
|:-:|------|-----------------|:---:|
| 6.1 | Production infrastructure | EC2, RDS, S3, CloudFront, ECR setup per [DEPLOYMENT.md §3](./DEPLOYMENT.md) | 8 |
| 6.2 | DNS + SSL | Hostinger records, Let's Encrypt cert, auto-renewal cron | 3 |
| 6.3 | Secrets Manager setup | Create secret, load-secrets.sh, IAM policies | 3 |
| 6.4 | Deploy workflow | `.github/workflows/deploy.yml` — full pipeline from [DEPLOYMENT.md §5.3](./DEPLOYMENT.md) | 4 |
| 6.5 | Production Docker Compose | `docker-compose.prod.yml` + `nginx.conf` | 3 |
| 6.6 | First production deploy | Tag v1.0.0-rc1, run pipeline, verify health | 4 |
| 6.7 | Grafana Cloud + Alloy | Grafana account, agent config, dashboards from [MONITORING.md §7](./MONITORING.md) | 4 |
| 6.8 | UptimeRobot monitors | API health + frontend monitors | 1 |
| 6.9 | Sentry projects | Backend + frontend projects, DSN configuration | 1 |
| 6.10 | CloudWatch log verification | Verify logs flowing, test queries from [MONITORING.md §8](./MONITORING.md) | 2 |
| 6.11 | Dependabot | `.github/dependabot.yml` from [SECURITY.md §10.2](./SECURITY.md) | 1 |
| 6.12 | Seed production data | 100+ companies, admin user, feature flags | 3 |
| 6.13 | Load testing | Test with 50-100 concurrent users on production infra | 6 |
| 6.14 | Security checklist | Run through [SECURITY.md §13](./SECURITY.md) checklist | 4 |
| 6.15 | Monitoring checklist | Run through [MONITORING.md §10](./MONITORING.md) checklist | 3 |
| 6.16 | Bug fixes + UX polish | Address issues found during testing | 12 |
| 6.17 | Vercel production deploy | Connect repo, set env vars, custom domain | 2 |
| 6.18 | Beta testing | Invite 5-10 users, gather feedback | 8 |
| 6.19 | README.md | Setup instructions, contributing guide | 2 |
| 6.20 | Tag v1.0.0 | Final release tag, production deploy | 1 |

**Total estimated: ~75 hours**

### 9.2 Definition of Done

- [ ] Production API responds at `https://api.careerdock.skriptvalley.com/api/health`
- [ ] Production frontend loads at `https://careerdock.skriptvalley.com`
- [ ] Full user journey works: register → browse → create list → pay → upload resume → get ATS score
- [ ] Admin panel functional on production
- [ ] Monitoring: Grafana dashboards show data, Sentry captures errors, UptimeRobot is green
- [ ] Security checklist complete
- [ ] Beta users have provided feedback
- [ ] v1.0.0 tagged and deployed

---

## 10. Effort Summary

| Sprint | Estimated Hours | Weeks (40h/week) |
|:---:|:---:|:---:|
| 0 — Foundation | 81 | 2.0 |
| 1 — Company Directory | 70 | 1.8 |
| 2 — Lists & Tracking | 63 | 1.6 |
| 3 — Payments & Resume | 95 | 2.4 |
| 4 — AI Features | 78 | 2.0 |
| 5 — Admin & Polish | 93 | 2.3 |
| 6 — Launch Prep | 75 | 1.9 |
| **Total** | **555** | **~14 weeks** |

**Buffer recommendation:** Add 20% buffer for unexpected issues → **~17 weeks total**.

Sprint 3 and Sprint 5 are the heaviest. Consider extending them to 2.5 weeks each if needed.

---

## 11. Risk Register

| Risk | Probability | Impact | Mitigation |
|------|:---:|:---:|-----------|
| AI API costs exceed estimates | Medium | Medium | Aggressive caching, token budgets, OpenAI fallback (cheaper), feature flag kill switch |
| Razorpay integration complexity | Low | High | Use test mode throughout development. Razorpay docs are well-maintained |
| PDF extraction quality | Medium | Medium | Benchmark pdfcpu vs unipdf early in Sprint 3. Claude PDF-native is the primary path |
| Scope creep | High | High | Strict MVP scope. All v2 features are explicitly deferred. Feature flags for WIP |
| Solo founder burnout | Medium | High | 2-week sprints with clear DoD. No overtime — sustainable pace |
| Free tier limits exceeded | Low | Low | Grafana 10K metrics, Sentry 5K events, CloudWatch 5GB — all have significant headroom |

---

## 12. Success Metrics (3 Months Post-Launch)

| Metric | Target |
|--------|--------|
| Registered users | 500+ |
| Starter Pack purchases | 50+ (10% conversion) |
| Companies in directory | 200+ |
| Average ATS checks per premium user | 8+ |
| Resume uploads | 100+ |
| Monthly active users (MAU) | 40%+ retention |
| Platform uptime | 99%+ |
| AI operation success rate | 98%+ |
| Monthly infrastructure cost | < ₹3,000 |
| Gross margin (AI cost vs revenue) | > 85% |

---

## 13. Post-Launch Roadmap (v2)

After v1.0.0 launch and initial user feedback:

| Priority | Feature | Trigger |
|:---:|---------|---------|
| 1 | CV generation for specific JDs | User demand from beta feedback |
| 2 | OAuth (Google) login | Reduce registration friction |
| 3 | Job description URL auto-fetch | Scrape JD from career pages |
| 4 | MeiliSearch for company search | If Postgres FTS is insufficient at 500+ companies |
| 5 | Bulk company import (CSV) | When company count needs to scale past 200 |
| 6 | Admin MFA | When team grows beyond solo founder |
| 7 | User impersonation | When support volume requires it |
| 8 | Follow-up reminders | Calendar integration for interview dates |
| 9 | Mobile PWA enhancements | Based on mobile traffic analytics |
| 10 | Community contributions | Beyond moderator edits, when community grows |

---

## 14. Document Cross-Reference

Every implementation task traces back to a design document:

| Design Document | Sprint References |
|----------------|-------------------|
| [PRD.md](./PRD.md) | All sprints (feature requirements) |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Sprint 0 (tech stack), Sprint 3 (AI module), Sprint 6 (infra) |
| [LLD/database.md](./LLD/database.md) | Sprint 0 (migrations), all sprints (repository layer) |
| [LLD/api.md](./LLD/api.md) | All sprints (endpoint implementation) |
| [LLD/payments.md](./LLD/payments.md) | Sprint 3 (Razorpay, credits) |
| [LLD/ai-service.md](./LLD/ai-service.md) | Sprint 3 (foundation), Sprint 4 (ATS, curated lists) |
| [LLD/frontend.md](./LLD/frontend.md) | All sprints (pages, components) |
| [CODE-STRUCTURE.md](./CODE-STRUCTURE.md) | Sprint 0 (project setup, patterns) |
| [SECURITY.md](./SECURITY.md) | Sprint 0 (auth), Sprint 5 (hardening), Sprint 6 (audit) |
| [DEPLOYMENT.md](./DEPLOYMENT.md) | Sprint 6 (production setup) |
| [ADMIN-PANEL.md](./ADMIN-PANEL.md) | Sprint 5 (admin panel) |
| [MONITORING.md](./MONITORING.md) | Sprint 5 (metrics, Sentry), Sprint 6 (Grafana, monitoring) |
