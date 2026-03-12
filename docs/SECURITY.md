# CareerDock — Security Design

> **Version:** 1.0
> **Status:** Draft (Phase 5)
> **Last updated:** 2026-03-12
> **Depends on:** [PRD.md](./PRD.md), [ARCHITECTURE.md](./ARCHITECTURE.md), [LLD/](./LLD/), [CODE-STRUCTURE.md](./CODE-STRUCTURE.md)

---

## 1. Security Overview

This document covers security controls **not already detailed** in prior phases. For reference:

| Topic | Primary Source |
|-------|---------------|
| JWT auth flow, token rotation, refresh blacklist | [ARCHITECTURE.md §3.2](./ARCHITECTURE.md), [LLD/api.md §1.2](./LLD/api.md) |
| Role-based access (user/moderator/admin/premium) | [LLD/api.md §1.2](./LLD/api.md) |
| Rate limiting (per-user, per-IP tiers) | [LLD/api.md §1.7](./LLD/api.md) |
| Razorpay webhook HMAC-SHA256 verification | [LLD/payments.md §3.2](./LLD/payments.md) |
| Atomic credit transactions, SELECT FOR UPDATE | [LLD/payments.md §5](./LLD/payments.md), [LLD/database.md §3.12-13](./LLD/database.md) |
| Admin audit logging | [LLD/database.md §3.16](./LLD/database.md) |
| Prompt injection defence (preamble + XML sandboxing) | [LLD/ai-service.md §4.7-4.8](./LLD/ai-service.md) |
| AI output validation (schema, score bounds, retry) | [LLD/ai-service.md §5](./LLD/ai-service.md) |
| Parameterised SQL via pgx (no string interpolation) | [CODE-STRUCTURE.md §2.7](./CODE-STRUCTURE.md) |

This document fills the gaps: security headers, CORS, CSRF, password policy, secrets management, input hardening, infrastructure security, dependency scanning, data privacy, and incident response.

---

## 2. Security Headers

All responses from the Go API pass through Nginx reverse proxy, which injects security headers. The frontend (Vercel) has its own header configuration via `next.config.js`.

### 2.1 Nginx Security Headers

```nginx
# infra/docker/nginx.conf (security headers section)

# Prevent clickjacking
add_header X-Frame-Options "DENY" always;

# Prevent MIME-type sniffing
add_header X-Content-Type-Options "nosniff" always;

# Enforce HTTPS (1 year, include subdomains)
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

# Control referrer information
add_header Referrer-Policy "strict-origin-when-cross-origin" always;

# Restrict browser features
add_header Permissions-Policy "camera=(), microphone=(), geolocation=(), payment=()" always;

# Content Security Policy
add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' https://*.cloudfront.net data:; connect-src 'self' https://api.careerdock.skriptvalley.com; font-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';" always;

# Hide server identity
server_tokens off;
```

### 2.2 Next.js Security Headers

```javascript
// frontend/next.config.js (headers section)
async headers() {
  return [
    {
      source: '/:path*',
      headers: [
        { key: 'X-Frame-Options', value: 'DENY' },
        { key: 'X-Content-Type-Options', value: 'nosniff' },
        { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
        { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=(), payment=()' },
      ],
    },
  ];
},
```

### 2.3 Header Reference

| Header | Value | Threat Mitigated |
|--------|-------|-----------------|
| `X-Frame-Options` | `DENY` | Clickjacking |
| `X-Content-Type-Options` | `nosniff` | MIME-type confusion attacks |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Protocol downgrade, cookie hijacking |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Referrer data leakage |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=(), payment=()` | Unnecessary browser API access |
| `Content-Security-Policy` | See §2.1 | XSS, data injection, clickjacking |
| `server_tokens` | `off` | Server fingerprinting |

---

## 3. CORS Policy

### 3.1 Configuration

```go
// internal/middleware/cors.go

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return cors.Handler(cors.Options{
        AllowedOrigins:   allowedOrigins,
        AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "X-Request-ID"},
        ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
        AllowCredentials: true,  // Required for httpOnly cookie auth
        MaxAge:           86400, // Preflight cache: 24 hours
    })
}
```

### 3.2 Allowed Origins by Environment

| Environment | Allowed Origins |
|-------------|----------------|
| Development | `http://localhost:3000` |
| Staging | `https://staging.careerdock.skriptvalley.com` |
| Production | `https://careerdock.skriptvalley.com` |

