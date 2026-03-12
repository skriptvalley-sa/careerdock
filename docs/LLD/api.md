# CareerDock — API Design (LLD)

> **Version:** 1.0
> **Status:** Draft (Phase 3)
> **Last updated:** 2026-03-11
> **Depends on:** [PRD.md](../PRD.md), [ARCHITECTURE.md](../ARCHITECTURE.md), [database.md](./database.md)

---

## 1. Conventions

### 1.1 Base URL

```
Production:  https://api.careerdock.skriptvalley.com
Local dev:   http://localhost:8080
```

No API versioning for MVP. All routes under `/api/`.

### 1.2 Authentication

- JWT access token (15-min TTL) + refresh token (7-day TTL) in **httpOnly, Secure, SameSite=Strict** cookies.
- No `Authorization` header — middleware reads cookies automatically.
- Endpoints marked with auth level: `public`, `authenticated`, `premium`, `moderator`, `admin`.
- `premium` = `authenticated` + `premium_since IS NOT NULL`.
- `moderator` = `authenticated` + `role IN ('moderator', 'admin')`.
- `admin` = `authenticated` + `role = 'admin'`.

### 1.3 Request Format

- `Content-Type: application/json` for all JSON endpoints.
- `Content-Type: multipart/form-data` for file uploads (resume).
- All request bodies are validated. Missing required fields or invalid types return `422`.

### 1.4 Response Format

**Success:**

```json
{
  "data": { ... }
}
```

**Success with pagination:**

```json
{
  "data": [ ... ],
  "pagination": {
    "next_cursor": "eyJpZCI6...",
    "has_more": true
  }
}
```

**Error:**

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": { ... }
  }
}
```

### 1.5 Error Codes

| HTTP Status | Error Code | When |
|-------------|-----------|------|
| 400 | `BAD_REQUEST` | Malformed request body |
| 401 | `UNAUTHORIZED` | Missing or expired token |
| 403 | `FORBIDDEN` | Insufficient role/permissions |
| 404 | `NOT_FOUND` | Resource doesn't exist |
| 409 | `CONFLICT` | Duplicate resource (email, slug) |
| 422 | `VALIDATION_ERROR` | Invalid field values |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | AI provider down, DB unreachable |

### 1.6 Pagination

**Cursor-based** for all public and user-facing list endpoints. Cursor is an opaque base64-encoded string (encodes the last item's `id` or `created_at`).

```
GET /api/companies?cursor=eyJpZCI6...&limit=20
```

- `limit`: 1-100, default 20.
- `cursor`: omit for first page.
- Response includes `pagination.next_cursor` (null if no more results) and `pagination.has_more`.

**Offset-based** for admin endpoints only (need page jumping for tables):

```
GET /api/admin/users?page=2&per_page=50
```

- Response includes `pagination.total`, `pagination.page`, `pagination.per_page`, `pagination.total_pages`.

### 1.7 Rate Limiting

Rate limits enforced via Redis sliding window counters. Response headers:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1710000000
```

| Tier | Scope | Limit |
|------|-------|-------|
| Public | Per IP | 60 req/min |
| Authenticated | Per user | 120 req/min |
| AI operations | Per user | 10 req/min |
| Payment | Per user | 10 req/min |
| Auth (login/register) | Per IP | 10 req/min |
| Admin | Per user | 120 req/min |

### 1.8 Common Headers

**Request:**
- `X-Request-ID`: Optional client-generated correlation ID. If absent, server generates one.

**Response:**
- `X-Request-ID`: Correlation ID (echoed or generated).
- `X-RateLimit-*`: Rate limit info.
- `ETag` / `Last-Modified`: On company list/detail for client-side caching.
- `Cache-Control`: For public company endpoints.

---

## 2. Auth Endpoints

### POST `/api/auth/register`

Register a new user account.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | Auth tier (10/min per IP) |

**Request:**

```json
{
  "email": "user@example.com",
  "password": "SecureP4ss",
  "name": "Sujay Kumar"
}
```

**Validation:**
- `email`: valid email format, max 255 chars.
- `password`: min 8 chars, at least 1 uppercase, 1 lowercase, 1 digit.
- `name`: 1-255 chars.

**Response (201):**

```json
{
  "data": {
    "id": "01912345-...",
    "email": "user@example.com",
    "name": "Sujay Kumar",
    "role": "user",
    "email_verified": false,
    "created_at": "2026-03-11T10:00:00Z"
  }
}
```

**Side effects:**
- Sends verification email via Resend (async, queued via Asynq).
- Sets access + refresh token cookies.

**Errors:**
- `409 CONFLICT` — email already registered.

---

### POST `/api/auth/login`

Authenticate with email + password.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | Auth tier (10/min per IP) |

**Request:**

```json
{
  "email": "user@example.com",
  "password": "SecureP4ss"
}
```

**Response (200):**

```json
{
  "data": {
    "id": "01912345-...",
    "email": "user@example.com",
    "name": "Sujay Kumar",
    "role": "user",
    "email_verified": true,
    "premium_since": "2026-03-10T08:00:00Z",
    "created_at": "2026-03-01T10:00:00Z"
  }
}
```

**Side effects:**
- Sets access token cookie (15-min TTL) + refresh token cookie (7-day TTL).

