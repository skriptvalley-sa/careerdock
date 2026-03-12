# CareerDock — Admin Panel Design

> **Version:** 1.0
> **Status:** Draft (Phase 7)
> **Last updated:** 2026-03-12
> **Depends on:** [PRD.md](./PRD.md), [LLD/api.md](./LLD/api.md), [LLD/database.md](./LLD/database.md), [LLD/frontend.md](./LLD/frontend.md), [LLD/payments.md](./LLD/payments.md)

---

## 1. Overview

The admin panel is integrated into the main Next.js frontend under `/admin/*` routes. It shares the same component library, auth system, and API client as the user-facing app. Access is controlled by role-based middleware on both backend and frontend.

### 1.1 Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Separate app vs integrated | Integrated (`/admin/*` routes) | Shares components, auth, API client. No duplicate code |
| Admin privilege levels | Single `admin` role | Solo founder — granular RBAC is overkill for MVP |
| User impersonation | Deferred to v2 | Security complexity (session isolation, privilege escalation). Admin can view user details for debugging |
| Bulk operations (CSV import/export) | Deferred to v2 | Manual CRUD sufficient at 200-500 companies |
| Audit log retention | Indefinite | Tiny data footprint (~1KB per entry). No purge needed |

### 1.2 Access Control

| Role | Admin Access | Capabilities |
|------|:---:|-------------|
| `user` | No | — |
| `moderator` | Partial | Submit company edits (moderation queue only, via `/api/companies/:id/edits`) |
| `admin` | Full | All admin features |

Both backend middleware (`auth.RequireAdmin`) and frontend route guards (`(admin)` layout) enforce access.

---

## 2. Admin Layout

### 2.1 Navigation Structure

```
┌─────────────────────────────────────────────────────────┐
│  CareerDock Admin                        [User ▼] [Exit]│
├──────────────┬──────────────────────────────────────────┤
│              │                                          │
│  Dashboard   │  ┌──────────────────────────────────┐   │
│              │  │                                    │   │
│  Users       │  │         Content Area               │   │
│              │  │                                    │   │
│  Companies   │  │   (page-specific content)          │   │
│              │  │                                    │   │
│  Moderation  │  │                                    │   │
│              │  │                                    │   │
│  Payments    │  │                                    │   │
│              │  │                                    │   │
│  AI Costs    │  │                                    │   │
│              │  │                                    │   │
│  Features    │  │                                    │   │
│              │  │                                    │   │
│  Audit Log   │  │                                    │   │
│              │  └──────────────────────────────────┘   │
└──────────────┴──────────────────────────────────────────┘
```

- **Fixed left sidebar** with navigation links.
- **Top bar** with admin name, "Exit Admin" link (returns to user dashboard).
- **Content area** renders the active page.
- Sidebar highlights active section. Badge counts on Moderation (pending edits count).

### 2.2 Route Map

| Route | Page | Description |
|-------|------|-------------|
| `/admin` | Dashboard | Overview stats, revenue chart, system health |
| `/admin/users` | User List | Search, filter, paginate users |
| `/admin/users/[id]` | User Detail | Profile, credits, activity, actions |
| `/admin/companies` | Company List | Search, filter, CRUD |
| `/admin/companies/new` | Create Company | Full company creation form |
| `/admin/companies/[id]/edit` | Edit Company | Edit existing company, trigger AI enrich |
| `/admin/moderation` | Moderation Queue | Pending community edits |
| `/admin/moderation/[id]` | Edit Review | Side-by-side diff, approve/reject |
| `/admin/payments` | Payment List | Transaction log with filters |
| `/admin/ai` | AI Cost Dashboard | Cost breakdown by operation and period |
| `/admin/features` | Feature Flags | Toggle list with create/edit |
| `/admin/audit-log` | Audit Log | Filterable admin action history |

---

## 3. Dashboard (`/admin`)

### 3.1 API Endpoint

**`GET /api/admin/dashboard`** — returns aggregated platform stats.

```json
{
  "data": {
    "users": {
      "total": 342,
      "premium": 45,
      "new_today": 3,
      "new_this_week": 18
    },
    "revenue": {
      "total_paise": 1795500,
      "this_month_paise": 239400,
      "today_paise": 39900
    },
    "companies": {
      "total": 287,
      "pending_edits": 4
    },
    "ai": {
      "operations_today": 23,
      "estimated_cost_today_paise": 2760
    },
    "system": {
      "api_version": "v1.2.0",
      "worker_queue_depth": 2,
      "failed_jobs_24h": 0
    }
  }
}
```