**Rules:**
- No wildcard (`*`) origins — ever. `AllowCredentials: true` forbids it anyway.
- Origins set via `ALLOWED_ORIGINS` env var (comma-separated).
- Requests from unlisted origins receive no `Access-Control-Allow-Origin` header — the browser blocks them.

---

## 4. CSRF Protection

### 4.1 Why SameSite=Strict Is Sufficient

CareerDock's auth uses **httpOnly cookies with `SameSite=Strict`**. This means:

1. Cookies are **never sent** on cross-origin requests (no cross-site form submission, no cross-site AJAX).
2. There is **no `Authorization` header** — the cookie is the only credential.
3. All state-changing endpoints require `Content-Type: application/json`, which cannot be sent by HTML forms (form submissions use `application/x-www-form-urlencoded` or `multipart/form-data`).

These three factors together eliminate CSRF without needing a separate token mechanism.

### 4.2 Exceptions

| Endpoint | CSRF Risk | Mitigation |
|----------|-----------|------------|
| `POST /api/webhooks/razorpay` | None — no cookie auth | HMAC-SHA256 signature verification |
| `GET /api/auth/verify-email/{token}` | None — idempotent, token-gated | Single-use token with 24h expiry |
| `POST /api/auth/reset-password` | Low — no session needed | Single-use token with 1h expiry |

### 4.3 Future Consideration

If CareerDock ever downgrades to `SameSite=Lax` (e.g., to support OAuth redirects), a double-submit cookie pattern must be added. Document this in an ADR if it happens.

---

## 5. Password Security

### 5.1 Hashing

| Setting | Value | Rationale |
|---------|-------|-----------|
| Algorithm | bcrypt | Industry standard, built into Go (`golang.org/x/crypto/bcrypt`) |
| Cost factor | 12 | ~250ms per hash on modern hardware. Balances security with UX on login |
| Upgrade strategy | Re-hash on next login if cost factor increases | Transparent to user |

```go
// internal/service/auth_service.go

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(bytes), err
}

func checkPassword(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

### 5.2 Password Policy

| Rule | Constraint | Enforced At |
|------|-----------|-------------|
| Minimum length | 8 characters | Handler (input validation) |
| Maximum length | 72 characters (bcrypt limit) | Handler |
| Complexity | At least 1 uppercase, 1 lowercase, 1 digit | Handler |
| Common password check | — | Deferred to v2 (HaveIBeenPwned k-anonymity API) |

### 5.3 Account Lockout

Brute-force protection via Redis sliding window:

```go
// internal/middleware/brute_force.go

// Key: bruteforce:{ip}:{email_hash}
// Threshold: 5 failed attempts in 15 minutes → lock for 15 minutes
// After lockout expires, counter resets