**Errors:**
- `401 UNAUTHORIZED` — invalid email or password (generic message, don't reveal which).
- `403 FORBIDDEN` — account suspended (soft-deleted).

---

### POST `/api/auth/logout`

Invalidate refresh token and clear cookies.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Request:** Empty body.

**Response (204):** No content.

**Side effects:**
- Adds refresh token JTI to Redis blacklist (7-day TTL).
- Clears access + refresh token cookies.

---

### POST `/api/auth/refresh`

Exchange refresh token for new access + refresh token pair.

| Aspect | Detail |
|--------|--------|
| Auth | Refresh token cookie required |
| Rate limit | Auth tier (10/min per IP) |

**Request:** Empty body. Refresh token read from cookie.

**Response (200):**

```json
{
  "data": {
    "message": "Tokens refreshed"
  }
}
```

**Side effects:**
- Old refresh token blacklisted in Redis.
- New access + refresh token cookies set (rotation).

**Errors:**
- `401 UNAUTHORIZED` — refresh token missing, expired, or blacklisted.

---

### POST `/api/auth/verify-email`

Verify email address using token from verification email.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | Auth tier |

**Request:**

```json
{
  "token": "abc123..."
}
```

**Response (200):**

```json
{
  "data": {
    "message": "Email verified successfully"
  }
}
```

**Errors:**
- `400 BAD_REQUEST` — token expired (24h) or already used.
- `404 NOT_FOUND` — invalid token.

---

### POST `/api/auth/forgot-password`

Request a password reset email.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | Auth tier (10/min per IP) |

**Request:**

```json
{
  "email": "user@example.com"
}
```

**Response (200):** Always returns success (don't reveal whether email exists).

```json
{
  "data": {
    "message": "If an account with that email exists, a reset link has been sent"
  }
}
```

**Side effects:**
- If email exists: queues password reset email with 1h-expiry token.
- If email doesn't exist: no-op (prevent user enumeration).

---

### POST `/api/auth/reset-password`

Reset password using token from reset email.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | Auth tier |

**Request:**

```json
{
  "token": "abc123...",
  "new_password": "NewSecure5"
}
```

**Validation:**
- `new_password`: same rules as registration password.

**Response (200):**

```json
{
  "data": {
    "message": "Password reset successfully"
  }
}
```

**Side effects:**
- Token marked as used.
- All existing refresh tokens for the user are invalidated (force re-login on all devices).

**Errors:**
- `400 BAD_REQUEST` — token expired (1h) or already used.

---

## 3. User Endpoints

### GET `/api/users/me`

Get current user's profile.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": {
    "id": "01912345-...",
    "email": "user@example.com",
    "name": "Sujay Kumar",
    "role": "user",
    "email_verified": true,
    "premium_since": "2026-03-10T08:00:00Z",
    "current_title": "Senior Software Engineer",
    "experience_level": "senior",
    "preferred_tech_stacks": ["Go", "Python", "Kubernetes"],
    "target_domains": ["Cloud", "Infra"],
    "target_locations": ["Bangalore", "Hyderabad"],
    "default_resume_id": "01912346-...",
    "created_at": "2026-03-01T10:00:00Z",
    "updated_at": "2026-03-11T08:00:00Z"
  }
}
```

---

### PUT `/api/users/me`

Update current user's profile.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Request (partial update — all fields optional):**

```json
{
  "name": "Sujay K",
  "current_title": "Staff Engineer",
  "experience_level": "staff_plus",
  "preferred_tech_stacks": ["Go", "Rust"],
  "target_domains": ["Cloud", "Infra", "Platform"],
  "target_locations": ["Bangalore"]
}
```

**Response (200):** Updated user object (same shape as GET).

---

### PUT `/api/users/me/password`

Change password (requires current password).

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Auth tier |

**Request:**

```json
{
  "current_password": "OldP4ss",
  "new_password": "NewSecure5"
}
```

**Response (200):**

```json
{
  "data": {
    "message": "Password changed successfully"
  }
}
```

**Side effects:**
- All other refresh tokens invalidated (keep current session only).

**Errors:**
- `401 UNAUTHORIZED` — current password incorrect.
- `422 VALIDATION_ERROR` — new password doesn't meet requirements.

---

### DELETE `/api/users/me`

Soft-delete account. 30-day grace period before hard delete.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Request:**

```json
{
  "password": "ConfirmP4ss"
}
```

**Response (200):**

```json
{
  "data": {
    "message": "Account scheduled for deletion. You have 30 days to recover it by logging in."
  }
}
```

**Side effects:**
- Sets `deleted_at = NOW()` on user.
- Clears all cookies (logs out).
- Scheduled job hard-deletes after 30 days (cascades to resumes, lists, etc.).

**Errors:**
- `401 UNAUTHORIZED` — password incorrect.

---

## 4. Company Endpoints (Public)

### GET `/api/companies`

Browse, search, and filter the company directory.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | Public tier (60/min per IP) |
| Caching | `Cache-Control: public, max-age=300`, `ETag` |

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Full-text search query |
| `tech_stack` | string (comma-separated) | Filter by tech stack (AND — must have all) |
| `domains` | string (comma-separated) | Filter by domain (OR — any match) |
| `size` | string (comma-separated) | Filter by size enum(s) |
| `hiring_status` | string | Filter: `active`, `paused`, `unknown` |
| `compensation_tier` | string (comma-separated) | Filter by tier(s) |
| `has_rsu` | boolean | Filter by RSU availability |
| `headquarters` | string | Filter by city (partial match) |
| `sort` | string | `name`, `size`, `compensation_tier`, `updated_at` (default: relevance if `q` present, `name` otherwise) |
| `order` | string | `asc` (default), `desc` |
| `cursor` | string | Pagination cursor |
| `limit` | integer | 1-100, default 20 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912345-...",
      "slug": "google",
      "name": "Google",
      "logo_url": "https://assets.careerdock.skriptvalley.com/google.png",
      "description": "Global technology company...",
      "size": "enterprise",
      "headquarters": "Bangalore, Karnataka",
      "tech_stack": ["Go", "C++", "Python", "gRPC", "Kubernetes"],
      "domains": ["Cloud", "AI/ML", "Infra", "Platform"],
      "hiring_status": "active",
      "compensation_tier": "tier_1",
      "has_rsu": true,
      "has_rsu_refresher": true,
      "updated_at": "2026-03-10T12:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "eyJpZCI6...",
    "has_more": true
  }
}
```

**Note:** List response returns a summary — no `interview_patterns`, `compensation_bands`, or `founded_year`. Those are on the detail endpoint.

**Full-text search:** When `q` is provided, uses `search_vector @@ plainto_tsquery('english', q)` with `ts_rank` for relevance sorting.

**Array filters:**
- `tech_stack=Go,Docker` → `companies.tech_stack @> ARRAY['Go', 'Docker']` (must have all).
- `domains=Cloud,SaaS` → `companies.domains && ARRAY['Cloud', 'SaaS']` (any match).

---

### GET `/api/companies/:slug`

Get full company profile.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | Public tier |
| Caching | `Cache-Control: public, max-age=600`, `ETag` |

**Response (200):**

```json
{
  "data": {
    "id": "01912345-...",
    "slug": "google",
    "name": "Google",
    "logo_url": "https://assets.careerdock.skriptvalley.com/google.png",
    "description": "Global technology company...",
    "size": "enterprise",
    "headquarters": "Bangalore, Karnataka",
    "founded_year": 1998,
    "careers_page_url": "https://careers.google.com",
    "glassdoor_url": "https://glassdoor.com/...",
    "ambitionbox_url": "https://ambitionbox.com/...",
    "linkedin_url": "https://linkedin.com/company/google",
    "tech_stack": ["Go", "C++", "Python", "gRPC", "Kubernetes"],
    "domains": ["Cloud", "AI/ML", "Infra", "Platform"],
    "hiring_status": "active",
    "interview_patterns": { "roles": [ ... ] },
    "compensation_tier": "tier_1",
    "has_rsu": true,
    "has_rsu_refresher": true,
    "compensation_bands": { "roles": [ ... ] },
    "last_verified_at": "2026-03-01T00:00:00Z",
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — no company with that slug.

---

### POST `/api/companies/:id/edits`

Submit a suggested edit to a company profile (moderator only).

| Aspect | Detail |
|--------|--------|
| Auth | `moderator` |
| Rate limit | Authenticated tier |

**Request:**

```json
{
  "changes": {
    "hiring_status": "active",
    "tech_stack": ["Go", "C++", "Python", "gRPC", "Kubernetes", "Spanner"]
  }
}
```

**Validation:** `changes` must only contain valid company field names. Values are validated against the same rules as the company schema.

**Response (201):**

```json
{
  "data": {
    "id": "01912347-...",
    "company_id": "01912345-...",
    "status": "pending",
    "changes": { ... },
    "created_at": "2026-03-11T10:00:00Z"
  }
}
```

---

## 5. List Endpoints

### GET `/api/lists`

Get all lists for the current user.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912350-...",
      "name": "Dream Companies",
      "description": "Top tier targets",
      "position": 0,
      "entry_count": 12,
      "created_at": "2026-03-01T10:00:00Z",
      "updated_at": "2026-03-11T08:00:00Z"
    }
  ]
}
```