### 3.2 Dashboard Cards

| Card | Metrics | Data Source |
|------|---------|------------|
| **Users** | Total, premium count, new today/this week | `COUNT(*)` on `users` table |
| **Revenue** | Total, this month, today (₹ display) | `SUM(amount_paise)` on `payments WHERE status = 'captured'` |
| **Companies** | Total count, pending edits badge | `COUNT(*)` on `companies` + `company_edits WHERE status = 'pending'` |
| **AI Costs** | Operations today, estimated cost today (₹) | Token counts from `ats_checks` and `resumes` JSONB fields |
| **System** | API version, queue depth, failed jobs | Health endpoint + Asynq inspector |

### 3.3 Revenue Chart

A line chart showing daily revenue for the last 30 days. Data from:

```sql
SELECT DATE(created_at) AS day, SUM(amount_paise) AS revenue
FROM payments
WHERE status = 'captured' AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY day;
```

Rendered with a lightweight chart library (e.g., Recharts or Chart.js).

---

## 4. User Management (`/admin/users`)

### 4.1 User List

**`GET /api/admin/users`** — offset-based pagination.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | string | — | Search by name or email (ILIKE) |
| `role` | string | — | Filter: `user`, `moderator`, `admin` |
| `premium` | boolean | — | Filter: premium users only |
| `suspended` | boolean | — | Filter: show only suspended (deleted_at IS NOT NULL) |
| `sort` | string | `created_at` | Sort by: `created_at`, `name`, `email` |
| `order` | string | `desc` | Sort order: `asc`, `desc` |
| `page` | int | 1 | Page number |
| `per_page` | int | 50 | Items per page (max 100) |

**Table columns:** Name, Email, Role, Premium (Yes/No), Resumes, Lists, Joined, Status, Actions.

**Actions:** View details, Suspend/Unsuspend.

### 4.2 User Detail (`/admin/users/[id]`)

**`GET /api/admin/users/:id`**

Shows:
- **Profile section:** Name, email, role, premium status, joined date, last login.
- **Credits section:** Table of credit types with current balance.
- **Activity section:** Resume count, list count, ATS check count, total payment amount.
- **Payments section:** Recent payments for this user (last 10).
- **Action buttons:** Suspend/Unsuspend, Change Role dropdown.

### 4.3 User Actions

| Action | Endpoint | Side Effects |
|--------|----------|-------------|
| Suspend user | `PUT /api/admin/users/:id/suspend` | Sets `deleted_at = NOW()`. Invalidates all sessions in Redis. User sees "Account suspended" on next request. Triggers 30-day hard-delete countdown |
| Unsuspend user | `PUT /api/admin/users/:id/unsuspend` | Clears `deleted_at`. User can log in again. Hard-delete countdown cancelled |
| Change role | `PUT /api/admin/users/:id/role` | Updates `role` field. New access token will reflect new role on next refresh |

All actions create an `admin_audit_log` entry.

### 4.4 User Hard Delete (Automated)

A scheduled Asynq periodic task runs daily:

```go
// Finds users where deleted_at < NOW() - 30 days
// For each user:
//   1. Delete resume PDFs from S3
//   2. Hard delete all related rows (lists, entries, resumes, ats_checks, credits, transactions, notifications)
//   3. Hard delete user row
//   4. Log in admin_audit_log (admin_id = system)
```

This is automated — no admin UI needed. Admin can see upcoming hard deletes by filtering users where `deleted_at IS NOT NULL`.

---

## 5. Company Management (`/admin/companies`)

### 5.1 Company List

**`GET /api/admin/companies`** — uses the same public company list endpoint but with additional admin filters.

| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Search by name (uses Postgres FTS) |
| `verified` | boolean | Filter: verified only |
| `sort` | string | `name`, `created_at`, `last_verified_at` |
| `order` | string | `asc`, `desc` |
| `page` | int | Page number |
| `per_page` | int | Items per page (max 100) |

**Table columns:** Name, Slug, Size, Headquarters, Tech Stack (tags), Verified, Last Verified, Actions.

**Actions:** Edit, Delete, Enrich (AI), Refresh (AI).

### 5.2 Create Company (`/admin/companies/new`)

