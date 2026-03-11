# CareerDock — High-Level Architecture

> **Version:** 1.0
> **Status:** Final (Phase 2)
> **Last updated:** 2026-03-11
> **Depends on:** [PRD.md](./PRD.md)

---

## 1. Architecture Overview

CareerDock is a monolithic application with async worker processing, deployed on AWS. The system consists of:

- **Next.js frontend** (Vercel) — SSR for public pages, SPA for authenticated pages
- **Go backend** (Chi) — REST API server
- **Go worker** (Asynq) — async job processor for AI operations
- **PostgreSQL** (RDS) — primary data store
- **Redis** (on EC2) — sessions, rate limiting, job queues, AI result cache
- **S3** — resume PDF archival, company logos
- **Third-party services** — Claude API, Razorpay, Resend, Sentry

### 1.1 Design Principles

1. **Single deployable backend** — API server + worker from the same Go binary (`cmd/api`, `cmd/worker`). No microservices.
2. **Async AI operations** — All LLM calls go through the job queue. No synchronous AI in request handlers.
3. **Postgres-first storage** — Resume text and structured data live in Postgres for fast access. S3 is archival only.
4. **Cost-conscious** — Minimal managed services. Redis co-located on EC2. Vercel free tier for frontend.
5. **Horizontal scaling deferred** — Single EC2 instance for MVP. Architecture supports scaling but doesn't require it.

---

## 2. System Architecture Diagram

```
                    ┌─────────────────────────┐
                    │       DNS (Route 53)     │
                    └────┬───────────────┬─────┘
                         │               │
                    ┌────▼────┐    ┌─────▼──────┐
                    │ Vercel  │    │ CloudFront  │
                    │(Next.js)│    │  (Logos/    │
                    │ Frontend│    │   Assets)   │
                    └────┬────┘    └─────┬───────┘
                         │               │
                         │ REST API      │ S3 Origin
                         ▼               ▼
              ┌──────────────────────────────────────┐
              │         EC2 (t3.medium)               │
              │         Docker Compose                │
              │                                       │
              │  ┌─────────────┐  ┌────────────────┐ │
              │  │  Go API     │  │   Go Worker    │ │
              │  │  (Chi)      │  │   (Asynq)      │ │
              │  │  :8080      │  │                 │ │
              │  │             │  │  Resume parse   │ │
              │  │  Handlers   │  │  ATS scoring    │ │
              │  │  Services   │  │  Curated lists  │ │
              │  │  Middleware │  │  Company enrich │ │
              │  │             │  │  Email sending  │ │
              │  └──────┬──────┘  └───────┬─────────┘ │
              │         │                 │            │
              │         ▼                 ▼            │
              │  ┌─────────────────────────────────┐  │
              │  │          Redis (:6379)           │  │
              │  │                                  │  │
              │  │  - Asynq job queues              │  │
              │  │  - Session store                 │  │
              │  │  - Rate limit counters           │  │
              │  │  - AI result cache (TTL 30d)     │  │
              │  │  - Refresh token blacklist       │  │
              │  └─────────────────────────────────┘  │
              └──────────┬────────────────────────────┘
                         │
              ┌──────────▼────────────────────────────┐
              │       RDS PostgreSQL (db.t3.micro)     │
              │                                        │
              │  users, companies, resumes (text+JSON), │
              │  lists, ats_checks, payments, credits,  │
              │  feature_flags, audit_log, company_edits│
              └────────────────────────────────────────┘

              ┌────────────────────────────────────────┐
              │              AWS S3                     │
              │                                        │
              │  careerdock-resumes/ (private)          │
              │    └── {user_id}/{resume_id}.pdf        │
              │  careerdock-logos/ (public via CF)       │
              │    └── {company_slug}.png               │
              └────────────────────────────────────────┘

              ┌────────────────────────────────────────┐
              │         External Services               │
              │                                        │
              │  Claude API ──── AI operations          │
              │  OpenAI API ──── Fallback               │
              │  Razorpay ────── Payments + Webhooks    │
              │  Resend ──────── Transactional email    │
              │  Sentry ──────── Error tracking         │
              └────────────────────────────────────────┘
```

---

## 3. Component Details

### 3.1 Frontend — Next.js (App Router)