**Note:** No pagination — max 5 lists per user. Returns all with `entry_count` (aggregate).

---

### POST `/api/lists`

Create a new list.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Request:**

```json
{
  "name": "Backup Options",
  "description": "Safety net companies"
}
```

**Response (201):** Created list object.

**Errors:**
- `422 VALIDATION_ERROR` — list limit reached (3 free, 5 premium).

---

### GET `/api/lists/:id`

Get list detail with all entries.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (owner only) |
| Rate limit | Authenticated tier |

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `status` | string | Filter entries by application status |
| `sort` | string | `position` (default), `status`, `date_applied`, `updated_at` |
| `order` | string | `asc` (default), `desc` |

**Response (200):**

```json
{
  "data": {
    "id": "01912350-...",
    "name": "Dream Companies",
    "description": "Top tier targets",
    "position": 0,
    "entries": [
      {
        "id": "01912360-...",
        "company": {
          "id": "01912345-...",
          "slug": "google",
          "name": "Google",
          "logo_url": "https://assets.careerdock.skriptvalley.com/google.png",
          "compensation_tier": "tier_1",
          "hiring_status": "active"
        },
        "role_title": "SDE-3 (L5)",
        "status": "interview",
        "date_applied": "2026-02-15",
        "notes": "Referred by friend in Cloud team",
        "position": 0,
        "round_count": 2,
        "created_at": "2026-02-10T10:00:00Z",
        "updated_at": "2026-03-11T08:00:00Z"
      }
    ],
    "created_at": "2026-03-01T10:00:00Z",
    "updated_at": "2026-03-11T08:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — list doesn't exist or belongs to another user.

---

### PUT `/api/lists/:id`

Update list metadata.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (owner only) |
| Rate limit | Authenticated tier |

**Request (partial update):**

```json
{
  "name": "Top Targets",
  "position": 1
}
```

**Response (200):** Updated list object.

---

### DELETE `/api/lists/:id`

Delete a list and all its entries.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (owner only) |
| Rate limit | Authenticated tier |

**Response (204):** No content.

---

### POST `/api/lists/:id/entries`

Add a company + role entry to a list.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (list owner) |
| Rate limit | Authenticated tier |

**Request:**

```json
{
  "company_id": "01912345-...",
  "role_title": "SDE-3 (L5)",
  "status": "not_applied",
  "date_applied": null,
  "notes": "Referred by friend"
}
```

**Response (201):** Created entry object.

**Errors:**
- `404 NOT_FOUND` — invalid list ID or company ID.

---

### PUT `/api/lists/:id/entries/:entryId`

Update an entry (status, notes, date, position).

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (list owner) |
| Rate limit | Authenticated tier |

**Request (partial update):**

```json
{
  "status": "interview",
  "notes": "Passed phone screen, technical interview next week"
}
```

**Response (200):** Updated entry object.

**Side effects:**
- If `status` changed: inserts a row into `application_status_history`.

---

### DELETE `/api/lists/:id/entries/:entryId`

Remove an entry from a list.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (list owner) |
| Rate limit | Authenticated tier |

**Response (204):** No content. Cascades delete to status history and interview rounds.

---

### GET `/api/lists/:id/entries/:entryId/history`

Get status change history for an entry.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (list owner) |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912370-...",
      "from_status": null,
      "to_status": "not_applied",
      "changed_at": "2026-02-10T10:00:00Z"
    },
    {
      "id": "01912371-...",
      "from_status": "not_applied",
      "to_status": "applied",
      "changed_at": "2026-02-15T09:00:00Z"
    },
    {
      "id": "01912372-...",
      "from_status": "applied",
      "to_status": "interview",
      "changed_at": "2026-03-01T14:00:00Z"
    }
  ]
}
```

---

### POST `/api/lists/:id/entries/:entryId/rounds`

Add an interview round to an entry.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (list owner) |
| Rate limit | Authenticated tier |

**Request:**

```json
{
  "round_number": 1,
  "round_type": "Technical",
  "scheduled_date": "2026-03-15",
  "outcome": "pending",
  "notes": "DSA + system design"
}
```

**Response (201):** Created round object.

---

### PUT `/api/lists/:id/entries/:entryId/rounds/:roundId`

Update an interview round.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (list owner) |
| Rate limit | Authenticated tier |

**Request (partial update):**

```json
{
  "outcome": "passed",
  "notes": "DSA went well, moving to next round"
}
```

**Response (200):** Updated round object.

---

### DELETE `/api/lists/:id/entries/:entryId/rounds/:roundId`

Delete an interview round.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (list owner) |
| Rate limit | Authenticated tier |

**Response (204):** No content.

---

## 6. Resume Endpoints

### GET `/api/resumes`

List current user's resumes (active, non-archived).

| Aspect | Detail |
|--------|--------|
| Auth | `premium` |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912380-...",
      "slot_number": 1,
      "file_name": "sujay_resume_v3.pdf",
      "file_size_bytes": 245760,
      "status": "ready",
      "is_default": true,
      "ats_general_score": 78,
      "parsed_data_summary": {
        "years_of_experience": 6,
        "role_level": "senior",
        "top_skills": ["Go", "Kubernetes", "AWS"],
        "domains": ["Cloud", "Infra"]
      },
      "created_at": "2026-03-05T10:00:00Z",
      "updated_at": "2026-03-05T11:00:00Z"
    }
  ]
}
```

**Note:** Returns a summary of parsed data and general ATS score (just the top-level score). Full details via GET `/api/resumes/:id`.

---

### POST `/api/resumes`

Upload a resume PDF.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` |
| Rate limit | AI tier (10/min) |
| Content-Type | `multipart/form-data` |

**Request (form fields):**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | file | yes | PDF file, max 5 MB |
| `slot_number` | integer | yes | 1, 2, or 3 |

**Response (201):**

```json
{
  "data": {
    "id": "01912380-...",
    "slot_number": 1,
    "file_name": "sujay_resume_v3.pdf",
    "file_size_bytes": 245760,
    "status": "extracting",
    "is_default": false,
    "created_at": "2026-03-05T10:00:00Z"
  }
}
```

**Processing pipeline (async):**
1. Validate file (PDF, ≤5 MB).
2. Upload to S3 (`{user_id}/{resume_id}.pdf`).
3. Extract text from PDF (Go-native library).
4. Store `extracted_text` in `resumes` table. Status → `parsing`.
5. Queue Asynq job `resume:parse_and_score`.
6. Worker sends text to Claude API for parsing + general ATS scoring.
7. Store `parsed_data` and `ats_general` in `resumes` table. Status → `ready`.
8. Send SSE notification `resume_parsed`.

**If slot is occupied:** Archives the existing resume (sets `is_archived = true`, `archived_at = NOW()`). Consumes 1 `resume_upload` credit.

**Errors:**
- `422 VALIDATION_ERROR` — not a PDF, exceeds 5 MB, invalid slot number.
- `422 VALIDATION_ERROR` — insufficient `resume_upload` credits.