**`POST /api/admin/companies`**

Form fields:

| Field | Type | Required | Validation |
|-------|------|:---:|-----------|
| `name` | text | Yes | Max 200 chars |
| `slug` | text | Auto-generated | Lowercase, alphanumeric + hyphens. Editable but must be unique |
| `website` | URL | No | Valid URL |
| `description` | textarea | No | Max 2000 chars |
| `logo_url` | URL | No | Valid URL (or upload to S3 logos bucket) |
| `size` | select | No | startup, small, medium, large, enterprise |
| `founded` | number | No | 4-digit year |
| `headquarters` | text | No | Max 100 chars |
| `hiring_active` | boolean | No | Default false |
| `careers_page_url` | URL | No | Valid URL |
| `tech_stack` | tag input | No | Array of strings |
| `domains` | tag input | No | Array of strings |
| `glassdoor_rating` | number | No | 0.0-5.0 |
| `interview_patterns` | JSON editor | No | JSONB — rounds, difficulty, typical_duration |
| `compensation_tier` | select | No | below_market, market, above_market, top_of_market |
| `has_rsu` | boolean | No | Default false |
| `compensation_bands` | JSON editor | No | JSONB — by role/level |

Auto-generates slug from name on blur (e.g., "Tata Consultancy Services" → `tata-consultancy-services`). Admin can override.

### 5.3 Edit Company (`/admin/companies/[id]/edit`)

**`PUT /api/admin/companies/:id`** — partial update (only changed fields sent).

Same form as create, pre-populated with current values. Additional actions:

| Button | Endpoint | Effect |
|--------|----------|--------|
| **Save** | `PUT /api/admin/companies/:id` | Updates company. Invalidates Redis ATS cache for this company |
| **AI Enrich** | `POST /api/admin/companies/:id/enrich` | Queues enrichment job. Returns 202 with job ID. Fills missing fields (tech stack, interview patterns, compensation) using LLM |
| **AI Refresh** | `POST /api/admin/companies/:id/refresh` | Queues full refresh job. Re-researches all fields from public sources. Returns 202 with job ID |
| **Delete** | `DELETE /api/admin/companies/:id` | Confirmation modal. Cascades to `list_entries`. Creates audit log |

### 5.4 AI Enrichment Feedback

After triggering AI Enrich or Refresh, show a status indicator:

```
[AI Enrich] button clicked
  → "Enrichment queued. Job ID: abc-123"
  → Poll /api/admin/companies/:id every 5s
  → When enrichment completes, reload form with new data
  → Highlight fields that were updated (green border)
```

No SSE needed for admin — simple polling is sufficient for rare AI operations.

---

## 6. Moderation Queue (`/admin/moderation`)

### 6.1 Queue List

**`GET /api/admin/moderation/edits`** — offset-based pagination.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | `pending` | Filter: `pending`, `approved`, `rejected`, `all` |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page |

**Table columns:** Company, Submitted By, Fields Changed, Submitted At, Status, Actions.

**Actions:** Review (opens detail page).

Badge count of pending edits shown in sidebar navigation.

### 6.2 Edit Review (`/admin/moderation/[id]`)

**`GET /api/admin/moderation/edits/:id`**

Shows a **side-by-side diff view**:

```
┌─────────────────────────────┬────────────────────────────┐
│ Current Value               │ Proposed Change             │
├─────────────────────────────┼────────────────────────────┤
│ tech_stack:                 │ tech_stack:                 │
│   [React, Node.js, Python]  │   [React, Node.js, Go,     │
│                             │    Python, PostgreSQL]       │
│                             │                             │
│ compensation_tier:          │ compensation_tier:          │
│   market                    │   above_market              │
│                             │                             │
│ interview_patterns:         │ interview_patterns:         │
│   { rounds: 3 }            │   { rounds: 4, typical_    │
│                             │     duration: "3 weeks" }   │
├─────────────────────────────┴────────────────────────────┤
│ Submitted by: jane@example.com on 2026-03-10             │
│ Notes: "Added Go and PostgreSQL to stack, updated        │
│         interview info from recent experience"            │
├──────────────────────────────────────────────────────────┤
│ Review Notes: [________________________]                  │
│                                                          │
│              [Approve]           [Reject]                 │
└──────────────────────────────────────────────────────────┘
```

**Actions:**