| Aspect | Decision |
|--------|----------|
| Framework | Next.js 14+ with App Router |
| Hosting | Vercel (free tier) |
| Rendering | SSR for `/companies/*` (SEO), client-side for authenticated routes |
| State management | **Zustand** for client state + **TanStack Query** for server state/caching |
| Styling | Tailwind CSS |
| Offline | Service Worker + IndexedDB for company directory caching |
| API communication | REST via fetch, TanStack Query for caching/revalidation |
| Auth | httpOnly cookies (JWT), no client-side token handling |
| Notifications | SSE (Server-Sent Events) for async job completion |
| Admin panel | Integrated under `/admin/*` routes with role-based access |

**Key frontend patterns:**
- Server components for data fetching on SSR pages.
- Client components with Zustand for interactive UI (list management, filters).
- TanStack Query for API calls with automatic caching, revalidation, and optimistic updates.
- SSE connection opened on premium dashboard for real-time job status updates.

### 3.2 Backend — Go (Chi)

| Aspect | Decision |
|--------|----------|
| Framework | Chi (stdlib `net/http` compatible) |
| Entry points | `cmd/api/` (HTTP server), `cmd/worker/` (Asynq processor), `cmd/migrate/`, `cmd/seed/` |
| Layering | Handler → Service → Repository (strict, no shortcuts) |
| Database driver | pgx/v5 (with pgxpool for connection pooling) |
| Redis client | go-redis/v9 |
| Config | Viper (env vars + config files, 12-factor) |
| Logging | slog (stdlib structured logging) |
| PDF text extraction | pdfcpu or unipdf (Go-native, no external dependencies) |
| API style | REST, JSON request/response |
| Error format | Standardised error envelope (see below) |

**Standard error response:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Resume file exceeds 5MB limit",
    "details": {
      "field": "file",
      "max_size_bytes": 5242880
    }
  }
}
```

**Backend package structure:**
```
backend/
├── cmd/
│   ├── api/           # HTTP server entry point
│   ├── worker/        # Asynq worker entry point
│   ├── migrate/       # Database migration runner
│   └── seed/          # Seed data loader
├── internal/
│   ├── config/        # Configuration (Viper)
│   ├── domain/        # Business entities, interfaces, errors
│   ├── handler/       # HTTP handlers (thin — validate, call service, respond)
│   ├── service/       # Business logic (orchestrates repos, AI, payments)
│   ├── repository/    # Database access (pgx queries)
│   ├── ai/            # LLM provider abstraction
│   │   ├── provider.go      # Interface definition
│   │   ├── claude.go        # Claude implementation
│   │   ├── openai.go        # OpenAI implementation
│   │   └── prompts/         # Prompt templates
│   ├── payment/       # Razorpay integration
│   ├── email/         # Resend/SES integration
│   ├── middleware/     # Auth, rate limiting, CORS, logging, request ID
│   ├── worker/        # Asynq task definitions and handlers
│   └── pdf/           # PDF text extraction
├── migrations/        # SQL migration files
└── seeds/             # Seed data (JSON files)
```

### 3.3 Database — PostgreSQL

| Aspect | Decision |
|--------|----------|
| Hosting | AWS RDS (db.t3.micro — free tier eligible for 12 months) |
| Version | PostgreSQL 16 |
| Driver | pgx/v5 with pgxpool |
| Migrations | golang-migrate (SQL-based, version controlled) |
| Search | PostgreSQL Full-Text Search (tsvector + GIN indexes) |
| Flexible fields | JSONB for interview patterns, compensation bands, resume parsed data |

**Full-Text Search setup (companies):**
```sql
-- Computed tsvector column combining searchable fields
ALTER TABLE companies ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(array_to_string(tech_stack, ' '), '')), 'B') ||
    setweight(to_tsvector('english', coalesce(array_to_string(domains, ' '), '')), 'C')
  ) STORED;