---

### GET `/api/resumes/:id`

Get full resume details including parsed data and general ATS score.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` (owner only) |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": {
    "id": "01912380-...",
    "slot_number": 1,
    "file_name": "sujay_resume_v3.pdf",
    "file_size_bytes": 245760,
    "status": "ready",
    "is_default": true,
    "parsed_data": {
      "name": "Sujay Kumar",
      "summary": "Senior backend engineer...",
      "years_of_experience": 6,
      "skills": { ... },
      "experience": [ ... ],
      "education": [ ... ],
      "domains": ["Cloud", "Infra"],
      "role_level": "senior"
    },
    "ats_general": {
      "score": 78,
      "breakdown": { ... },
      "suggestions": [ ... ],
      "generated_at": "2026-03-05T10:30:00Z"
    },
    "created_at": "2026-03-05T10:00:00Z",
    "updated_at": "2026-03-05T11:00:00Z"
  }
}
```

---

### PUT `/api/resumes/:id/default`

Set a resume as the default (used for AI-curated lists and ATS pre-selection).

| Aspect | Detail |
|--------|--------|
| Auth | `premium` (owner only) |
| Rate limit | Authenticated tier |

**Request:** Empty body.

**Response (200):**

```json
{
  "data": {
    "message": "Resume set as default"
  }
}
```

**Side effects:**
- Unsets `is_default` on all other resumes for this user.
- Updates `users.default_resume_id`.

---

### DELETE `/api/resumes/:id`

Archive a resume (remove from active slots).

| Aspect | Detail |
|--------|--------|
| Auth | `premium` (owner only) |
| Rate limit | Authenticated tier |

**Response (204):** No content.

**Side effects:**
- Sets `is_archived = true`, `archived_at = NOW()`.
- If this was the default resume, clears `users.default_resume_id`.
- S3 object retained for 90 days.
- Does NOT refund credits.

---

### GET `/api/resumes/:id/download`

Get a signed S3 URL for downloading the original PDF.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` (owner only) |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": {
    "download_url": "https://careerdock-resumes.s3.amazonaws.com/...?X-Amz-Signature=...",
    "expires_in_seconds": 900
  }
}
```

**Note:** Signed URL expires in 15 minutes.

---

## 7. ATS Endpoints

### POST `/api/ats/company`

Request a company-specific ATS check. Evaluates all active resumes against the company.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` |
| Rate limit | AI tier (10/min) |

**Request:**

```json
{
  "company_id": "01912345-..."
}
```

**Response (202):**

```json
{
  "data": {
    "id": "01912390-...",
    "check_type": "company",
    "company_id": "01912345-...",
    "status": "processing",
    "created_at": "2026-03-11T10:00:00Z"
  }
}
```

**Processing (async):**
1. Check Redis cache for existing result (`ats_company:{hash}:{company_id}`).
2. If cache miss: queue Asynq job `ats:company_check`.
3. Worker sends all active resumes + company profile to Claude API.
4. Store result in `ats_checks` table + Redis cache (30-day TTL).
5. Send SSE notification `ats_result_ready`.
6. Deduct 1 `ats_check` credit.

**Errors:**
- `422 VALIDATION_ERROR` — insufficient `ats_check` credits.
- `422 VALIDATION_ERROR` — no active resumes uploaded.
- `404 NOT_FOUND` — invalid company_id.

---

### POST `/api/ats/job`

Request a job-specific ATS check. User pastes JD text.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` |
| Rate limit | AI tier (10/min) |

**Request:**

```json
{
  "job_description": "We are looking for a Senior Backend Engineer with 5+ years of experience in Go, microservices architecture..."
}
```

**Validation:**
- `job_description`: 100-10,000 characters.

**Response (202):** Same shape as company check, with `check_type: "job"`.

**Processing:** Same as company check, but scored against JD text instead of company profile.

---

### GET `/api/ats/:id`

Get ATS check result.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` (owner only) |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": {
    "id": "01912390-...",
    "check_type": "company",
    "company": {
      "id": "01912345-...",
      "name": "Google",
      "slug": "google"
    },
    "result": {
      "score": 72,
      "breakdown": { ... },
      "best_resume": { ... },
      "suggestions": [ ... ],
      "generated_at": "2026-03-11T10:05:00Z"
    },
    "created_at": "2026-03-11T10:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — check doesn't exist, belongs to another user, or still processing (status field indicates).

---

### GET `/api/ats`

List ATS check history for the current user.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` |
| Rate limit | Authenticated tier |

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `check_type` | string | Filter: `company` or `job` |
| `cursor` | string | Pagination cursor |
| `limit` | integer | 1-100, default 20 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912390-...",
      "check_type": "company",
      "company": {
        "id": "01912345-...",
        "name": "Google",
        "slug": "google"
      },
      "score": 72,
      "created_at": "2026-03-11T10:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "eyJpZCI6...",
    "has_more": false
  }
}
```

---

## 8. Curated List Endpoints

### POST `/api/curated-lists`

Generate a new AI-curated company list based on default resume + user preferences.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` |
| Rate limit | AI tier (10/min) |

**Request:** Empty body (uses default resume + profile preferences).

**Response (202):**

```json
{
  "data": {
    "id": "01912395-...",
    "status": "processing",
    "created_at": "2026-03-11T12:00:00Z"
  }
}
```

**Processing (async):**
1. Read default resume's `parsed_data` + user's preferences.
2. Queue Asynq job `ai:curate_company_list`.
3. Worker sends resume profile + all company summaries to Claude API.
4. Store result in `curated_lists` table.
5. Send SSE notification `curated_list_ready`.
6. Deduct 1 `curated_list` credit.

**Errors:**
- `422 VALIDATION_ERROR` — no default resume set.
- `422 VALIDATION_ERROR` — insufficient `curated_list` credits.

---

### GET `/api/curated-lists`

List previous curated list generations.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912395-...",
      "resume_file_name": "sujay_resume_v3.pdf",
      "total_matches": 25,
      "top_score": 92,
      "created_at": "2026-03-11T12:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": null,
    "has_more": false
  }
}
```

---

### GET `/api/curated-lists/:id`

Get full curated list result.

| Aspect | Detail |
|--------|--------|
| Auth | `premium` (owner only) |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": {
    "id": "01912395-...",
    "resume_id": "01912380-...",
    "resume_file_name": "sujay_resume_v3.pdf",
    "result": {
      "total_companies_evaluated": 150,
      "matches": [
        {
          "company_id": "01912345-...",
          "company_name": "Google",
          "company_slug": "google",
          "match_score": 92,
          "reasoning": "Strong Go and K8s overlap...",
          "key_matches": ["Go", "Kubernetes", "Cloud"],
          "gaps": ["C++ experience preferred"]
        }
      ],
      "generated_at": "2026-03-11T12:05:00Z"
    },
    "created_at": "2026-03-11T12:00:00Z"
  }
}
```

---

## 9. Payment Endpoints

### POST `/api/payments/orders`

Create a Razorpay order for a purchase.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Payment tier (10/min) |

**Request:**

```json
{
  "product_type": "starter_pack"
}
```

