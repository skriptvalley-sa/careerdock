# CareerDock — Career Intelligence Platform

## Claude Code Master Prompt

> **How to use this prompt**: Copy this entire file as the initial prompt for Claude Code in your repo root. It is designed for a **discussion-first** workflow — Claude Code will walk through each phase with you, document decisions in `docs/`, and only begin implementation after you approve the design.

---

## PHASE 0: PROJECT BOOTSTRAP & CONTEXT

You are helping me build **CareerDock** — a SaaS career intelligence platform for job seekers (primarily tech professionals in India, expanding later). I am a senior backend engineer (Go, Kubernetes, AWS). This is a solo founder project that needs to be lean, maintainable, and incrementally shippable.

### Your Role
- Act as a **senior staff engineer + product architect** co-building this with me.
- Before writing ANY code, we will discuss and finalise designs phase-by-phase.
- Every decision, architecture diagram, and spec MUST be persisted in `docs/` within this repo.
- Challenge my assumptions. Suggest trade-offs. Propose simplifications where I'm over-engineering.
- Default currency context: INR. Target market: India-first.

### Working Agreements
1. **Discussion-first**: For each phase, present your recommendation, explain trade-offs, and wait for my approval before proceeding.
2. **Document everything**: Create/update markdown files in `docs/` for every finalised decision.
3. **Incremental delivery**: Design for MVP-first. Flag features that can be deferred to v2.
4. **Cost-conscious**: I'm bootstrapping. Optimise for low operational cost (serverless where sensible, managed services where they save time).
5. **Production-ready mindset**: Even for MVP, include proper error handling, logging, auth, and basic observability.

---

## PHASE 1: REQUIREMENTS REFINEMENT

Before any design, help me refine and validate the requirements below. Identify gaps, conflicts, and edge cases. Propose a finalised PRD and save it to `docs/PRD.md`.

### 1.1 Core Platform — Company Directory

**What exists:**
- A browsable directory/dashboard of tech companies (initially India-focused, ~200-500 companies).
- Each company has a **profile page** with:
  - Overview: Name, logo, size, HQ, founded year, careers page link, Glassdoor/AmbitionBox links.
  - Tech Stack: Languages, frameworks, infra, cloud providers (crowdsourced + AI-enriched).
  - Domains: What the company works on (fintech, SaaS, infra, etc.).
  - Interview Patterns: Role-wise breakdown (SDE-1, SDE-2, Senior, Staff) — number of rounds, typical question types (DSA, system design, behavioural, take-home), difficulty rating, common topics.
  - Compensation Bands: Approximate ranges by role (crowdsourced, clearly marked as estimates).
  - Hiring Status: Active/Paused/Unknown (can be community-updated or scraped).

**Open questions to discuss:**
- Where does company data come from initially? (Seed data strategy — manual curation, scraping, AI generation, community contribution?)
- How do we keep it fresh? (Community edits? Periodic AI refresh? Admin moderation queue?)
- Should interview patterns be structured (JSON schema) or semi-structured (rich text)?
- Do we need a "suggest an edit" flow for community contributions?

### 1.2 General Dashboard (No Auth Required)

- Browse and search the company directory.
- Filter by: tech stack, domain, company size, location, hiring status.
- Sort by: name, size, rating, recently updated.
- Company profile pages are publicly accessible (good for SEO).
- **Client-side caching**: Use browser cache (service worker / IndexedDB) for offline browsing of previously viewed companies and search results.
- No localStorage for auth tokens — use httpOnly cookies.

### 1.3 Authentication & User Management

- **Sign up / Login**: Email + password (with email verification). OAuth (Google) as stretch goal for v1.
- **User Profile**: Name, email, current role, experience level, preferred tech stacks, target domains.
- **Account Settings**: Change password, delete account, manage payment history.
- **Session Management**: JWT in httpOnly cookies, refresh token rotation.

### 1.4 Free Features (Authenticated Users)