const (
    maxLoginAttempts  = 5
    lockoutWindow     = 15 * time.Minute
)
```

| Attempt | Result |
|---------|--------|
| 1-5 | Normal login (success or "Invalid credentials") |
| 6+ | `429 RATE_LIMITED` — "Too many login attempts. Try again in 15 minutes." |
| After 15 min | Counter resets, user can try again |

**Important:** Error messages never reveal whether the email exists. Both "wrong email" and "wrong password" return the same generic message: `"Invalid email or password"`.

### 5.4 Password Change Flow

1. User provides current password + new password.
2. Verify current password against stored hash.
3. Hash new password, update in DB.
4. **Invalidate all existing refresh tokens** for this user (force re-login on all devices).
5. Issue new access + refresh token pair for the current session.

---

## 6. JWT & Token Security

### 6.1 Signing Configuration

| Setting | Value | Rationale |
|---------|-------|-----------|
| Algorithm | HS256 (HMAC-SHA256) | Symmetric — sufficient for single-backend architecture |
| Secret length | Minimum 256 bits (32 bytes) | Matches HMAC-SHA256 key size |
| Secret source | `JWT_SECRET` env var | Loaded via Viper |

### 6.2 Token Claims

**Access token (15-min TTL):**

```json
{
  "sub": "01912345-6789-7abc-def0-123456789abc",
  "email": "user@example.com",
  "role": "user",
  "premium": true,
  "iat": 1709251200,
  "exp": 1709252100,
  "jti": "unique-token-id"
}
```

**Refresh token (7-day TTL):**

```json
{
  "sub": "01912345-6789-7abc-def0-123456789abc",
  "iat": 1709251200,
  "exp": 1709856000,
  "jti": "unique-refresh-token-id"
}
```

**Rules:**
- Access token carries role and premium status — avoids DB lookup on every request.
- Refresh token is minimal — only used to issue a new access token (which re-reads from DB).
- `jti` (JWT ID) is a UUID v4 — used for blacklisting on logout.

### 6.3 Token Lifecycle

```
Login
  ├── Verify credentials (bcrypt)
  ├── Generate access token (15 min) + refresh token (7 days)
  ├── Store refresh token JTI in Redis: session:{user_id}:{jti} → TTL 7d
  └── Set both as httpOnly/Secure/SameSite=Strict cookies

API Request
  ├── Middleware reads access token from cookie
  ├── Verify signature (HS256) + expiry
  ├── Extract claims (sub, role, premium)
  └── Proceed to handler

Token Refresh (access token expired)
  ├── Read refresh token from cookie
  ├── Verify signature + expiry
  ├── Check JTI not in blacklist (Redis)
  ├── Blacklist old refresh token JTI
  ├── Issue new access token + new refresh token (rotation)
  └── Set new cookies

Logout
  ├── Blacklist current refresh token JTI in Redis (TTL = remaining token lifetime)
  └── Clear both cookies (Set-Cookie with Max-Age=0)

Password Change
  ├── Delete ALL session:{user_id}:* keys in Redis
  └── Issue new token pair for current session only
```

### 6.4 Key Rotation Strategy

For HS256, key rotation requires a dual-key window:

1. Generate new `JWT_SECRET_NEW` alongside existing `JWT_SECRET`.
2. Deploy with both keys:
   - **Sign** new tokens with `JWT_SECRET_NEW`.
   - **Verify** with `JWT_SECRET_NEW` first, fallback to `JWT_SECRET`.
3. Wait 7 days (max refresh token lifetime) for all old tokens to expire.
4. Remove `JWT_SECRET` (old key).
5. Rename `JWT_SECRET_NEW` → `JWT_SECRET`.

```go
// internal/middleware/auth.go

func (a *Auth) verifyToken(tokenString string) (*Claims, error) {
    // Try current key first
    claims, err := parseJWT(tokenString, a.currentKey)
    if err == nil {
        return claims, nil
    }
    // Fallback to previous key (during rotation window)
    if a.previousKey != "" {
        return parseJWT(tokenString, a.previousKey)
    }
    return nil, err
}
```

**Rotation frequency:** At minimum annually, or immediately if key compromise is suspected.

---

## 7. Input Validation & Hardening

### 7.1 Request Size Limits

Applied at Nginx (first line of defence) and Go handler (second line):

| Limit | Nginx | Go Handler |
|-------|-------|------------|
| JSON body | `client_max_body_size 1m` | `http.MaxBytesReader(w, r.Body, 1<<20)` |
| File upload (resume) | `client_max_body_size 6m` (5MB + overhead) | Multipart reader with 5MB file limit |
| URL length | `large_client_header_buffers 4 8k` | — (Go default 1MB is fine) |
| Header size | `large_client_header_buffers 4 8k` | — |

### 7.2 File Upload Validation

Resume uploads are the only file upload in the system. Validate defensively:

```go
// internal/handler/resume.go — upload validation