**Valid product types and amounts:**

| Product Type | Amount (paise) | Display Price |
|-------------|---------------:|--------------|
| `starter_pack` | 39900 | ₹399 |
| `resume_upload` | 4900 | ₹49 |
| `ats_bundle` | 9900 | ₹99 |
| `rebuy_pack` | 39900 | ₹399 |

**Response (201):**

```json
{
  "data": {
    "payment_id": "01912400-...",
    "razorpay_order_id": "order_abc123",
    "amount_paise": 39900,
    "currency": "INR",
    "product_type": "starter_pack",
    "razorpay_key_id": "rzp_live_xxxxx"
  }
}
```

**Frontend** uses `razorpay_order_id` and `razorpay_key_id` to open Razorpay Checkout widget.

---

### POST `/api/payments/webhook`

Razorpay webhook handler. Called by Razorpay on payment events.

| Aspect | Detail |
|--------|--------|
| Auth | Razorpay webhook signature verification |
| Rate limit | None (Razorpay-originated) |

**Request:** Razorpay webhook payload (verified via `X-Razorpay-Signature` header).

**Processing:**
1. Verify webhook signature using Razorpay webhook secret.
2. Extract `razorpay_order_id` and `razorpay_payment_id`.
3. Look up payment by `razorpay_order_id` (idempotency check — if already captured, return 200).
4. Update payment status to `captured`.
5. Allocate credits based on `product_type`:

| Product | Credits Allocated |
|---------|------------------|
| `starter_pack` | 9 `resume_upload` + 20 `ats_check` + 3 `curated_list`. Set `premium_since` if not already set. |
| `resume_upload` | 1 `resume_upload` |
| `ats_bundle` | 10 `ats_check` |
| `rebuy_pack` | Same as `starter_pack` |

6. Create `credit_transactions` entries.
7. Send SSE notification `credits_updated`.
8. Queue payment receipt email.

**Response (200):** `{ "status": "ok" }`

**Errors:**
- `400 BAD_REQUEST` — invalid signature.

---

### GET `/api/credits`

Get current credit balances.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Response (200):**

```json
{
  "data": {
    "resume_upload": 7,
    "ats_check": 18,
    "curated_list": 2,
    "cv_generation": 0
  }
}
```

---

### GET `/api/credits/transactions`

Get credit transaction history.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `credit_type` | string | Filter by type |
| `cursor` | string | Pagination cursor |
| `limit` | integer | 1-100, default 20 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912410-...",
      "credit_type": "ats_check",
      "amount": -1,
      "balance_after": 18,
      "reason": "ats_check_consumed",
      "created_at": "2026-03-11T10:05:00Z"
    },
    {
      "id": "01912411-...",
      "credit_type": "ats_check",
      "amount": 20,
      "balance_after": 20,
      "reason": "starter_pack_purchase",
      "created_at": "2026-03-10T08:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "eyJpZCI6...",
    "has_more": true
  }
}
```

---

## 10. Notification Endpoints

### GET `/api/notifications/stream`

SSE (Server-Sent Events) stream for real-time notifications.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Content-Type | `text/event-stream` |
| Rate limit | 1 connection per user |

**Event format:**

```
event: resume_parsed
data: {"notification_id": "01912420-...", "resume_id": "01912380-...", "title": "Resume analysis complete", "score": 78}

event: ats_result_ready
data: {"notification_id": "01912421-...", "check_id": "01912390-...", "title": "ATS check complete", "score": 72}

event: curated_list_ready
data: {"notification_id": "01912422-...", "curated_list_id": "01912395-...", "title": "AI-curated list ready", "total_matches": 25}

event: credits_updated
data: {"notification_id": "01912423-...", "title": "Credits added", "credits": {"ats_check": 20}}

event: heartbeat
data: {}
```

**Connection management:**
- Heartbeat every 30 seconds to keep connection alive.
- Client should reconnect on drop with `Last-Event-ID` header.
- Server sends missed notifications since `Last-Event-ID` on reconnect.
- Max 1 SSE connection per user. New connection replaces old one.

---

### GET `/api/notifications`

List notification history.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Query parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `unread_only` | boolean | Only unread notifications (default false) |
| `cursor` | string | Pagination cursor |
| `limit` | integer | 1-100, default 20 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912420-...",
      "type": "resume_parsed",
      "title": "Resume analysis complete",
      "message": "Your resume sujay_resume_v3.pdf has been analysed. General ATS score: 78/100.",
      "data": { "resume_id": "01912380-...", "score": 78 },
      "read_at": null,
      "created_at": "2026-03-11T10:05:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "eyJpZCI6...",
    "has_more": false
  }
}
```

---

### PUT `/api/notifications/:id/read`

Mark a notification as read.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` (owner only) |
| Rate limit | Authenticated tier |

**Response (204):** No content.

---

### PUT `/api/notifications/read-all`

Mark all notifications as read.

| Aspect | Detail |
|--------|--------|
| Auth | `authenticated` |
| Rate limit | Authenticated tier |

**Response (204):** No content.

---

## 11. Admin Endpoints

All admin endpoints require `admin` role. Prefix: `/api/admin/`.

### 11.1 Dashboard

#### GET `/api/admin/dashboard`

Platform overview stats.

**Response (200):**

```json
{
  "data": {
    "users": {
      "total": 850,
      "premium": 95,
      "new_today": 12,
      "new_this_week": 45
    },
    "revenue": {
      "total_paise": 3500000,
      "this_month_paise": 450000,
      "today_paise": 39900
    },
    "companies": {
      "total": 156,
      "pending_edits": 3
    },
    "ai": {
      "total_operations_today": 34,
      "estimated_cost_today_paise": 4200
    },
    "system": {
      "api_uptime_percent": 99.8,
      "worker_queue_depth": 2,
      "failed_jobs_24h": 0
    }
  }
}
```

---

### 11.2 User Management

#### GET `/api/admin/users`

List users with search and filters.

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Search by name or email |
| `role` | string | Filter by role |
| `premium` | boolean | Filter premium status |
| `sort` | string | `created_at`, `name`, `email` |
| `order` | string | `asc`, `desc` |
| `page` | integer | Page number (offset pagination) |
| `per_page` | integer | 1-100, default 50 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912345-...",
      "email": "user@example.com",
      "name": "Sujay Kumar",
      "role": "user",
      "premium_since": "2026-03-10T08:00:00Z",
      "email_verified": true,
      "resume_count": 2,
      "list_count": 3,
      "created_at": "2026-03-01T10:00:00Z"
    }
  ],
  "pagination": {
    "total": 850,
    "page": 1,
    "per_page": 50,
    "total_pages": 17
  }
}
```

---

#### GET `/api/admin/users/:id`

Get detailed user info (profile + credits + recent activity).

**Response (200):**

```json
{
  "data": {
    "id": "01912345-...",
    "email": "user@example.com",
    "name": "Sujay Kumar",
    "role": "user",
    "premium_since": "2026-03-10T08:00:00Z",
    "email_verified": true,
    "deleted_at": null,
    "credits": {
      "resume_upload": 7,
      "ats_check": 18,
      "curated_list": 2,
      "cv_generation": 0
    },
    "stats": {
      "resume_count": 2,
      "list_count": 3,
      "ats_check_count": 5,
      "total_spent_paise": 39900
    },
    "created_at": "2026-03-01T10:00:00Z",
    "updated_at": "2026-03-11T08:00:00Z"
  }
}
```