| Action | Endpoint | Effect |
|--------|----------|--------|
| Approve | `POST /api/admin/moderation/edits/:id/approve` | Merges `changes` JSONB into company record. Sets edit status = `approved`, `reviewed_by` = admin ID. Creates audit log |
| Reject | `POST /api/admin/moderation/edits/:id/reject` | Sets edit status = `rejected`, `reviewed_by` = admin ID, stores `review_notes`. Creates audit log |

Both require `review_notes` (optional for approve, required for reject — explain why).

### 6.3 Moderator Permissions

Moderators can **only** submit edit suggestions via the public company profile page (a "Suggest Edit" button visible when `role IN ('moderator', 'admin')`). They cannot:
- Directly edit companies.
- Approve/reject other edits.
- Access any `/admin/*` pages.

Only admins see the moderation queue and can review edits.

---

## 7. Payment Management (`/admin/payments`)

### 7.1 Payment List

**`GET /api/admin/payments`** — offset-based pagination.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | — | Filter: `created`, `captured`, `failed`, `refunded` |
| `product_type` | string | — | Filter: `starter_pack`, `resume_upload`, `ats_bundle`, `rebuy_pack` |
| `user_id` | UUID | — | Filter: specific user |
| `from` | date | — | Filter: created after this date |
| `to` | date | — | Filter: created before this date |
| `page` | int | 1 | Page number |
| `per_page` | int | 50 | Items per page |

**Table columns:** Date, User (name + email), Product, Amount (₹), Status (badge), Receipt #, Actions.

**Status badges:** Created (grey), Captured (green), Failed (red), Refunded (orange).

### 7.2 Refund Flow

**`POST /api/admin/payments/:id/refund`**

**Pre-conditions (all must be true):**
1. Payment status is `captured`.
2. Payment is within 7 days of `created_at`.
3. No credits from this payment have been consumed.

**Request:**
```json
{
  "reason": "Customer requested refund within 7-day window"
}
```

**Server-side logic (in a single transaction):**
1. Verify pre-conditions.
2. Deduct allocated credits from `user_credits` (reverse the allocation from [payments.md §2.1](./LLD/payments.md)).
3. If this was a `starter_pack` and no other captured payments exist → clear `premium_since`.
4. Set `payments.status = 'refunded'`, `refund_reason`, `refunded_at = NOW()`, `refunded_by = admin_id`.
5. Log `credit_transactions` entries (negative amounts) for each reversed credit type.
6. Queue async job: call Razorpay Refund API (outside transaction — idempotent).
7. Create `admin_audit_log` entry.

**UI flow:**
1. Admin clicks "Refund" button on payment row.
2. Confirmation modal shows: amount, user, product, credits to reverse.
3. Admin enters reason (required).
4. On confirm: calls API, shows success/error toast.

### 7.3 Reconciliation View

For payments stuck in `created` status (webhook may have been missed):

```sql
SELECT * FROM payments
WHERE status = 'created' AND created_at < NOW() - INTERVAL '30 minutes'
ORDER BY created_at DESC;
```

**Displayed as a warning section** at the top of the payments page when count > 0:

```
⚠️ 2 payments in 'created' status for 30+ minutes (possible missed webhooks)
[View] → shows these payments with a "Verify on Razorpay" link (opens Razorpay dashboard)
```

Manual reconciliation: admin checks Razorpay dashboard, and if payment was captured there, triggers a manual webhook replay or updates the status directly. This is a rare edge case — documented for completeness.

---

## 8. AI Cost Dashboard (`/admin/ai`)

### 8.1 API Endpoint

**`GET /api/admin/ai/costs`**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `from` | date | 30 days ago | Start date |
| `to` | date | today | End date |
| `group_by` | string | `day` | Grouping: `day`, `week`, `month` |

**Response:**
```json
{
  "data": {
    "total_operations": 456,
    "total_estimated_cost_paise": 54720,
    "by_operation": {
      "resume_parse":   { "count": 120, "cost_paise": 6000,  "avg_tokens": 3000 },
      "ats_general":    { "count": 120, "cost_paise": 9600,  "avg_tokens": 4000 },
      "ats_company":    { "count": 89,  "cost_paise": 8900,  "avg_tokens": 5000 },
      "ats_job":        { "count": 67,  "cost_paise": 8040,  "avg_tokens": 6000 },
      "curated_list":   { "count": 35,  "cost_paise": 7000,  "avg_tokens": 10000 },
      "company_enrich": { "count": 25,  "cost_paise": 1500,  "avg_tokens": 3000 }
    },
    "by_period": [
      { "date": "2026-03-01", "operations": 18, "cost_paise": 2160 },
      { "date": "2026-03-02", "operations": 22, "cost_paise": 2640 }
    ],
    "cache_stats": {
      "total_requests": 612,
      "cache_hits": 156,
      "hit_rate_percent": 25.5
    }
  }
}
```