#### 1.4.1 Company Lists (Curated Tracking Lists)
- A user can create up to **3 lists** (e.g., "Dream Companies", "Backup Options", "Startups").
- Each list is populated by **filtering the company directory and selecting** companies to add.
- Lists can be renamed, reordered, and companies can be added/removed later.
- Each company in a list has **application tracking fields** (manually updated by user):
  - Roles applied (free text or selection from known roles at that company).
  - Application status: `Not Applied` → `Applied` → `Phone Screen` → `Interview` → `Offer` → `Rejected` → `Accepted` → `Withdrawn`.
  - Date applied, notes, follow-up reminders (stretch).
  - Interview round tracking (stretch): Which round, date, outcome, notes.

#### 1.4.2 Free Dashboard
- Overview of: Total companies tracked, applications by status (funnel view), recent activity.
- Quick links to their lists.

### 1.5 Paid Features (Premium Users)

#### 1.5.1 Resume Management
- Upload up to **3 resumes** (PDF) on initial one-time purchase.
- Each resume is **individually analysed by AI** (Claude or OpenAI) upon upload:
  - Extracted skills, experience summary, years of experience, domains, education.
  - Stored as structured data for matching.
- One resume can be set as **default** — this configures defaults across the platform (e.g., suggested companies, default for ATS checks).
- **Resumes cannot be updated in-place** on the base plan — each upload is final and analysed once. Updates require purchasing an "update" action (see Pay Model).

#### 1.5.2 AI-Curated Company Lists
- Based on the default resume's extracted profile, the platform generates a **smart curated list** of best-matching companies from the directory.
- Matching criteria: tech stack overlap, domain relevance, experience level fit, hiring status.
- This is an **AI-powered operation** (uses Claude/OpenAI tokens) — runs once on resume upload/default change.
- User can regenerate (costs 1 action from their bundle, or included in initial purchase — discuss).

#### 1.5.3 ATS Scoring (3 Tiers)
**Tier 1 — General Resume Score** (included with upload):
- Each uploaded resume gets a general ATS compatibility score.
- Checks: formatting, keyword density, section completeness, readability, ATS-parser friendliness.
- Score breakdown with actionable suggestions.

**Tier 2 — Company-Specific Score** (up to 10 checks included):
- Score a resume against a specific company's profile (tech stack match, domain relevance, culture fit keywords).
- Shows gap analysis: "You're missing keywords X, Y, Z that this company values."

**Tier 3 — Job-Specific Score** (up to 10 checks included):
- User provides a **job posting URL** + selects one of their resumes.
- AI fetches/parses the job posting, scores the resume against it.
- Detailed match report: required vs. your skills, experience gap, keyword recommendations.

#### 1.5.4 Premium Dashboard Enhancements
- Resume health overview (scores at a glance for all resumes).
- AI-curated list summary with match scores.
- ATS check usage tracker (remaining checks by tier).
- Recommended actions: "Your resume scores low against Company X — consider updating keywords."

### 1.6 Pay Model

**Discuss and decide the exact pricing, but here's the structure:**

#### One-Time Purchase — "Starter Pack"
- Unlocks all paid features.
- Includes: 3 resume uploads (with general ATS scoring each), 10 company ATS checks, 10 job ATS checks, 1 AI-curated list generation, 5 custom lists (upgrade from 3 free).
- Price: ₹XXX (discuss — thinking ₹299-499 range).

#### Individual Action Purchases (À La Carte)
| Action | Included in Starter | Additional Purchase |
|--------|-------------------|-------------------|
| Upload new resume (max +2 more = 5 total) | 3 | ₹XX each (includes general ATS) |
| Update existing resume (re-analysis) | 0 | ₹XX per update (max 3 updates at a time) |
| ATS Check Bundle (Company or Job level) | 10 + 10 | 30 checks: ₹XX / 50: ₹XX / 100: ₹XX |
| New List Creation | 5 total (3 free + 2 bonus) | Bundle of 3 lists: ₹XX / 5 lists: ₹XX |
| AI-Curated List Regeneration | 1 | ₹XX per regeneration |