---

#### PUT `/api/admin/users/:id/suspend`

Soft-delete (suspend) a user.

**Response (200):**

```json
{
  "data": {
    "message": "User suspended"
  }
}
```

**Side effects:** Sets `deleted_at = NOW()`. Audit log entry created.

---

#### PUT `/api/admin/users/:id/unsuspend`

Reactivate a suspended user.

**Response (200):**

```json
{
  "data": {
    "message": "User unsuspended"
  }
}
```

**Side effects:** Clears `deleted_at`. Audit log entry created.

---

#### PUT `/api/admin/users/:id/role`

Change a user's role.

**Request:**

```json
{
  "role": "moderator"
}
```

**Response (200):** Updated user object.

**Side effects:** Audit log entry created.

---

### 11.3 Company Management

#### POST `/api/admin/companies`

Create a new company.

**Request:** Full company object (all fields from the company schema).

**Response (201):** Created company object.

**Side effects:** `slug` auto-generated from `name` if not provided. Audit log entry.

---

#### PUT `/api/admin/companies/:id`

Update a company.

**Request:** Partial update — any company fields.

**Response (200):** Updated company object.

**Side effects:** Invalidates Redis cache for any ATS checks involving this company. Audit log entry.

---

#### DELETE `/api/admin/companies/:id`

Delete a company.

**Response (204):** No content.

**Side effects:** Cascades to `list_entries` (entries referencing this company are removed). Audit log entry.

---

#### POST `/api/admin/companies/:id/enrich`

Trigger AI enrichment for a company (tech stack verification, interview patterns, compensation research).

**Response (202):**

```json
{
  "data": {
    "message": "Enrichment job queued",
    "job_id": "asynq-job-id"
  }
}
```

---

#### POST `/api/admin/companies/:id/refresh`

Trigger AI data refresh for a company (re-research all fields from public sources).

**Response (202):** Same as enrich.

---

### 11.4 Moderation Queue

#### GET `/api/admin/moderation/edits`

List company edit suggestions.

| Param | Type | Description |
|-------|------|-------------|
| `status` | string | `pending` (default), `approved`, `rejected`, `all` |
| `page` | integer | Page number |
| `per_page` | integer | 1-100, default 20 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912347-...",
      "company": {
        "id": "01912345-...",
        "name": "Google",
        "slug": "google"
      },
      "submitted_by": {
        "id": "01912348-...",
        "name": "Moderator User",
        "email": "mod@example.com"
      },
      "status": "pending",
      "changes": {
        "hiring_status": "active",
        "tech_stack": ["Go", "C++", "Python", "gRPC", "Kubernetes", "Spanner"]
      },
      "created_at": "2026-03-11T10:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

---

#### GET `/api/admin/moderation/edits/:id`

Get edit detail with current company data for comparison.

**Response (200):**

```json
{
  "data": {
    "id": "01912347-...",
    "company": { ... },
    "submitted_by": { ... },
    "status": "pending",
    "changes": { ... },
    "current_values": {
      "hiring_status": "unknown",
      "tech_stack": ["Go", "C++", "Python", "gRPC", "Kubernetes"]
    },
    "created_at": "2026-03-11T10:00:00Z"
  }
}
```

---

#### POST `/api/admin/moderation/edits/:id/approve`

Approve an edit and apply changes to the company.

**Request (optional):**

```json
{
  "review_notes": "Verified via LinkedIn job postings"
}
```

**Response (200):**

```json
{
  "data": {
    "message": "Edit approved and applied"
  }
}
```

**Side effects:** Applies `changes` to the company record. Sets edit status to `approved`. Audit log entry.

---

#### POST `/api/admin/moderation/edits/:id/reject`

Reject an edit.

**Request:**

```json
{
  "review_notes": "Cannot verify this information"
}
```

**Response (200):**

```json
{
  "data": {
    "message": "Edit rejected"
  }
}
```

---

### 11.5 Payment Management

#### GET `/api/admin/payments`

List all payments with filters.

| Param | Type | Description |
|-------|------|-------------|
| `status` | string | Filter: `created`, `captured`, `failed`, `refunded` |
| `product_type` | string | Filter by product |
| `user_id` | uuid | Filter by user |
| `from` | date | Start date |
| `to` | date | End date |
| `page` | integer | Page number |
| `per_page` | integer | 1-100, default 50 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912400-...",
      "user": {
        "id": "01912345-...",
        "name": "Sujay Kumar",
        "email": "user@example.com"
      },
      "razorpay_order_id": "order_abc123",
      "razorpay_payment_id": "pay_xyz789",
      "amount_paise": 39900,
      "currency": "INR",
      "product_type": "starter_pack",
      "status": "captured",
      "receipt_number": "CDOCK-20260311-0001",
      "created_at": "2026-03-11T08:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

---

#### POST `/api/admin/payments/:id/refund`

Issue a refund. Only allowed if no credits have been consumed from this payment, and within 7 days.

**Request:**

```json
{
  "reason": "Customer requested refund, no credits used"
}
```

**Response (200):**

```json
{
  "data": {
    "message": "Refund processed",
    "refunded_amount_paise": 39900
  }
}
```

**Side effects:**
- Sets payment status to `refunded`.
- Deducts credits allocated by this payment.
- Triggers Razorpay refund API.
- If all credits from Starter Pack are refunded and no other purchases exist, clears `premium_since`.
- Audit log entry.

**Errors:**
- `422 VALIDATION_ERROR` — credits from this payment have been consumed.
- `422 VALIDATION_ERROR` — payment older than 7 days.
- `422 VALIDATION_ERROR` — payment not in `captured` status.

---

### 11.6 Feature Flags

#### GET `/api/admin/feature-flags`