### 8.2 Cost Calculation

Costs are estimated from token counts stored in JSONB result fields:

```go
// internal/service/admin_service.go

// Token prices (per 1M tokens) — update when providers change pricing
var tokenPrices = map[string]struct{ Input, Output float64 }{
    "claude-sonnet": {Input: 3.00, Output: 15.00},  // USD per 1M tokens
    "gpt-4o-mini":   {Input: 0.15, Output: 0.60},
}

// Convert USD cost to paise: cost_usd * 84 (INR/USD) * 100 (paise)
func estimateCostPaise(inputTokens, outputTokens int, provider string) int {
    prices := tokenPrices[provider]
    costUSD := (float64(inputTokens) * prices.Input / 1_000_000) +
               (float64(outputTokens) * prices.Output / 1_000_000)
    return int(math.Round(costUSD * 84 * 100))
}
```

Token counts are extracted from:
- `resumes.parsed_data → tokens_used` (resume parse + general ATS)
- `ats_checks.result → tokens_used` (company/job ATS)
- `curated_lists.result → tokens_used` (curated lists)

### 8.3 Dashboard Layout

| Component | Content |
|-----------|---------|
| **Total cost card** | Total ₹ for selected period |
| **Operations card** | Total operations count |
| **Cache hit rate card** | Percentage of requests served from cache |
| **Cost trend chart** | Line chart — daily/weekly/monthly cost over time |
| **Breakdown table** | Operation type, count, total cost, avg tokens, avg cost per operation |

### 8.4 Cost Alerts

Simple threshold-based alert (checked by a daily Asynq periodic task):

| Condition | Action |
|-----------|--------|
| Daily AI cost > ₹500 | Log warning + send email to admin |
| Daily AI cost > ₹1,000 | Log error + send email + consider auto-disabling AI operations via feature flag |

Thresholds configurable via feature flags:
- `ai_cost_warn_threshold_paise` (default: 50000)
- `ai_cost_critical_threshold_paise` (default: 100000)

---

## 9. Feature Flags (`/admin/features`)

### 9.1 Feature Flag UI

A simple table with toggle switches:

```
┌──────────────────────────┬─────────┬───────────────────────────────┬────────────┐
│ Flag Key                 │ Enabled │ Description                   │ Updated    │
├──────────────────────────┼─────────┼───────────────────────────────┼────────────┤
│ premium_features         │   ✅    │ Enable premium tier features   │ 2026-03-01 │
│ ai_curated_lists         │   ✅    │ Enable AI curated list gen     │ 2026-03-01 │
│ cv_generation            │   ❌    │ Enable CV generation (v2)      │ 2026-03-05 │
│ ai_provider_openai_only  │   ❌    │ Use OpenAI only (skip Claude)  │ 2026-03-10 │
│ maintenance_mode         │   ❌    │ Show maintenance page to users  │ 2026-03-01 │
├──────────────────────────┴─────────┴───────────────────────────────┴────────────┤
│                                              [+ Create New Flag]                │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 9.2 Flag Operations

| Action | Endpoint | Notes |
|--------|----------|-------|
| List all | `GET /api/admin/feature-flags` | Returns all flags |
| Create | `POST /api/admin/feature-flags` | Requires: `key` (unique), `enabled`, `description` |
| Toggle | `PUT /api/admin/feature-flags/:key` | Toggles `enabled`. Invalidates Redis cache (5-min TTL). Creates audit log |

### 9.3 Backend Usage Pattern

```go
// internal/service/feature_flags.go

type FeatureFlagService struct {
    repo  domain.FeatureFlagRepository
    cache *redis.Client
}