CREATE INDEX idx_companies_search ON companies USING GIN(search_vector);
```

At 200-500 companies, this provides sub-millisecond search with relevance ranking. MeiliSearch becomes relevant only if we need typo tolerance, faceted search, or 10K+ records.

### 3.4 Redis

| Purpose | Key Pattern | TTL |
|---------|-------------|-----|
| Session store | `session:{user_id}` | 7 days |
| Refresh token blacklist | `blacklist:{token_jti}` | 7 days |
| Rate limiting | `ratelimit:{user_id}:{endpoint}` | 1-15 minutes |
| AI result cache | `ats:{hash(resume_text+target)}` | 30 days |
| Asynq job queues | Managed by Asynq | N/A |

Redis runs as a Docker container on the same EC2 instance. No managed service needed at MVP scale. Migrate to ElastiCache when availability becomes critical.

### 3.5 Job Queue — Asynq

| Aspect | Decision |
|--------|----------|
| Library | hibiken/asynq |
| Backend | Redis (shared instance) |
| Worker | Separate process (`cmd/worker/`) |
| Concurrency | 5 workers (configurable) |
| Retry policy | 3 retries with exponential backoff |
| Dead letter | Failed jobs retained for 7 days, visible in admin dashboard |

**Task types:**
```go
const (
    TaskResumeParseAndScore  = "resume:parse_and_score"   // Parse + General ATS
    TaskATSCompanyCheck      = "ats:company_check"
    TaskATSJobCheck          = "ats:job_check"
    TaskCurateCompanyList    = "ai:curate_company_list"
    TaskCompanyEnrich        = "admin:company_enrich"
    TaskCompanyRefresh       = "admin:company_refresh"
    TaskSendEmail            = "email:send"
)
```

**Priority queues:**
- `critical` — payment webhooks, email verification
- `default` — ATS checks, curated lists
- `low` — company enrichment, data refresh

### 3.6 AI Module

#### 3.6.1 Provider Interface

```go
type LLMProvider interface {
    ParseResume(ctx context.Context, resumeText string) (*ParsedResume, error)
    ScoreATSGeneral(ctx context.Context, resumeText string, parsedData *ParsedResume) (*ATSResult, error)
    ScoreATSCompany(ctx context.Context, resumeText string, parsedData *ParsedResume, company *CompanyProfile) (*ATSResult, error)
    ScoreATSJob(ctx context.Context, resumeText string, parsedData *ParsedResume, jobDescription string) (*ATSResult, error)
    CurateCompanyList(ctx context.Context, parsedResume *ParsedResume, prefs *UserPreferences, companies []CompanyProfile) (*CuratedList, error)
    EnrichCompanyProfile(ctx context.Context, companyName string, existing *CompanyProfile) (*CompanyProfile, error)
}
```

#### 3.6.2 Provider Strategy

```
Request → Check cache (Redis) → Hit: return cached result
                               → Miss: Try Claude API
                                        → Success: cache + return
                                        → Failure: Try OpenAI API
                                                    → Success: cache + return
                                                    → Failure: Return error, retry via Asynq
```

#### 3.6.3 AI Token Budget Estimates

| Operation | Input Tokens (est.) | Output Tokens (est.) | Claude Sonnet Cost (est.) | Notes |
|-----------|--------------------:|---------------------:|--------------------------:|-------|
| Resume parse | ~2,000 | ~1,000 | ~₹0.50 | 2-page PDF text |
| General ATS score | ~2,500 | ~1,500 | ~₹0.80 | Resume text + scoring rubric |
| Company ATS score | ~3,500 | ~1,500 | ~₹1.00 | Resume + company profile |
| Job ATS score | ~4,000 | ~2,000 | ~₹1.20 | Resume + JD text |
| Curated list | ~8,000 | ~2,000 | ~₹2.00 | Resume + all company summaries |
| Company enrich | ~1,000 | ~2,000 | ~₹0.60 | Company name + existing data |

**Total AI cost per fully consumed Starter Pack:**
- 9 uploads × (parse ₹0.50 + general ATS ₹0.80) = ₹11.70
- 3 curated lists × ₹2.00 = ₹6.00
- 10 company ATS × ₹1.00 = ₹10.00
- 10 job ATS × ₹1.20 = ₹12.00
- **Total: ~₹40 AI cost per ₹399 pack = ~90% gross margin**

**CV Generation (future à la carte):** Estimated ~₹1.50-2.00 per generation. Suggested price: ₹29-49 per CV.

#### 3.6.4 AI Result Caching

All AI results are deterministic for the same inputs. Cache aggressively:

| Operation | Cache Key | TTL | Invalidation |
|-----------|-----------|-----|-------------|
| Resume parse | `parse:{sha256(pdf_bytes)}` | Forever (stored in DB) | On re-upload |
| General ATS | `ats_general:{sha256(resume_text)}` | Forever (stored in DB) | On re-upload |
| Company ATS | `ats_company:{sha256(resume_text)}:{company_id}` | 30 days | On company data update |
| Job ATS | `ats_job:{sha256(resume_text)}:{sha256(jd_text)}` | 30 days | N/A |
| Curated list | `curated:{sha256(resume_text)}:{prefs_hash}` | 7 days | On resume/pref change |

Parse results and general ATS scores are stored directly in the `resumes` table — they never expire. Company/Job ATS results are cached in Redis with TTL and also persisted in the `ats_checks` table.

### 3.7 File Storage — S3

#### 3.7.1 Resume Storage Strategy

**Postgres-first approach** — S3 is archival only:

```
User uploads PDF
  │
  ├─► S3: Store original PDF (archival, rare access)
  │        Bucket: careerdock-resumes (private)
  │        Key: {user_id}/{resume_id}.pdf
  │
  ├─► Go: Extract text from PDF (pdfcpu/unipdf)
  │        Store extracted_text in resumes.extracted_text (TEXT column)
  │
  └─► Asynq Job: Send extracted_text to Claude API
           Store parsed structured data in resumes.parsed_data (JSONB)
           Store general ATS score in resumes.ats_general (JSONB)
           Update resumes.status = 'ready'
           User confirms/edits parsed data on frontend