**Open questions:**
- Payment gateway: Razorpay (India-first) vs Stripe (global)?
- Should we have a subscription model or purely one-time + à la carte? (I'm leaning one-time for simplicity).
- How to handle refunds?
- Do we need a "trial" for premium features? (e.g., 1 free ATS check to demonstrate value?)
- Credit/wallet system vs direct purchase per action?

---

## PHASE 2: HIGH-LEVEL ARCHITECTURE

After PRD is finalised, design the system architecture. Save to `docs/ARCHITECTURE.md`.

### What to cover:

#### 2.1 System Components
- **Frontend**: React (Next.js for SSR/SEO on company pages) or plain React SPA? Discuss trade-offs.
- **Backend API**: Go (Gin/Echo/Chi — discuss). RESTful with potential GraphQL for complex queries?
- **Database**: PostgreSQL (primary). Redis (caching, sessions). Discuss if we need a search engine (MeiliSearch/Typesense for company search?).
- **AI Service**: Isolated microservice or module within the backend that handles all LLM interactions (resume parsing, ATS scoring, list curation). Must support both Claude and OpenAI as providers with a clean interface.
- **File Storage**: S3 (or compatible — MinIO for local dev) for resume PDFs and company logos.
- **Job Queue**: For async AI operations (resume analysis, ATS scoring). Options: Go-native (Asynq with Redis), or managed (SQS).
- **Payment Service**: Razorpay integration module.
- **Admin Panel**: Separate lightweight app or integrated? (Discuss).

#### 2.2 Architecture Diagram
- Create a Mermaid or ASCII diagram showing all components and their interactions.
- Show data flow for key operations: resume upload → analysis → storage, ATS check flow, payment flow.

#### 2.3 Third-Party Services
- LLM Provider: Claude API (primary) / OpenAI (fallback). Design the abstraction layer.
- Payment: Razorpay.
- Email: SES or Resend for transactional emails.
- File Storage: AWS S3.
- Monitoring: Discuss options (Prometheus + Grafana vs managed).
- Error Tracking: Sentry.

#### 2.4 Scalability Considerations
- What can be serverless (Lambda for AI processing)?
- Caching strategy (Redis for hot data, CDN for static assets, browser cache for directory).
- Rate limiting on AI endpoints (prevent abuse).
- Cost estimation for AI token usage per operation.

---

## PHASE 3: LOW-LEVEL DESIGN

After architecture is approved, design each component in detail. Save to `docs/LLD/` folder with separate files per component.

### 3.1 Database Design (`docs/LLD/database.md`)
- Complete ER diagram (Mermaid).
- All table schemas with types, constraints, indexes.
- Key tables (at minimum):
  - `users` — profile, auth, subscription status.
  - `companies` — all company data (consider JSONB for flexible fields like tech_stack, interview_patterns).
  - `company_edits` — moderation queue for community edits.
  - `user_lists` — list metadata (name, user_id, type: manual/ai-curated).
  - `list_companies` — junction table with application tracking fields.
  - `resumes` — metadata, S3 path, extracted data (JSONB), scores.
  - `ats_checks` — log of all ATS checks with scores and results.
  - `payments` — transaction log.
  - `user_credits` — remaining credits/actions by type.
  - `feature_flags` — platform-wide feature toggles.
  - `admin_audit_log` — admin action tracking.
- Migration strategy: golang-migrate or Atlas?

### 3.2 API Design (`docs/LLD/api.md`)
- Complete API spec (OpenAPI 3.0 or at minimum a detailed route list).
- Group by domain: Auth, Companies, Lists, Resumes, ATS, Payments, Admin.
- For each endpoint: method, path, request/response schema, auth requirement, rate limit.
- Error response format (standardised).
- Pagination strategy (cursor-based for feeds, offset for admin).

### 3.3 AI Service Design (`docs/LLD/ai-service.md`)
- Provider abstraction interface (Go interface with Claude and OpenAI implementations).
- Prompt templates for each operation:
  - Resume parsing → structured extraction.
  - General ATS scoring → score + breakdown.
  - Company ATS scoring → match analysis.
  - Job ATS scoring → job posting parsing + match.
  - Company list curation → matching algorithm.
- Token budget per operation (estimate costs).
- Retry and fallback strategy (Claude fails → try OpenAI).
- Response validation (ensure AI returns expected schema).
- Caching: Cache AI results aggressively (resume analysis doesn't change).

### 3.4 Frontend Design (`docs/LLD/frontend.md`)
- Page/route structure.
- Component hierarchy for key pages (Dashboard, Company Profile, List Manager, Resume Manager).
- State management approach (React Context + hooks? Zustand? Redux Toolkit?).
- Client-side caching strategy (service worker, IndexedDB for company data).
- Responsive design requirements (mobile-first?).

### 3.5 Payment Flow Design (`docs/LLD/payments.md`)
- Razorpay integration flow (order creation → payment → webhook → credit allocation).
- Idempotency handling.
- Credit system: How credits are tracked, consumed, and replenished.
- Receipt/invoice generation.
- Edge cases: failed payments, partial refunds, duplicate webhooks.

---

## PHASE 4: CODE STRUCTURE & MANAGEMENT

Design the repo structure and development workflow. Save to `docs/CODE-STRUCTURE.md`.

### 4.1 Repo Structure
```
careerdock/
├── docs/                    # All design documents
│   ├── PRD.md
│   ├── ARCHITECTURE.md
│   ├── LLD/
│   └── ADR/                 # Architecture Decision Records
├── backend/                 # Go backend
│   ├── cmd/                 # Entry points (api server, worker, migrate, seed)
│   ├── internal/
│   │   ├── domain/          # Business entities and interfaces
│   │   ├── handler/         # HTTP handlers
│   │   ├── service/         # Business logic
│   │   ├── repository/      # Database access
│   │   ├── ai/              # LLM provider abstraction
│   │   ├── payment/         # Razorpay integration
│   │   ├── middleware/       # Auth, rate limiting, logging
│   │   └── config/          # Configuration management
│   ├── migrations/          # SQL migrations
│   ├── seeds/               # Seed data (companies)
│   └── pkg/                 # Shared utilities
├── frontend/                # React/Next.js frontend
│   ├── src/
│   │   ├── pages/           # Route pages
│   │   ├── components/      # Reusable components
│   │   ├── hooks/           # Custom hooks
│   │   ├── services/        # API client
│   │   ├── store/           # State management
│   │   └── utils/           # Helpers
│   └── public/              # Static assets
├── admin/                   # Admin panel (if separate)
├── infra/                   # Infrastructure as Code
│   ├── terraform/           # AWS infra
│   ├── k8s/                 # Kubernetes manifests (if used)
│   └── docker/              # Dockerfiles
├── scripts/                 # Dev scripts, data seeding, etc.
├── .github/                 # CI/CD workflows
├── docker-compose.yml       # Local dev environment
├── Makefile                 # Common commands
└── README.md
```

### 4.2 Version Control Strategy
- Branch strategy: trunk-based or GitFlow? (Discuss — I prefer trunk-based with feature flags).
- Commit conventions: Conventional Commits.
- PR template with checklist.
- Protected main branch with CI checks.

### 4.3 Development Workflow
- Local dev environment: Docker Compose for all dependencies (Postgres, Redis, MinIO, etc.).
- Hot reload for both Go backend and React frontend.
- Makefile targets: `make dev`, `make test`, `make lint`, `make migrate`, `make seed`, `make build`.
- Pre-commit hooks: lint, format, test.

---

## PHASE 5: SECURITY DESIGN

Design the security model. Save to `docs/SECURITY.md`.

### Must Cover:
- **Authentication**: JWT + refresh tokens in httpOnly cookies. Token rotation.
- **Authorisation**: Role-based (user, premium_user, admin). Middleware enforcement.
- **Data Protection**: Encryption at rest (S3, database). HTTPS everywhere. PII handling.
- **Resume Security**: Resumes are sensitive — access control, signed S3 URLs (time-limited), no public access.
- **API Security**: Rate limiting (per user, per IP). CORS policy. Input validation. SQL injection prevention (parameterised queries). XSS prevention.
- **Payment Security**: PCI compliance considerations (Razorpay handles card data). Webhook signature verification.
- **Admin Security**: Separate auth, MFA (stretch), audit logging.
- **Secrets Management**: How API keys, DB credentials, LLM tokens are stored (env vars, AWS Secrets Manager, etc.).
- **OWASP Top 10**: Brief review against each.

---

## PHASE 6: DEPLOYMENT, RELEASE & UPGRADE DESIGN

Design the deployment pipeline and release management. Save to `docs/DEPLOYMENT.md`.

### 6.1 Infrastructure
- **Hosting**: AWS (discuss — ECS Fargate vs EKS vs plain EC2 vs Lambda-heavy?).
- **Database**: RDS PostgreSQL (or Supabase for faster start?).
- **CDN**: CloudFront for frontend and static assets.
- **DNS**: Route 53.
- **SSL**: ACM.

### 6.2 CI/CD Pipeline
- GitHub Actions workflow:
  - On PR: lint, test, build.
  - On merge to main: build Docker image, push to ECR, deploy to staging.
  - Manual promotion to production.
- Database migration as part of deploy pipeline.
- Rollback strategy.

### 6.3 Feature Flags
- Simple feature flag system (DB-backed, admin-controllable).
- Flags for: premium features, AI providers, payment gateway, new UI features.
- Admin dashboard to toggle flags without redeployment.

### 6.4 Release Dashboard
- A simple admin view showing:
  - Current deployed version (git SHA, build time).
  - Active feature flags and their states.
  - Recent deployments log.
  - Quick toggle for feature flags.
  - Health check status of all services.

### 6.5 Environment Strategy
- `local` → `staging` → `production`.
- Environment-specific configs (12-factor app).
- Staging mirrors production topology at smaller scale.

---

## PHASE 7: ADMIN PANEL

Design the admin panel. Save to `docs/ADMIN-PANEL.md`.

### Must include:

#### 7.1 User Management
- User list with search, filter (free/premium/suspended).
- View user details, subscription status, credit balance.
- Suspend/unsuspend users. Issue credit refunds.
- Impersonate user (for debugging, with audit log).

#### 7.2 Company Management
- CRUD for companies (add, edit, remove).
- Moderation queue for community-submitted edits.
- Bulk import/export (CSV).
- AI-refresh trigger (re-enrich company data).

#### 7.3 Financial Dashboard
- Revenue overview: total, by product, by time period.
- Transaction log with filters.
- Refund management.
- Cost tracking: AI API costs per operation, total monthly spend.

#### 7.4 Feature & Platform Management
- Feature flag controls.
- Rate limit configuration.
- AI provider selection (switch between Claude/OpenAI).
- System health dashboard.

#### 7.5 Content Management
- Manage seed data / static content.
- Announcement banner control.
- FAQ / help content editor (stretch).

---

## PHASE 8: METRICS & MONITORING

Design observability. Save to `docs/MONITORING.md`.

### 8.1 Application Metrics
- Request latency (p50, p95, p99) by endpoint.
- Error rates by endpoint and error type.
- Active users (DAU, WAU, MAU).
- Feature usage: ATS checks/day, resumes uploaded/day, lists created.
- AI service metrics: latency, token usage, cost per request, failure rate.
- Payment metrics: conversion rate, revenue, failed payments.

### 8.2 Infrastructure Metrics
- CPU, memory, disk for all services.
- Database: connections, query latency, slow queries.
- Redis: memory usage, hit rate.
- S3: storage size, request count.

### 8.3 Alerting
- Critical: Service down, error rate spike, payment webhook failures.
- Warning: High AI costs, approaching rate limits, slow queries.
- Info: New user signup, premium conversion.

### 8.4 Logging
- Structured logging (JSON) with correlation IDs.
- Log levels: DEBUG (local only), INFO, WARN, ERROR.
- Centralised log aggregation (CloudWatch or self-hosted).

### 8.5 Tools
- Discuss: Prometheus + Grafana (self-hosted) vs CloudWatch vs Datadog (expensive) vs Grafana Cloud (free tier).
- Sentry for error tracking.
- Uptime monitoring (UptimeRobot or similar).

---

## PHASE 9: BUILD PLAN

**Only after Phases 1-8 are discussed and documented**, create the implementation plan. Save to `docs/BUILD-PLAN.md`.

### Structure the plan as:

#### Sprint 0 — Foundation (Week 1-2)
- Repo setup, CI/CD, Docker Compose, Makefile.
- Database schema + migrations.
- Backend skeleton: project structure, middleware, config, health check.
- Frontend skeleton: Next.js setup, routing, layout, design system.
- Auth system (signup, login, sessions).

#### Sprint 1 — Company Directory (Week 3-4)
- Company CRUD APIs.
- Seed data (initial 50-100 companies).
- Frontend: Browse, search, filter, company profile pages.
- Client-side caching.
- Admin: Company management.

#### Sprint 2 — User Lists & Tracking (Week 5-6)
- List CRUD APIs.
- Application tracking within lists.
- Frontend: List management UI, status tracking, dashboard.
- Free user dashboard.

#### Sprint 3 — Payment & Premium Foundation (Week 7-8)
- Razorpay integration.
- Credit system.
- Resume upload + storage.
- AI service foundation (provider abstraction, resume parsing).
- Premium gating middleware.

#### Sprint 4 — AI Features (Week 9-10)
- ATS scoring (all 3 tiers).
- AI-curated list generation.
- Premium dashboard.
- Usage tracking.

#### Sprint 5 — Admin & Polish (Week 11-12)
- Admin panel.
- Feature flags.
- Monitoring setup.
- Performance optimisation.
- Security audit.
- Bug fixes and UX polish.

#### Sprint 6 — Launch Prep (Week 13)
- Staging environment.
- Load testing.
- Documentation.
- Beta testing with select users.
- Production deployment.

---

## HOW TO PROCEED

Start with **Phase 1: Requirements Refinement**. For each phase:

1. Read the section above.
2. Present your analysis, recommendations, and any questions.
3. Wait for my decisions.
4. Document the finalised output in the appropriate `docs/` file.
5. Only then move to the next phase.

**Begin now with Phase 1.** Review the requirements I've outlined, identify gaps or issues, propose refinements, and prepare to write `docs/PRD.md` once we align.

---

## APPENDIX: TECHNOLOGY PREFERENCES

For reference, here are my preferences (open to discussion):

| Area | Preference | Open to |
|------|-----------|---------|
| Backend | Go (Chi or Echo) | Gin if compelling reason |
| Frontend | Next.js (App Router) | React SPA if SSR not needed |
| Database | PostgreSQL | Supabase if it accelerates MVP |
| Cache | Redis | — |
| Search | MeiliSearch | Typesense, Postgres FTS |
| AI | Claude API (primary) | OpenAI fallback |
| Payments | Razorpay | — |
| Hosting | AWS (ECS Fargate) | Fly.io, Railway for MVP |
| CI/CD | GitHub Actions | — |
| IaC | Terraform | Pulumi |
| Monitoring | Grafana Cloud free tier | CloudWatch |
| Error Tracking | Sentry | — |
| Email | Resend | AWS SES |

---

## APPENDIX: WHAT "DONE" LOOKS LIKE

The project is shippable when:
1. A user can browse companies, sign up, create lists, and track applications (free).
2. A user can pay, upload resumes, get ATS scores, and see AI-curated suggestions (premium).
3. An admin can manage companies, users, features, and see financials.
4. The system is deployed, monitored, and can handle ~1000 concurrent users.
5. All design documents are complete and up-to-date in `docs/`.