List all feature flags.

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912500-...",
      "key": "ai_curated_lists",
      "enabled": true,
      "description": "Enable AI-curated company list generation",
      "updated_at": "2026-03-11T08:00:00Z"
    }
  ]
}
```

---

#### POST `/api/admin/feature-flags`

Create a new feature flag.

**Request:**

```json
{
  "key": "cv_generation",
  "enabled": false,
  "description": "Enable CV generation feature"
}
```

**Response (201):** Created flag object.

---

#### PUT `/api/admin/feature-flags/:key`

Update a feature flag (toggle, update description).

**Request:**

```json
{
  "enabled": true
}
```

**Response (200):** Updated flag object.

**Side effects:** Invalidates Redis cache for this flag. Audit log entry.

---

### 11.7 AI Cost Tracking

#### GET `/api/admin/ai/costs`

AI operation cost summary.

| Param | Type | Description |
|-------|------|-------------|
| `from` | date | Start date (default: 30 days ago) |
| `to` | date | End date (default: today) |
| `group_by` | string | `day` (default), `week`, `month` |

**Response (200):**

```json
{
  "data": {
    "total_operations": 340,
    "total_estimated_cost_paise": 42000,
    "by_operation": {
      "resume_parse": { "count": 45, "cost_paise": 2250 },
      "ats_general": { "count": 45, "cost_paise": 3600 },
      "ats_company": { "count": 120, "cost_paise": 12000 },
      "ats_job": { "count": 100, "cost_paise": 12000 },
      "curated_list": { "count": 20, "cost_paise": 4000 },
      "company_enrich": { "count": 10, "cost_paise": 600 }
    },
    "by_period": [
      { "date": "2026-03-11", "operations": 34, "cost_paise": 4200 },
      { "date": "2026-03-10", "operations": 28, "cost_paise": 3500 }
    ]
  }
}
```

**Note:** Cost estimates are computed from token counts stored per AI operation (tracked in `ats_checks.result` and `resumes.ats_general` JSONB — each includes `tokens_used` field).

---

### 11.8 Audit Log

#### GET `/api/admin/audit-log`

Admin action log.

| Param | Type | Description |
|-------|------|-------------|
| `admin_id` | uuid | Filter by admin |
| `entity_type` | string | Filter: `user`, `company`, `payment`, `feature_flag` |
| `entity_id` | uuid | Filter by entity |
| `from` | date | Start date |
| `to` | date | End date |
| `page` | integer | Page number |
| `per_page` | integer | 1-100, default 50 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "01912600-...",
      "admin": {
        "id": "01912001-...",
        "name": "Admin User"
      },
      "action": "user_suspended",
      "entity_type": "user",
      "entity_id": "01912345-...",
      "details": {
        "reason": "Terms of service violation"
      },
      "ip_address": "203.0.113.1",
      "created_at": "2026-03-11T10:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

---

## 12. Health Check

### GET `/api/health`

Basic health check for uptime monitoring.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | None |

**Response (200):**

```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2026-03-11T10:00:00Z"
}
```

---

### GET `/api/health/ready`

Readiness check — verifies DB and Redis connectivity.

| Aspect | Detail |
|--------|--------|
| Auth | `public` |
| Rate limit | None |

**Response (200):**

```json
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "redis": "ok"
  }
}
```

**Response (503):** If any dependency is unreachable.

---

## 13. Endpoint Summary

| # | Method | Path | Auth | Rate Tier | Description |
|---|--------|------|------|-----------|-------------|
| | **Auth** | | | | |
| 1 | POST | `/api/auth/register` | public | auth | Register |
| 2 | POST | `/api/auth/login` | public | auth | Login |
| 3 | POST | `/api/auth/logout` | authenticated | default | Logout |
| 4 | POST | `/api/auth/refresh` | refresh cookie | auth | Refresh tokens |
| 5 | POST | `/api/auth/verify-email` | public | auth | Verify email |
| 6 | POST | `/api/auth/forgot-password` | public | auth | Request reset |
| 7 | POST | `/api/auth/reset-password` | public | auth | Reset password |
| | **Users** | | | | |
| 8 | GET | `/api/users/me` | authenticated | default | Get profile |
| 9 | PUT | `/api/users/me` | authenticated | default | Update profile |
| 10 | PUT | `/api/users/me/password` | authenticated | auth | Change password |
| 11 | DELETE | `/api/users/me` | authenticated | default | Delete account |
| | **Companies** | | | | |
| 12 | GET | `/api/companies` | public | public | List/search |
| 13 | GET | `/api/companies/:slug` | public | public | Get company |
| 14 | POST | `/api/companies/:id/edits` | moderator | default | Submit edit |
| | **Lists** | | | | |
| 15 | GET | `/api/lists` | authenticated | default | List user's lists |
| 16 | POST | `/api/lists` | authenticated | default | Create list |
| 17 | GET | `/api/lists/:id` | authenticated | default | Get list detail |
| 18 | PUT | `/api/lists/:id` | authenticated | default | Update list |
| 19 | DELETE | `/api/lists/:id` | authenticated | default | Delete list |
| 20 | POST | `/api/lists/:id/entries` | authenticated | default | Add entry |
| 21 | PUT | `/api/lists/:id/entries/:entryId` | authenticated | default | Update entry |
| 22 | DELETE | `/api/lists/:id/entries/:entryId` | authenticated | default | Remove entry |
| 23 | GET | `/api/lists/:id/entries/:entryId/history` | authenticated | default | Status history |
| 24 | POST | `/api/lists/:id/entries/:entryId/rounds` | authenticated | default | Add round |
| 25 | PUT | `/api/lists/:id/entries/:entryId/rounds/:roundId` | authenticated | default | Update round |
| 26 | DELETE | `/api/lists/:id/entries/:entryId/rounds/:roundId` | authenticated | default | Delete round |
| | **Resumes** | | | | |
| 27 | GET | `/api/resumes` | premium | default | List resumes |
| 28 | POST | `/api/resumes` | premium | ai | Upload resume |
| 29 | GET | `/api/resumes/:id` | premium | default | Get resume detail |
| 30 | PUT | `/api/resumes/:id/default` | premium | default | Set default |
| 31 | DELETE | `/api/resumes/:id` | premium | default | Archive resume |
| 32 | GET | `/api/resumes/:id/download` | premium | default | Get download URL |
| | **ATS** | | | | |
| 33 | POST | `/api/ats/company` | premium | ai | Company ATS check |
| 34 | POST | `/api/ats/job` | premium | ai | Job ATS check |
| 35 | GET | `/api/ats/:id` | premium | default | Get ATS result |
| 36 | GET | `/api/ats` | premium | default | ATS history |
| | **Curated Lists** | | | | |
| 37 | POST | `/api/curated-lists` | premium | ai | Generate curated list |
| 38 | GET | `/api/curated-lists` | premium | default | List history |
| 39 | GET | `/api/curated-lists/:id` | premium | default | Get result |
| | **Payments** | | | | |
| 40 | POST | `/api/payments/orders` | authenticated | payment | Create order |
| 41 | POST | `/api/payments/webhook` | razorpay sig | none | Webhook |
| 42 | GET | `/api/credits` | authenticated | default | Credit balances |
| 43 | GET | `/api/credits/transactions` | authenticated | default | Credit history |
| | **Notifications** | | | | |
| 44 | GET | `/api/notifications/stream` | authenticated | 1/user | SSE stream |
| 45 | GET | `/api/notifications` | authenticated | default | List notifications |
| 46 | PUT | `/api/notifications/:id/read` | authenticated | default | Mark read |
| 47 | PUT | `/api/notifications/read-all` | authenticated | default | Mark all read |
| | **Admin** | | | | |
| 48 | GET | `/api/admin/dashboard` | admin | admin | Dashboard stats |
| 49 | GET | `/api/admin/users` | admin | admin | List users |
| 50 | GET | `/api/admin/users/:id` | admin | admin | User detail |
| 51 | PUT | `/api/admin/users/:id/suspend` | admin | admin | Suspend user |
| 52 | PUT | `/api/admin/users/:id/unsuspend` | admin | admin | Unsuspend user |
| 53 | PUT | `/api/admin/users/:id/role` | admin | admin | Change role |
| 54 | POST | `/api/admin/companies` | admin | admin | Create company |
| 55 | PUT | `/api/admin/companies/:id` | admin | admin | Update company |
| 56 | DELETE | `/api/admin/companies/:id` | admin | admin | Delete company |
| 57 | POST | `/api/admin/companies/:id/enrich` | admin | admin | AI enrich |
| 58 | POST | `/api/admin/companies/:id/refresh` | admin | admin | AI refresh |
| 59 | GET | `/api/admin/moderation/edits` | admin | admin | List edits |
| 60 | GET | `/api/admin/moderation/edits/:id` | admin | admin | Edit detail |
| 61 | POST | `/api/admin/moderation/edits/:id/approve` | admin | admin | Approve edit |
| 62 | POST | `/api/admin/moderation/edits/:id/reject` | admin | admin | Reject edit |
| 63 | GET | `/api/admin/payments` | admin | admin | List payments |
| 64 | POST | `/api/admin/payments/:id/refund` | admin | admin | Issue refund |
| 65 | GET | `/api/admin/feature-flags` | admin | admin | List flags |
| 66 | POST | `/api/admin/feature-flags` | admin | admin | Create flag |
| 67 | PUT | `/api/admin/feature-flags/:key` | admin | admin | Update flag |
| 68 | GET | `/api/admin/ai/costs` | admin | admin | AI cost tracking |
| 69 | GET | `/api/admin/audit-log` | admin | admin | Audit log |
| | **Health** | | | | |
| 70 | GET | `/api/health` | public | none | Liveness |
| 71 | GET | `/api/health/ready` | public | none | Readiness |

**Total: 71 endpoints**

---

## 14. Chi Router Structure

Mapping to Go Chi router groups:

```go
r := chi.NewRouter()