func (s *FeatureFlagService) IsEnabled(ctx context.Context, key string) bool {
    // 1. Check Redis cache
    cacheKey := fmt.Sprintf("ff:%s", key)
    if val, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
        return val == "1"
    }

    // 2. Cache miss — read from DB
    flag, err := s.repo.GetByKey(ctx, key)
    if err != nil || flag == nil {
        return false // Unknown flag defaults to disabled
    }

    // 3. Cache result (5-min TTL)
    s.cache.Set(ctx, cacheKey, boolToStr(flag.Enabled), 5*time.Minute)
    return flag.Enabled
}
```

### 9.4 Initial Flags (Seeded in Sprint 0)

| Key | Default | Purpose |
|-----|---------|---------|
| `premium_features` | `true` | Master switch for all premium features |
| `ai_curated_lists` | `true` | Enable/disable curated list generation |
| `cv_generation` | `false` | CV generation feature (v2, not implemented) |
| `ai_provider_openai_only` | `false` | Force OpenAI-only (skip Claude) for cost control |
| `maintenance_mode` | `false` | Show maintenance page, block all non-admin requests |
| `registration_enabled` | `true` | Allow new user registration |

---

## 10. Audit Log (`/admin/audit-log`)

### 10.1 Log List

**`GET /api/admin/audit-log`** — offset-based pagination.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `admin_id` | UUID | — | Filter by admin who performed the action |
| `entity_type` | string | — | Filter: `user`, `company`, `payment`, `feature_flag`, `moderation` |
| `entity_id` | UUID | — | Filter: specific entity |
| `action` | string | — | Filter: specific action (e.g., `user_suspended`) |
| `from` | date | — | Filter: after this date |
| `to` | date | — | Filter: before this date |
| `page` | int | 1 | Page number |
| `per_page` | int | 50 | Items per page |

**Table columns:** Timestamp, Admin, Action, Entity Type, Entity, Details (expandable), IP Address.

### 10.2 Logged Actions

| Action | Entity Type | Details (JSONB) |
|--------|------------|-----------------|
| `user_suspended` | `user` | `{ user_email }` |
| `user_unsuspended` | `user` | `{ user_email }` |
| `user_role_changed` | `user` | `{ user_email, from_role, to_role }` |
| `user_hard_deleted` | `user` | `{ user_email, resume_count, list_count }` |
| `company_created` | `company` | `{ name, slug }` |
| `company_updated` | `company` | `{ changed_fields: [...] }` |
| `company_deleted` | `company` | `{ name, slug }` |
| `company_enrich_triggered` | `company` | `{ name, job_id }` |
| `company_refresh_triggered` | `company` | `{ name, job_id }` |
| `edit_approved` | `moderation` | `{ company_name, submitted_by, changed_fields }` |
| `edit_rejected` | `moderation` | `{ company_name, submitted_by, review_notes }` |
| `refund_issued` | `payment` | `{ user_email, amount_paise, product_type, reason }` |
| `feature_flag_toggled` | `feature_flag` | `{ key, from_enabled, to_enabled }` |
| `feature_flag_created` | `feature_flag` | `{ key, enabled, description }` |

### 10.3 Audit Log Implementation

```go
// internal/service/audit.go

type AuditEntry struct {
    AdminID    uuid.UUID
    Action     string
    EntityType string
    EntityID   uuid.UUID
    Details    map[string]any
    IPAddress  string
}

func (s *AuditService) Log(ctx context.Context, entry AuditEntry) error {
    return s.repo.Create(ctx, &domain.AuditLog{
        ID:         uuid.Must(uuid.NewV7()),
        AdminID:    entry.AdminID,
        Action:     entry.Action,
        EntityType: entry.EntityType,
        EntityID:   entry.EntityID,
        Details:    entry.Details,
        IPAddress:  entry.IPAddress,
        CreatedAt:  time.Now(),
    })
}
```

The admin's IP is extracted from `X-Real-IP` header (set by Nginx) in the middleware and stored in context.

### 10.4 Retention

Audit logs are kept **indefinitely**. At MVP scale (a few hundred entries per month), this is negligible storage. No purge policy needed.

---

## 11. Admin API Summary

Complete list of admin-only endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/admin/dashboard` | Platform overview stats |
| `GET` | `/api/admin/users` | List users (paginated, filterable) |
| `GET` | `/api/admin/users/:id` | User detail with credits and activity |
| `PUT` | `/api/admin/users/:id/suspend` | Suspend user |
| `PUT` | `/api/admin/users/:id/unsuspend` | Unsuspend user |
| `PUT` | `/api/admin/users/:id/role` | Change user role |
| `POST` | `/api/admin/companies` | Create company |
| `PUT` | `/api/admin/companies/:id` | Update company |
| `DELETE` | `/api/admin/companies/:id` | Delete company |
| `POST` | `/api/admin/companies/:id/enrich` | Trigger AI enrichment |
| `POST` | `/api/admin/companies/:id/refresh` | Trigger AI refresh |
| `GET` | `/api/admin/moderation/edits` | List moderation queue |
| `GET` | `/api/admin/moderation/edits/:id` | Edit detail with diff |
| `POST` | `/api/admin/moderation/edits/:id/approve` | Approve edit |
| `POST` | `/api/admin/moderation/edits/:id/reject` | Reject edit |
| `GET` | `/api/admin/payments` | List payments (paginated, filterable) |
| `POST` | `/api/admin/payments/:id/refund` | Issue refund |
| `GET` | `/api/admin/ai/costs` | AI cost breakdown |
| `GET` | `/api/admin/feature-flags` | List feature flags |
| `POST` | `/api/admin/feature-flags` | Create feature flag |
| `PUT` | `/api/admin/feature-flags/:key` | Toggle feature flag |
| `GET` | `/api/admin/audit-log` | Admin action history |