func validateResumeUpload(file multipart.File, header *multipart.FileHeader) error {
    // 1. Size check (5MB max)
    if header.Size > 5*1024*1024 {
        return domain.ValidationError("File exceeds 5MB limit", map[string]any{
            "max_size_bytes": 5 * 1024 * 1024,
            "actual_size":   header.Size,
        })
    }

    // 2. Extension check
    ext := strings.ToLower(filepath.Ext(header.Filename))
    if ext != ".pdf" {
        return domain.ValidationError("Only PDF files are accepted", map[string]any{
            "allowed_extensions": []string{".pdf"},
        })
    }

    // 3. Magic bytes check (PDF starts with %PDF-)
    buf := make([]byte, 5)
    if _, err := file.Read(buf); err != nil {
        return domain.ValidationError("Unable to read file", nil)
    }
    if string(buf) != "%PDF-" {
        return domain.ValidationError("File is not a valid PDF", nil)
    }

    // 4. Reset reader position
    file.Seek(0, io.SeekStart)

    return nil
}
```

### 7.3 Validation Rules Reference

Consolidated from [LLD/api.md](./LLD/api.md):

| Field | Rules | Max Length |
|-------|-------|-----------|
| `email` | RFC 5322 format, lowercase, trimmed | 254 chars |
| `password` | 8-72 chars, 1 upper + 1 lower + 1 digit | 72 (bcrypt limit) |
| `name` | Non-empty, trimmed | 100 chars |
| `company.slug` | Lowercase alphanumeric + hyphens, no leading/trailing hyphens | 100 chars |
| `company.name` | Non-empty, trimmed | 200 chars |
| `company.website` | Valid URL (https preferred) | 500 chars |
| `list.name` | Non-empty, trimmed | 100 chars |
| `list.description` | Optional | 500 chars |
| `ats.job_url` | Valid URL | 2000 chars |
| `ats.job_description` | Non-empty | 10,000 chars |
| `application_status` | One of: wishlist, applied, screening, interviewing, offer, accepted, rejected, withdrawn | — |
| `cursor` | Base64-encoded, validated on decode | 200 chars |
| `limit` (pagination) | Integer, 1-100, default 20 | — |

### 7.4 SQL Injection Prevention

Already enforced by architecture (parameterised queries via pgx), but explicitly codified:

**Allowed:**
```go
conn.QueryRow(ctx, `SELECT * FROM users WHERE email = $1`, email)
```

**Forbidden (will fail code review):**
```go
conn.QueryRow(ctx, fmt.Sprintf(`SELECT * FROM users WHERE email = '%s'`, email))
```

The `golangci-lint` config includes `gocritic` which flags string concatenation in SQL-like contexts.

---

## 8. Infrastructure Security

### 8.1 Network Architecture

```
Internet
    │
    ▼