```

**Why Postgres-first?**
- ATS checks read `extracted_text` and `parsed_data` from Postgres — no S3 round trip.
- User confirmation flow reads/writes structured data in Postgres.
- S3 is only accessed for original PDF download (rare) via signed URLs (15-min expiry).
- At MVP scale (1000 users × 3 resumes × ~10KB text = 30MB) — trivial for Postgres.

#### 3.7.2 S3 Buckets

| Bucket | Access | Purpose | Lifecycle |
|--------|--------|---------|-----------|
| `careerdock-resumes` | Private (signed URLs) | Original resume PDFs | Retain 90 days after account deletion |
| `careerdock-logos` | Public (via CloudFront) | Company logos | Indefinite |

### 3.8 Payment — Razorpay

**Integration pattern:**

```
Frontend                    Backend                         Razorpay
   │                          │                                │
   │ POST /api/payments/order │                                │
   │─────────────────────────►│                                │
   │                          │ Create Order (amount, currency)│
   │                          │───────────────────────────────►│
   │                          │◄───────────────────────────────│
   │  { order_id, amount }    │   { id, amount, status }       │
   │◄─────────────────────────│                                │
   │                          │                                │
   │ Open Razorpay Checkout   │                                │
   │─────────────────────────────────────────────────────────►│
   │                          │                                │
   │                          │  Webhook: payment.captured     │
   │                          │◄───────────────────────────────│
   │                          │  Verify signature              │
   │                          │  Idempotency check             │
   │                          │  Allocate credits              │
   │                          │  Store transaction              │
   │                          │  ──► 200 OK                    │
   │                          │───────────────────────────────►│
   │                          │                                │
   │  SSE: credits_updated    │                                │
   │◄─────────────────────────│                                │
```

**Idempotency:** Each Razorpay order ID is stored in `payments` table with a unique constraint. Duplicate webhooks are detected and ignored.

### 3.9 Email — Resend

| Email Type | Trigger | Template |
|------------|---------|----------|
| Email verification | Registration | Verification link (24h expiry) |
| Password reset | Forgot password | Reset link (1h expiry) |
| Payment receipt | Successful payment | Order summary + credits |
| ATS result ready | Job completion | Brief notification with link |

Resend free tier: 3,000 emails/month — sufficient for MVP.

### 3.10 Notifications — Server-Sent Events (SSE)

For async job completion notifications (ATS results ready, resume parsing complete):

```
Frontend                         Backend
   │                                │
   │  GET /api/notifications/stream │
   │  Accept: text/event-stream     │
   │───────────────────────────────►│
   │                                │ Hold connection open
   │                                │
   │  (Asynq job completes)         │
   │                                │ Write to SSE stream:
   │  event: ats_result_ready       │
   │  data: {"check_id": "abc123"} │
   │◄───────────────────────────────│
   │                                │
   │  (Frontend refetches data)     │