All endpoints require `admin` role. All state-changing endpoints create an audit log entry.

---

## 12. Admin Components (Frontend)

### 12.1 Shared Components

| Component | Used In | Description |
|-----------|---------|-------------|
| `AdminLayout` | All admin pages | Sidebar + top bar wrapper |
| `StatsCard` | Dashboard | Metric card with label, value, optional trend |
| `DataTable` | Users, Companies, Payments, Audit Log | Sortable, filterable table with pagination |
| `ConfirmModal` | Suspend, Delete, Refund | Confirmation dialog with reason input |
| `DiffView` | Moderation review | Side-by-side JSON diff |
| `FlagToggle` | Feature flags | Toggle switch with instant save |
| `StatusBadge` | Payments, Moderation | Coloured badge (green/red/orange/grey) |
| `TagInput` | Company form | Multi-value input for tech_stack, domains |
| `JsonEditor` | Company form | Simple JSON editor for interview_patterns, compensation_bands |
| `LineChart` | Dashboard, AI costs | Recharts or Chart.js line chart wrapper |

### 12.2 Data Fetching Pattern

All admin pages use TanStack Query with admin-specific query keys:

```typescript
// lib/query-keys.ts (additions)

export const queryKeys = {
  // ... existing keys ...
  admin: {
    dashboard:    ['admin', 'dashboard'] as const,
    users:        (filters: UserFilters) => ['admin', 'users', filters] as const,
    userDetail:   (id: string) => ['admin', 'users', 'detail', id] as const,
    companies:    (filters: CompanyFilters) => ['admin', 'companies', filters] as const,
    moderation:   (status: string) => ['admin', 'moderation', status] as const,
    editDetail:   (id: string) => ['admin', 'moderation', 'detail', id] as const,
    payments:     (filters: PaymentFilters) => ['admin', 'payments', filters] as const,
    aiCosts:      (params: AICostParams) => ['admin', 'ai', 'costs', params] as const,
    featureFlags: ['admin', 'feature-flags'] as const,
    auditLog:     (filters: AuditFilters) => ['admin', 'audit-log', filters] as const,
  },
} as const;
```

Mutations (suspend, refund, toggle flag) use TanStack Query mutations with optimistic updates where appropriate (e.g., flag toggle).

---

## 13. Deferred to v2

| Feature | Reason |
|---------|--------|
| User impersonation | Security complexity — session isolation, privilege escalation prevention. Admin can view user details for now |
| Bulk CSV import/export | Manual CRUD sufficient at 200-500 companies. Add when company count grows |
| Granular admin permissions | Single `admin` role is enough for solo founder. Add when team grows |
| Admin MFA | Stretch goal. Adds TOTP/WebAuthn complexity. Rely on strong password + session security for now |
| Advanced analytics | DAU/WAU/MAU, conversion funnels, cohort analysis. Use external tools (PostHog, Mixpanel) if needed |
| Admin notification preferences | Email alerts hardcoded for now. Configurable preferences when there are multiple admins |
| Batch moderation actions | Review one-by-one for now. Batch approve when edit volume grows |
| Company profile versioning | Current diff view is sufficient. Full version history when edit volume grows |