┌──────────────────────────────────────────────┐
│  EC2 Security Group: careerdock-api-sg        │
│                                              │
│  Inbound Rules:                              │
│    443 (HTTPS) ← 0.0.0.0/0                 │
│    80  (HTTP)  ← 0.0.0.0/0  → redirect 443 │
│    22  (SSH)   ← your-ip/32 only            │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │ Nginx (:443) — SSL termination         │  │
│  │   ├── /api/*  → Go API (:8080)         │  │
│  │   └── health  → Go API (:8080)         │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  ┌────────────┐  ┌────────────┐             │
│  │ Go API     │  │ Go Worker  │             │
│  │ :8080      │  │ (no port)  │             │
│  │ (internal) │  │ (internal) │             │
│  └─────┬──────┘  └─────┬──────┘             │
│        │                │                    │
│  ┌─────▼────────────────▼──────┐             │
│  │ Redis :6379 (internal only)  │             │
│  │ bind 127.0.0.1               │             │
│  │ requirepass <strong-pass>    │             │
│  └──────────────────────────────┘             │
└──────────────────────┬───────────────────────┘
                       │ (VPC internal)
          ┌────────────▼────────────────┐
          │  RDS PostgreSQL              │
          │  Security Group:             │
          │    5432 ← careerdock-api-sg  │
          │  No public IP                │
          │  Encryption at rest: ON      │
          └─────────────────────────────┘

          ┌─────────────────────────────┐
          │  S3                          │
          │  careerdock-resumes (private)│
          │  careerdock-logos (public/CF)│
          │  SSE-S3 encryption           │
          └─────────────────────────────┘
```

### 8.2 EC2 Hardening

| Setting | Value |
|---------|-------|
| AMI | Amazon Linux 2023 (latest) |
| SSH access | Key-based only, restricted to admin IP |
| Unattended updates | Enabled for security patches |
| Open ports | 443, 80 (redirect), 22 (restricted) |
| Docker rootless | Not for MVP — standard Docker with non-root containers |
| Go API container | Runs as non-root user (UID 1000) |

**Dockerfile non-root user:**
```dockerfile
# Final stage of Dockerfile.api / Dockerfile.worker
RUN adduser -D -u 1000 appuser
USER appuser
```

### 8.3 RDS Security

| Setting | Value |
|---------|-------|
| Public access | Disabled |
| Security group | Inbound 5432 only from `careerdock-api-sg` |
| Encryption at rest | AWS-managed KMS key (default, free) |
| Encryption in transit | `sslmode=require` in connection string |
| Automated backups | Enabled, 7-day retention |
| Multi-AZ | Disabled for MVP (cost). Enable when revenue supports it |
| Deletion protection | Enabled in production |
| IAM auth | Not for MVP — password-based via Secrets Manager |

**Connection string (production):**
```
postgres://careerdock:<password>@careerdock-db.xxxxx.rds.amazonaws.com:5432/careerdock?sslmode=require
```

### 8.4 Redis Security

| Setting | Value |
|---------|-------|
| Bind address | `127.0.0.1` (localhost only, Docker network in compose) |
| Authentication | `requirepass` set in production |
| TLS | Not for MVP (localhost-only traffic). Add if Redis moves to ElastiCache |
| Persistence | AOF enabled (appendonly yes) for durability |
| Max memory | 256MB (plenty for MVP: sessions + rate limits + AI cache) |
| Eviction policy | `allkeys-lru` (least-recently-used eviction when full) |

**Redis config (production):**
```conf
bind 127.0.0.1
requirepass <strong-random-password>
appendonly yes
maxmemory 256mb
maxmemory-policy allkeys-lru
```

### 8.5 S3 Security

**`careerdock-resumes` (private):**

| Setting | Value |
|---------|-------|
| Public access | Block all public access (bucket policy) |
| Encryption | SSE-S3 (AES-256, AWS-managed keys) |
| Versioning | Disabled (no need — PDFs are immutable, identified by resume_id) |
| Access | IAM user with scoped policy (PutObject, GetObject, DeleteObject on this bucket only) |
| Signed URLs | 15-minute expiry, generated server-side |

**IAM policy for API/Worker:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"],
      "Resource": "arn:aws:s3:::careerdock-resumes/*"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:PutObject", "s3:GetObject"],
      "Resource": "arn:aws:s3:::careerdock-logos/*"
    }
  ]
}
```

**`careerdock-logos` (public via CloudFront):**

| Setting | Value |
|---------|-------|
| Public access | Via CloudFront OAC (Origin Access Control) only |
| Direct S3 access | Blocked (bucket policy allows only CloudFront) |
| Encryption | SSE-S3 |

### 8.6 Nginx Hardening

```nginx
# infra/docker/nginx.conf

# Force HTTPS
server {
    listen 80;
    server_name api.careerdock.skriptvalley.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.careerdock.skriptvalley.com;

    # SSL (Let's Encrypt via certbot)
    ssl_certificate     /etc/letsencrypt/live/api.careerdock.skriptvalley.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.careerdock.skriptvalley.com/privkey.pem;

    # TLS configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # OCSP stapling
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security headers (from §2.1)
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=(), payment=()" always;
    server_tokens off;

    # Request size limits
    client_max_body_size 6m;                    # 5MB PDF + multipart overhead
    large_client_header_buffers 4 8k;

    # Timeouts
    proxy_connect_timeout 10s;
    proxy_read_timeout    60s;                  # AI operations can take up to 60s
    proxy_send_timeout    30s;

    # Proxy to Go API
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
    }

    # Health check (no auth, no rate limit)
    location = /api/health {
        proxy_pass http://127.0.0.1:8080;
    }

    # Block everything else
    location / {
        return 404;
    }
}
```

---

## 9. Secrets Management

### 9.1 Secret Inventory

| Secret | Environment Variable | Rotation Frequency | Notes |
|--------|---------------------|-------------------|-------|
| JWT signing key | `JWT_SECRET` | Annually (or on compromise) | 256-bit minimum. Dual-key rotation (§6.4) |
| JWT previous key | `JWT_SECRET_PREVIOUS` | Removed 7 days after rotation | Only present during rotation window |
| Database password | `DATABASE_URL` | Annually | RDS master password |
| Redis password | `REDIS_PASSWORD` | Annually | Set via `requirepass` |
| Claude API key | `CLAUDE_API_KEY` | Per Anthropic policy | Regenerate on suspected leak |
| OpenAI API key | `OPENAI_API_KEY` | Per OpenAI policy | Fallback provider |
| Razorpay key ID | `RAZORPAY_KEY_ID` | Annually | Separate test/live keys |
| Razorpay key secret | `RAZORPAY_KEY_SECRET` | Annually | Never log this value |
| Razorpay webhook secret | `RAZORPAY_WEBHOOK_SECRET` | Annually | Used for HMAC-SHA256 verification |
| Resend API key | `RESEND_API_KEY` | Annually | Transactional email |
| S3 access key | `S3_ACCESS_KEY_ID` | Annually | Scoped IAM user |
| S3 secret key | `S3_SECRET_ACCESS_KEY` | Annually | Scoped IAM user |
| Sentry DSN | `SENTRY_DSN` | Never (not sensitive) | Project identifier, not a secret |

### 9.2 Storage by Environment

| Environment | Storage | Access |
|-------------|---------|--------|
| Local dev | `.env` file (gitignored) | Developer's machine |
| Staging | AWS Secrets Manager | ECS task role / EC2 instance profile |
| Production | AWS Secrets Manager | ECS task role / EC2 instance profile |

### 9.3 `.env` Protection

```gitignore
# .gitignore
.env
.env.local
.env.*.local
*.pem
*.key
```

**Pre-commit hook** (from CODE-STRUCTURE.md) uses `detect-secrets`:

```bash
# One-time baseline setup
detect-secrets scan > .secrets.baseline

# Hook checks for new secrets on every commit
# If a real secret is detected, the commit is blocked
```

### 9.4 What Must Never Be Logged

The structured logger (slog) must never include:

| Data | Why |
|------|-----|
| Passwords (plain or hashed) | Credential exposure |
| JWT tokens (access or refresh) | Session hijacking |
| API keys (Claude, OpenAI, Razorpay, etc.) | Third-party account takeover |
| Full resume text | PII exposure (names, addresses, phones) |
| Credit card data | N/A (Razorpay handles this), but guard against accidental logging |
| S3 signed URLs | Contain temporary credentials |
| Password reset tokens | Account takeover |

**Enforcement:** Code review + custom slog handler that redacts fields matching known secret patterns.

```go
// internal/middleware/logger.go — redaction example

var sensitiveFields = map[string]bool{
    "password": true, "token": true, "secret": true,
    "authorization": true, "cookie": true, "api_key": true,
}

// Custom slog handler that replaces sensitive field values with "[REDACTED]"
```

---

## 10. Dependency Security

### 10.1 CI Vulnerability Scanning

**Go (backend):**

```yaml
# In .github/workflows/ci.yml — add to backend job

- name: Vulnerability scan
  working-directory: backend
  run: |
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...
```

**Node.js (frontend):**

```yaml
# In .github/workflows/ci.yml — add to frontend job

- name: Vulnerability scan
  working-directory: frontend
  run: npm audit --audit-level=high
```

### 10.2 Automated Dependency Updates

Enable GitHub Dependabot:

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/backend"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "backend"

  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "frontend"

  - package-ecosystem: "docker"
    directory: "/infra/docker"
    schedule:
      interval: "monthly"
    labels:
      - "dependencies"
      - "infra"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "monthly"
    labels:
      - "dependencies"
      - "ci"
```

### 10.3 Dependency Review Process

1. Dependabot opens PR with version bump.
2. CI runs lint + test + build + vulnerability scan.
3. Review changelog for breaking changes.
4. Merge if CI passes and no breaking changes.
5. For major version bumps: manual review required.

---

## 11. Data Protection & Privacy

### 11.1 Data Classification

| Classification | Examples | Access | Encryption |
|---------------|----------|--------|------------|
| **Public** | Company directory, company logos | Anyone | In transit (HTTPS) |
| **Internal** | Feature flags, admin audit log | Admin only | At rest (RDS) + in transit |
| **Confidential** | User email, name, resume text, parsed data | User + admin | At rest (RDS/S3) + in transit |
| **Restricted** | Password hashes, JWT secrets, API keys | System only | At rest + in transit |

### 11.2 User Data Lifecycle

```
Registration
  └── Stored: email, name, password_hash

Premium Purchase
  └── Stored: payment record (no card data), credits allocated

Resume Upload
  └── Stored: PDF in S3, extracted_text + parsed_data in Postgres

Account Deletion Request
  ├── Immediate: Soft delete (deleted_at set), all tokens invalidated
  ├── 30 days: Hard delete of user + cascaded data (lists, credits, transactions)
  └── 90 days: S3 resume PDFs purged (cleanup job)
```

### 11.3 Data Export

Users can export their data via `GET /api/settings/export` (authenticated):

```json
{
  "user": {
    "email": "user@example.com",
    "name": "Jane Doe",
    "role": "user",
    "premium_since": "2026-01-15T10:00:00Z",
    "created_at": "2025-12-01T08:30:00Z"
  },
  "lists": [ ... ],
  "resumes": [
    {
      "filename": "jane-doe-resume.pdf",
      "uploaded_at": "2026-01-20T14:00:00Z",
      "parsed_data": { ... },
      "download_url": "https://...signed-url..."
    }
  ],
  "ats_checks": [ ... ],
  "credits": { ... },
  "payments": [ ... ]
}
```

Resume PDFs are included as signed download URLs (15-minute expiry). No credit card data is included (Razorpay handles all card data).

### 11.4 PII Minimisation

| Principle | Implementation |
|-----------|---------------|
| Collect only what's needed | Registration: email + name + password. No phone, no address |
| Don't store what you don't need | Card data never touches CareerDock (Razorpay handles it) |
| Limit access | User data scoped by user_id in all queries. Admin access logged |
| Limit retention | Soft delete → hard delete cascade (30 days). S3 cleanup (90 days) |

---

## 12. Incident Response

Lightweight incident response plan for a solo-founder MVP.

### 12.1 Severity Levels

| Level | Definition | Response Time | Example |
|-------|-----------|:---:|---------|
| **P1 — Critical** | Data breach, full system down, payment compromise | Immediate | Database exposed, API keys leaked, payment double-charge |
| **P2 — High** | Partial outage, auth bypass, data loss risk | < 4 hours | Login broken, resume uploads failing, credit deduction bug |
| **P3 — Medium** | Degraded performance, non-critical bug, single user affected | < 24 hours | Slow API response, one user's ATS check stuck, Sentry spike |

### 12.2 Response Playbook

**For any incident:**

1. **Detect** — Sentry alert, uptime monitor, user report, or log anomaly.
2. **Assess** — Determine severity (P1/P2/P3) and blast radius.
3. **Contain** — Stop the bleeding:
   - P1: Take affected service offline if necessary.
   - Rotate any compromised credentials immediately.
   - Force-expire all sessions: `redis-cli DEL session:*`
4. **Fix** — Deploy hotfix (tag, push, deploy).
5. **Verify** — Confirm fix via monitoring + manual testing.
6. **Postmortem** — Write a brief incident report in `docs/ADR/` (what happened, root cause, fix, prevention).

### 12.3 Key Emergency Actions

| Scenario | Immediate Action |
|----------|-----------------|
| JWT secret compromised | Rotate `JWT_SECRET` (§6.4). All sessions invalidated |
| Database credentials leaked | Rotate RDS password. Redeploy API + worker |
| API key (Claude/OpenAI) leaked | Regenerate key in provider dashboard. Update Secrets Manager. Redeploy |
| Razorpay keys leaked | Regenerate in Razorpay dashboard. Update webhook secret. Redeploy |
| Resume data breach | Rotate S3 credentials. Invalidate all signed URLs (they expire in 15 min). Assess scope. Notify affected users |
| DDoS attack | Enable AWS Shield Basic (free, auto-enabled). Consider CloudFront in front of API. Rate limits provide first defence |

---

## 13. Security Checklist — Sprint 0

Before going live, verify every item:

### Authentication & Sessions
- [ ] JWT tokens in httpOnly/Secure/SameSite=Strict cookies
- [ ] Access token: 15-minute TTL
- [ ] Refresh token: 7-day TTL with rotation
- [ ] Logout blacklists refresh token in Redis
- [ ] Password change invalidates all sessions
- [ ] bcrypt cost factor = 12
- [ ] Brute-force lockout active (5 attempts / 15 min)
- [ ] Generic error on failed login ("Invalid email or password")

### API Security
- [ ] CORS: explicit origin allowlist (no wildcards)
- [ ] Rate limiting active on all endpoints
- [ ] All input validated (handler layer)
- [ ] JSON body size limited (1MB)
- [ ] Resume upload: PDF magic bytes checked, 5MB max
- [ ] SQL: all queries parameterised (no string concatenation)

### Headers
- [ ] X-Frame-Options: DENY
- [ ] X-Content-Type-Options: nosniff
- [ ] Strict-Transport-Security present
- [ ] Content-Security-Policy present
- [ ] Nginx `server_tokens off`

### Infrastructure
- [ ] HTTPS everywhere (Let's Encrypt for API, Vercel for frontend)
- [ ] SSH restricted to admin IP only
- [ ] RDS: no public access, encryption at rest, `sslmode=require`
- [ ] Redis: `requirepass` set, bound to localhost
- [ ] S3: block public access on resume bucket
- [ ] Docker containers run as non-root user

### Secrets
- [ ] `.env` in `.gitignore`
- [ ] `detect-secrets` pre-commit hook active
- [ ] No secrets in logs (redaction handler active)
- [ ] Production secrets in AWS Secrets Manager

### Dependencies
- [ ] `govulncheck` in CI
- [ ] `npm audit` in CI
- [ ] Dependabot enabled

### Data Protection
- [ ] User soft-delete works (30-day grace)
- [ ] Data export endpoint functional
- [ ] S3 cleanup job scheduled (90-day post-deletion)
- [ ] Admin audit log captures all admin actions

---

## 14. Deferred to v2

| Item | Reason |
|------|--------|
| MFA (TOTP/WebAuthn) | Complexity. Add for admin accounts first, then optionally for all users |
| HaveIBeenPwned password checking | Nice-to-have. k-anonymity API is easy to integrate later |
| AWS WAF / Shield Advanced | Cost. Shield Basic (free) + rate limiting is sufficient for MVP |
| Penetration testing | Schedule after launch when features are stable |
| SOC 2 compliance | Only relevant at enterprise scale |
| Content Security Policy reporting (`report-uri`) | Set up after launch to monitor violations without blocking |
| Redis TLS | Only needed if Redis moves off localhost (e.g., ElastiCache) |
| Database IAM authentication | Simpler with password for MVP. IAM auth adds cold-start latency |