// Global middleware
r.Use(middleware.RequestID)
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.RateLimiter)
r.Use(middleware.CORS)

r.Route("/api", func(r chi.Router) {
    // Health (no auth)
    r.Get("/health", handler.Health)
    r.Get("/health/ready", handler.HealthReady)

    // Auth (public, auth-rate-limited)
    r.Route("/auth", func(r chi.Router) {
        r.Use(middleware.AuthRateLimit)
        r.Post("/register", handler.Register)
        r.Post("/login", handler.Login)
        r.Post("/verify-email", handler.VerifyEmail)
        r.Post("/forgot-password", handler.ForgotPassword)
        r.Post("/reset-password", handler.ResetPassword)

        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireAuth)
            r.Post("/logout", handler.Logout)
            r.Post("/refresh", handler.Refresh)
        })
    })

    // Companies (public)
    r.Route("/companies", func(r chi.Router) {
        r.Use(middleware.PublicRateLimit)
        r.Get("/", handler.ListCompanies)
        r.Get("/{slug}", handler.GetCompany)

        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireAuth)
            r.Use(middleware.RequireModerator)
            r.Post("/{id}/edits", handler.SubmitCompanyEdit)
        })
    })

    // Authenticated routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth)

        // Users
        r.Route("/users", func(r chi.Router) {
            r.Get("/me", handler.GetProfile)
            r.Put("/me", handler.UpdateProfile)
            r.Put("/me/password", handler.ChangePassword)
            r.Delete("/me", handler.DeleteAccount)
        })

        // Lists
        r.Route("/lists", func(r chi.Router) {
            r.Get("/", handler.ListLists)
            r.Post("/", handler.CreateList)
            r.Route("/{listId}", func(r chi.Router) {
                r.Get("/", handler.GetList)
                r.Put("/", handler.UpdateList)
                r.Delete("/", handler.DeleteList)
                r.Post("/entries", handler.AddListEntry)
                r.Route("/entries/{entryId}", func(r chi.Router) {
                    r.Put("/", handler.UpdateListEntry)
                    r.Delete("/", handler.DeleteListEntry)
                    r.Get("/history", handler.GetEntryHistory)
                    r.Post("/rounds", handler.AddInterviewRound)
                    r.Put("/rounds/{roundId}", handler.UpdateInterviewRound)
                    r.Delete("/rounds/{roundId}", handler.DeleteInterviewRound)
                })
            })
        })

        // Notifications
        r.Route("/notifications", func(r chi.Router) {
            r.Get("/stream", handler.NotificationStream)
            r.Get("/", handler.ListNotifications)
            r.Put("/{id}/read", handler.MarkNotificationRead)
            r.Put("/read-all", handler.MarkAllNotificationsRead)
        })

        // Credits
        r.Get("/credits", handler.GetCredits)
        r.Get("/credits/transactions", handler.GetCreditTransactions)

        // Payments
        r.Route("/payments", func(r chi.Router) {
            r.Use(middleware.PaymentRateLimit)
            r.Post("/orders", handler.CreatePaymentOrder)
        })

        // Premium routes
        r.Group(func(r chi.Router) {
            r.Use(middleware.RequirePremium)

            // Resumes
            r.Route("/resumes", func(r chi.Router) {
                r.Get("/", handler.ListResumes)
                r.Post("/", handler.UploadResume) // AI rate limit applied in handler
                r.Route("/{id}", func(r chi.Router) {
                    r.Get("/", handler.GetResume)
                    r.Put("/default", handler.SetDefaultResume)
                    r.Delete("/", handler.ArchiveResume)
                    r.Get("/download", handler.GetResumeDownloadURL)
                })
            })

            // ATS
            r.Route("/ats", func(r chi.Router) {
                r.Get("/", handler.ListATSChecks)
                r.Post("/company", handler.RequestCompanyATS) // AI rate limit
                r.Post("/job", handler.RequestJobATS)         // AI rate limit
                r.Get("/{id}", handler.GetATSResult)
            })

            // Curated lists
            r.Route("/curated-lists", func(r chi.Router) {
                r.Get("/", handler.ListCuratedLists)
                r.Post("/", handler.GenerateCuratedList) // AI rate limit
                r.Get("/{id}", handler.GetCuratedList)
            })
        })

        // Admin routes
        r.Route("/admin", func(r chi.Router) {
            r.Use(middleware.RequireAdmin)
            // ... all admin sub-routes
        })
    })

    // Webhook (signature-verified, no JWT auth)
    r.Post("/payments/webhook", handler.RazorpayWebhook)
})
```

---

## 15. Cross-Reference

| Architecture Decision | API Implementation |
|----------------------|-------------------|
| JWT in httpOnly cookies (§5) | No Authorization header. Middleware reads cookies. |
| Async AI operations (§3.5) | AI endpoints return `202 Accepted`. Results via SSE + polling. |
| SSE for notifications (§3.10) | `GET /api/notifications/stream` with heartbeat + reconnection. |
| Razorpay webhooks (§3.8) | `POST /api/payments/webhook` with signature verification. |
| Role-based access (§5) | Middleware chain: `RequireAuth` → `RequirePremium` / `RequireModerator` / `RequireAdmin`. |
| Postgres FTS (§3.3) | `GET /api/companies?q=...` uses `plainto_tsquery`. |
| Cursor-based pagination (§3.1.3) | All public + user endpoints. Offset for admin only. |
| Idempotent payments (§3.8) | UNIQUE on `razorpay_order_id`. Duplicate webhooks return 200. |