```

- SSE connection opened when user is on premium dashboard or waiting for results.
- Fallback: polling every 10 seconds if SSE connection drops.
- No WebSocket needed — notifications are server→client only.

---

## 4. Infrastructure

### 4.1 AWS Architecture

```
┌──────────────────────────────────────────────────────┐
│                     AWS Account                       │
│                                                       │
│  ┌─────────────┐     ┌────────────────────────────┐  │
│  │  Route 53   │     │       CloudFront            │  │
│  │  DNS        │────►│  - api.careerdock.in        │  │
│  │             │     │  - assets.careerdock.in     │  │
│  └─────────────┘     └─────────────┬──────────────┘  │
│                                    │                  │
│                         ┌──────────▼──────────┐      │
│                         │  EC2 (t3.medium)    │      │
│                         │  2 vCPU, 4GB RAM    │      │
│                         │                     │      │
│                         │  Docker Compose:    │      │
│                         │  - go-api (:8080)   │      │
│                         │  - go-worker        │      │
│                         │  - redis (:6379)    │      │
│                         │  - nginx (:443)     │      │
│                         └──────────┬──────────┘      │
│                                    │                  │
│  ┌──────────────────┐   ┌─────────▼──────────┐      │
│  │  S3 Buckets      │   │  RDS PostgreSQL    │      │
│  │  - resumes       │   │  db.t3.micro       │      │
│  │  - logos         │   │  20GB gp3          │      │
│  └──────────────────┘   └────────────────────┘      │
│                                                       │
│  ┌──────────────────┐                                │
│  │  ACM (SSL)       │  Certs for careerdock.in      │
│  └──────────────────┘                                │
└──────────────────────────────────────────────────────┘

┌──────────────────┐
│  Vercel          │
│  Next.js frontend│
│  careerdock.in   │
└──────────────────┘
```

### 4.2 Cost Estimate (MVP)

| Service | Spec | Monthly Cost (₹) |
|---------|------|------------------:|
| EC2 | t3.medium (on-demand) | ~₹2,500 |
| RDS | db.t3.micro (free tier year 1) | ₹0 (then ~₹1,200) |
| S3 | <1GB storage + minimal requests | ~₹10 |
| CloudFront | Free tier (1TB/month) | ₹0 |
| Route 53 | 1 hosted zone + queries | ~₹50 |
| ACM | Free | ₹0 |
| Vercel | Free tier | ₹0 |
| Resend | Free tier (3K emails) | ₹0 |
| Sentry | Free tier (5K events) | ₹0 |
| **Total (Year 1)** | | **~₹2,600/month** |
| **Total (After free tier)** | | **~₹3,800/month** |

**Note:** EC2 cost can be reduced to ~₹1,500/month with a 1-year Reserved Instance or ~₹800/month with a 3-year RI. Spot instances are not recommended for a web server.

### 4.3 Docker Compose (Production on EC2)

```yaml
# Simplified production docker-compose
services:
  api:
    build: ./backend
    command: /app/api
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis

  worker:
    build: ./backend
    command: /app/worker
    environment:
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./infra/nginx/nginx.conf:/etc/nginx/nginx.conf
      - /etc/letsencrypt:/etc/letsencrypt
    depends_on:
      - api
```

### 4.4 Local Development Docker Compose

```yaml
# Local dev — includes all dependencies
services:
  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: careerdock
      POSTGRES_USER: careerdock
      POSTGRES_PASSWORD: devpassword
    volumes:
      - pg_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data

  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"   # SMTP
      - "8025:8025"   # Web UI

volumes:
  pg_data:
  minio_data:
```

---

## 5. Security Architecture (Summary)

Detailed security design is in Phase 5. Key architectural decisions:

| Concern | Approach |
|---------|----------|
| Authentication | JWT (15-min access) + refresh token (7-day) in httpOnly, Secure, SameSite=Strict cookies |
| Authorisation | Role-based middleware: visitor, user, premium, moderator, admin |
| API protection | Rate limiting (Redis-based), CORS whitelist, input validation |
| Resume access | Private S3 bucket, signed URLs (15-min), user-scoped access checks |
| Payment security | Razorpay webhook signature verification, idempotent processing |
| Secrets | AWS Secrets Manager (production), `.env` file (local dev) |
| HTTPS | ACM certificate + CloudFront (API), Vercel auto-SSL (frontend) |

---

## 6. Monitoring Architecture (Summary)

Detailed monitoring design is in Phase 8. Key decisions:

| Layer | Tool | Free Tier |
|-------|------|-----------|
| Error tracking | Sentry | 5K events/month |
| Metrics | Grafana Cloud | 10K metrics |
| Uptime | UptimeRobot | 50 monitors |
| Logging | CloudWatch Logs (via Docker driver) | 5GB ingestion/month |
| AI cost tracking | Custom admin dashboard (internal) | N/A |

---

## 7. Scaling Strategy

### 7.1 MVP (Current Architecture) — Up to ~1,000 Concurrent Users

Single EC2 instance handles all backend traffic. Bottlenecks at this scale:
- **CPU:** AI operations are offloaded to async workers — API stays responsive.
- **Database:** Connection pooling via pgx (max 20 connections). RDS db.t3.micro supports ~60 connections.
- **Redis:** Single instance easily handles 10K+ ops/sec.

### 7.2 Growth Phase — 1,000-10,000 Users

| Change | When | Effort |
|--------|------|--------|
| Upgrade EC2 to t3.large | CPU consistently >60% | Config change |
| Move Redis to ElastiCache | If EC2 memory constrained | Minimal (change connection string) |
| Upgrade RDS to db.t3.small | Connection limit or storage | Config change |
| Add read replica (RDS) | If read-heavy queries slow down | Moderate |
| Add CloudFront for API | If latency from edge locations matters | Moderate |

### 7.3 Scale Phase — 10,000+ Users

| Change | When | Effort |
|--------|------|--------|
| Move to ECS Fargate | Need multiple API instances | Moderate (Dockerised already) |
| Add ALB | Multiple API instances need load balancing | Moderate |
| Add MeiliSearch | If company directory exceeds 5K or needs typo tolerance | Moderate |
| Separate worker EC2 | AI processing starves API resources | Low (separate docker-compose) |
| S3 → CloudFront | If resume download frequency increases | Low |

---

## 8. Technology Stack Summary

| Layer | Technology | Version |
|-------|-----------|---------|
| **Frontend** | Next.js (App Router) | 14+ |
| Frontend state | Zustand + TanStack Query | Latest |
| Frontend styling | Tailwind CSS | 3.x |
| **Backend** | Go + Chi | Go 1.22+, Chi v5 |
| Database driver | pgx/v5 | Latest |
| Redis client | go-redis/v9 | Latest |
| Job queue | Asynq | Latest |
| PDF extraction | pdfcpu or unipdf | Latest |
| Config | Viper | Latest |
| Logging | slog (stdlib) | Go 1.21+ |
| **Database** | PostgreSQL | 16 |
| **Cache** | Redis | 7.x |
| **Search** | PostgreSQL FTS | (built-in) |
| **Infrastructure** | AWS (EC2, RDS, S3, CloudFront, Route 53, ACM) | — |
| Frontend hosting | Vercel | — |
| **AI (primary)** | Claude API (Sonnet) | Latest |
| **AI (fallback)** | OpenAI (GPT-4o-mini) | Latest |
| **Payments** | Razorpay | — |
| **Email** | Resend | — |
| **Error tracking** | Sentry | — |
| **Monitoring** | Grafana Cloud | Free tier |
| **Uptime** | UptimeRobot | Free tier |
| **CI/CD** | GitHub Actions | — |
| **Containerisation** | Docker + Docker Compose | — |

---

## 9. Open Items for Phase 3

Items to resolve during Low-Level Design:

1. **Database schema** — Complete ER diagram with all tables, indexes, constraints.
2. **API specification** — Full route list with request/response schemas.
3. **AI prompt templates** — Actual prompts for each operation with expected output schemas.
4. **Frontend component hierarchy** — Page structure, shared components, routing.
5. **Payment webhook handling** — Exact credit allocation logic and edge cases.
6. **SSE implementation** — Connection management, reconnection strategy, message format.
7. **PDF extraction library choice** — pdfcpu vs unipdf benchmarking with sample resumes.
8. **Migration strategy** — golang-migrate setup, versioning scheme, rollback approach.
