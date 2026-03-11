# CareerDock — Product Requirements Document (PRD)

> **Version:** 1.0
> **Status:** Final (Phase 1)
> **Last updated:** 2026-03-11

---

## 1. Product Overview

### 1.1 Vision

CareerDock is a career intelligence platform for tech job seekers in India. It combines a curated company directory with AI-powered resume analysis and ATS scoring to help candidates make smarter application decisions.

### 1.2 Target Users

- **Primary:** Tech professionals in India (software engineers, data scientists, DevOps, product managers) actively or passively job seeking.
- **Secondary:** Fresh graduates entering the tech job market.

### 1.3 Core Value Proposition

1. **Discover:** Browse a curated, admin-maintained directory of 200-500 Indian tech companies with structured data (tech stacks, interview patterns, compensation bands).
2. **Track:** Organise target companies into lists and track application progress across roles and rounds.
3. **Optimise:** Upload resumes, get AI-powered ATS scores (general, company-specific, job-specific), and receive actionable improvement suggestions.
4. **Match:** Get AI-curated company recommendations based on resume profile and preferences.

---

## 2. User Roles

| Role | Description | Access Level |
|------|-------------|-------------|
| **Visitor** | Unauthenticated user | Browse company directory, view company profiles |
| **Free User** | Registered, email-verified | All visitor features + create up to 3 company lists, application tracking |
| **Premium User** | Purchased Starter Pack or actions | All free features + resume uploads, ATS scoring, AI-curated lists, CV generation |
| **Moderator** | Appointed by admin | All premium features + suggest edits to company profiles (goes to admin review queue) |
| **Admin** | Platform operator | Full system access: user management, company CRUD, financial dashboard, feature flags, moderation queue |

---

## 3. Feature Specifications

### 3.1 Company Directory (Public — No Auth Required)

#### 3.1.1 Company Profile

Each company has a structured profile containing:

| Field | Type | Notes |
|-------|------|-------|
| Name | String | Required |
| Logo | Image (URL/S3) | Optional, fallback to generated initial |
| Description | Text | Short overview of the company |
| Size | Enum | Startup (<50), Small (50-200), Mid (200-1000), Large (1000-5000), Enterprise (5000+) |
| Headquarters | String | City, State |
| Founded Year | Integer | Optional |
| Careers Page URL | URL | Link to company's careers/jobs page |
| Glassdoor URL | URL | Optional |
| AmbitionBox URL | URL | Optional |
| LinkedIn URL | URL | Optional |
| Tech Stack | String[] | Languages, frameworks, infra, cloud providers |
| Domains | String[] | e.g., fintech, SaaS, infra, edtech, healthtech |
| Hiring Status | Enum | Active / Paused / Unknown |
| Interview Patterns | JSON (structured) | See schema below |
| Compensation Tier | Enum | Tier 1 (FAANG-level), Tier 2 (Top MNCs), Tier 3 (Solid MNCs), Tier 4 (Indian Product Cos) |
| RSU | Boolean | Whether the company grants RSUs |
| RSU Refresher | Boolean | Whether annual RSU refreshers are standard |
| Compensation Bands | JSON (structured) | See schema below |
| Last Verified | Timestamp | When data was last reviewed/refreshed |

**Interview Patterns Schema:**
```json
{
  "roles": [
    {
      "title": "SDE-1",
      "total_rounds": 4,
      "rounds": [
        {
          "type": "Online Assessment",
          "difficulty": "Medium",
          "topics": ["DSA", "Problem Solving"],
          "duration_minutes": 90
        },
        {
          "type": "Technical Interview",
          "difficulty": "Medium-Hard",
          "topics": ["DSA", "System Design Basics"],
          "duration_minutes": 60
        }
      ],
      "typical_timeline_days": 14,
      "notes": "Optional free-text notes"
    }
  ]
}
```

**Compensation Bands Schema:**
```json
{
  "roles": [
    {
      "title": "SDE-1",
      "min_ctc_lakhs": 8,
      "max_ctc_lakhs": 15,
      "equity_component": "RSU",
      "equity_notes": "Vests over 4 years, annual refreshers",
      "currency": "INR",
      "source": "estimate",
      "last_updated": "2026-01"
    }
  ]
}
```

**Compensation Tier Definitions:**

| Tier | Label | Examples | Typical Range (SDE-2) |
|------|-------|---------|----------------------|
| Tier 1 | FAANG-level | Google, Microsoft, Amazon, Apple, Meta | ₹40-80+ LPA |
| Tier 2 | Top MNCs (RSU + refreshers) | Adobe, Salesforce, Atlassian, Uber, LinkedIn | ₹30-55 LPA |
| Tier 3 | Solid MNCs / good equity | MongoDB, Cloudflare, Rubrik, Databricks | ₹25-45 LPA |
| Tier 4 | Indian product companies | Flipkart, PhonePe, Razorpay, CRED, Freshworks | ₹20-40 LPA |

#### 3.1.2 Company Data Management

- **Source:** All company data is admin-curated. No community contributions.
- **Initial seed:** 50-100 companies, generated as structured JSON and reviewed manually.
- **Ongoing additions:** Admins add new companies via the admin dashboard. On addition, the company profile is enriched using Claude API (tech stack verification, interview pattern research, etc.).
- **Data refresh:** Admin-triggered AI refresh per company. Browses the company's public information using Claude API and updates the profile.
- **Moderator workflow:** Moderators (appointed by admin) can suggest edits to company profiles. Suggestions go into a moderation queue for admin approval/rejection. Only moderators can suggest edits — regular users cannot.
- **No auto-scraping** in MVP. All data flows through admin review.

#### 3.1.3 Browsing & Search

- **Search:** Full-text search across company names, tech stacks, domains.
- **Filters:** Tech stack, domain, company size, location (city), hiring status, compensation tier, RSU availability.
- **Sort:** Name (A-Z), size, compensation tier, recently updated.
- **Pagination:** Cursor-based for the company list API.
- **SEO:** Company profile pages are server-side rendered (Next.js SSR) with proper meta tags, structured data (JSON-LD), and clean URLs (`/companies/{slug}`).
- **Client-side caching:** Service worker + IndexedDB for offline browsing of previously viewed companies and search results. Cache invalidation via ETags or last-modified headers.

---

### 3.2 Authentication & User Management

#### 3.2.1 Registration & Login (MVP)

- **Email + password** registration with email verification.
- Email verification required before accessing authenticated features (lists, paid features).
- Password requirements: minimum 8 characters, at least 1 uppercase, 1 lowercase, 1 number.
- Forgot password flow via email reset link (time-limited, single-use token).

#### 3.2.2 Session Management

- JWT access token (short-lived, ~15 minutes) + refresh token (longer-lived, ~7 days).
- Tokens stored in **httpOnly, Secure, SameSite=Strict cookies only**. No localStorage.
- Refresh token rotation: each refresh issues a new refresh token and invalidates the old one.
- Logout invalidates refresh token server-side.

#### 3.2.3 User Profile

- Fields: Name, email, current role/title, experience level (Fresher/Junior/Mid/Senior/Staff+), preferred tech stacks, target domains, target locations.
- Profile data is used for AI matching and curated list generation.

#### 3.2.4 Account Management

- Change password (requires current password).
- Delete account (soft delete with 30-day grace period, then hard delete).
- View payment/action history.

#### 3.2.5 Deferred to v2

- OAuth (Google) login.
- Magic link login.
- Multi-factor authentication.

---

### 3.3 Company Lists & Application Tracking (Free Feature)

#### 3.3.1 Lists

- Free users can create up to **3 lists**. Premium users get **5 lists** (3 free + 2 bonus from Starter Pack).
- Each list has: name (user-defined), description (optional), created/updated timestamps.
- Lists can be renamed, reordered, and deleted.
- Companies are added to lists by browsing/searching the directory and selecting "Add to List."
- A company can appear in multiple lists.

#### 3.3.2 Application Tracking

Each entry in a list is a **company + role** pair (not just a company). A user can have multiple entries for the same company with different roles.

**Fields per entry:**

| Field | Type | Notes |
|-------|------|-------|
| Company | FK to companies | Required |
| Role/Position | String (free text) | Required |
| Application Status | Enum | See status flow below |
| Date Applied | Date | Optional, user-set |
| Notes | Text | Free-form notes |
| Created At | Timestamp | Auto |
| Updated At | Timestamp | Auto |

**Application Status Flow:**

```
Not Applied → Applied → Phone Screen → Interview → Offer → Accepted
                                          ↓           ↓
                                       Rejected    Withdrawn
```

Valid statuses: `Not Applied`, `Applied`, `Phone Screen`, `Interview`, `Offer`, `Rejected`, `Accepted`, `Withdrawn`.

#### 3.3.3 Status History & Round Tracking

- Every status change is logged with timestamp (audit trail).
- Users can manually add interview rounds to an application entry:
  - Round number, round type (HR, Technical, System Design, Managerial, etc.), date, outcome (Passed/Failed/Pending), notes.
- Status history and round details are viewable per application entry.

#### 3.3.4 Free Dashboard

- Overview: Total companies tracked across lists, applications by status (funnel visualization), recent activity feed.
- Quick links to each list.
- No AI features on free dashboard.

#### 3.3.5 Deferred to Later Releases

- Follow-up reminders / calendar integration.
- Bulk import/export of lists.
- Interview round-level analytics.

---

### 3.4 Paid Features (Premium)

#### 3.4.1 Payment Model

**No subscriptions. No recurring payments.** All purchases are one-time.

**Starter Pack — ₹399 (one-time)**

Unlocks all premium features and includes an initial bundle of actions:

| Feature | Included Quantity |
|---------|------------------|
| Resume uploads | 3 initial + 6 re-uploads = **9 total upload actions** |
| General ATS score + suggestions | **Automatic** on each upload (included) |
| AI-curated company lists | **3 list generations** |
| Company-specific ATS checks | **10 checks** |
| Job-specific ATS checks | **10 checks** |
| Bonus lists (beyond free 3) | **+2 lists** (5 total) |

**Notes on resume uploads:**
- Users can hold a maximum of 3 resumes at any time.
- Initial 3 uploads fill all 3 slots.
- Re-uploads replace an existing resume in a slot (the old resume is archived, the new one is analysed fresh).
- 9 total upload actions means: 3 initial + up to 6 replacements across all slots.

**À La Carte Purchases:**

| Action | Price | Notes |
|--------|-------|-------|
| Additional resume upload (1) | ₹49 | Adds 1 upload action to account |
| ATS check bundle (10) | ₹99 | Redeemable for any of the 3 ATS check types (general, company, job) |
| CV generation for JD | TBD | Generate tailored CV for a specific job description using selected resume. Pricing TBD after AI cost estimation. |
| Re-buy Init Bundle | ₹399 | Re-purchase the full Starter Pack bundle (all quantities reset/add) |

**Credit System:**
- All purchased actions are tracked as **credits** in the user's account.
- Credits are consumed on use. No expiry.
- Credit types: `resume_upload`, `ats_check`, `curated_list`, `cv_generation`.
- ATS check credits are fungible — usable for general, company-specific, or job-specific checks.
- Users can view remaining credits on their premium dashboard.

**Refund Policy:**
- No refunds on consumed actions/credits.
- Full refund available if no credits have been consumed, within 7 days of purchase.

#### 3.4.2 Resume Management

- **Format:** PDF only.
- **Max file size:** 5 MB per resume.
- **Max active resumes:** 3 at any time.
- **Storage:** Resumes stored in S3 (MinIO for local dev) with access control — no public URLs. Signed URLs (time-limited) for viewing/downloading.
- **Upload flow:**
  1. User uploads PDF.
  2. File is validated (PDF, <5MB).
  3. File is stored in S3.
  4. Async job is queued for AI analysis.
  5. AI extracts: skills, experience summary, years of experience, education, domains, role level.
  6. General ATS score is computed automatically (consumes 0 ATS credits — included with upload).
  7. Results are stored as structured JSON alongside the resume record.
- **Default resume:** One resume can be marked as "default" — used as the basis for AI-curated lists and as the pre-selected resume for ATS checks.
- **Re-upload:** Replaces an existing resume in a slot. The old resume is archived (metadata kept, S3 object retained for 90 days), and the new one is analysed fresh. Consumes 1 upload credit.

#### 3.4.3 ATS Scoring

All ATS operations are **asynchronous** (queued via job queue, results available when complete). The user is notified (in-app) when results are ready.

**Tier 1 — General Resume Score (Automatic on Upload)**
- Triggered automatically when a resume is uploaded or re-uploaded.
- Checks: formatting quality, keyword density, section completeness (summary, experience, education, skills), readability, ATS-parser friendliness.
- Output: Score (0-100), breakdown by category, actionable improvement suggestions.
- **No credit cost** — included with each upload action.

**Tier 2 — Company-Specific Score (Costs 1 ATS Credit)**
- User selects a company from the directory + the system evaluates all uploaded resumes.
- Scoring: Tech stack match, domain relevance, experience level fit, culture-fit keywords.
- Output: Score (0-100), gap analysis ("You're missing keywords X, Y, Z"), best resume recommendation from user's set, improvement suggestions.
- The system automatically determines which resume is the best match and presents results for all resumes with a recommendation.

**Tier 3 — Job-Specific Score (Costs 1 ATS Credit)**
- User **pastes job description text** (not URL) + the system evaluates all uploaded resumes.
- Scoring: Required skills match, experience level fit, keyword alignment, qualifications gap.
- Output: Score (0-100), required vs. possessed skills matrix, best resume recommendation, keyword suggestions, improvement suggestions.
- **No URL fetching in MVP.** Users paste JD text directly.

#### 3.4.4 AI-Curated Company Lists

- Based on the **default resume's** extracted profile + user's preferences (from their profile: target domains, locations, tech stacks).
- The AI analyses the user's resume profile against all companies in the directory.
- Output: A ranked list of best-matching companies with match scores and reasoning.
- Runs **asynchronously** (may take 30-60 seconds for large directories).
- Consumes 1 curated list credit per generation.
- User can regenerate (consumes another credit) — useful after updating default resume or preferences.

#### 3.4.5 CV Generation (Future À La Carte Feature)

- User selects a resume + provides a job description (paste text).
- AI generates a tailored CV optimised for the specific JD.
- Output: Downloadable PDF.
- Consumes 1 CV generation credit.
- **Pricing TBD** after AI cost estimation in Phase 2.

#### 3.4.6 Premium Dashboard

- **Resume health overview:** All uploaded resumes with their general ATS scores at a glance.
- **AI-curated list summary:** Top matched companies with match scores.
- **Credit tracker:** Remaining credits by type (uploads, ATS checks, curated lists, CV generations).
- **ATS check history:** Log of all past ATS checks with scores, sortable/filterable.
- **Recommendations:** "Your resume scores low against Company X — consider updating these keywords."

---

### 3.5 Payment Integration

#### 3.5.1 Payment Gateway

**Razorpay** — India-first, native ₹ support, UPI, cards, net banking, wallets.

#### 3.5.2 Payment Flow

```
User selects purchase → Frontend creates order via API → Backend creates Razorpay order →
Frontend opens Razorpay checkout → User completes payment → Razorpay sends webhook →
Backend verifies webhook signature → Credits allocated to user account →
Confirmation shown to user
```

#### 3.5.3 Payment Requirements

- All amounts in INR.
- Idempotent order creation (prevent duplicate charges).
- Webhook signature verification (Razorpay's standard mechanism).
- Payment receipt/invoice generated and viewable in account settings.
- Transaction log stored for admin financial dashboard.
- **GST considerations:** Prices may need to be GST-inclusive or clearly marked. Discuss during implementation.

#### 3.5.4 Edge Cases

- **Failed payment:** No credits allocated. User can retry.
- **Duplicate webhook:** Idempotency key prevents double credit allocation.
- **Refund:** Admin-initiated only via admin dashboard. Credits are deducted on refund. Only if no credits consumed within 7 days.

---

## 4. Technical Constraints & Requirements

### 4.1 Performance

- Company directory page load: <2 seconds (SSR + CDN).
- API response times: <200ms for CRUD operations, <500ms for search.
- AI operations: Async — no blocking. Target <60 seconds for results.
- Client-side caching for offline company browsing.

### 4.2 Security

- HTTPS everywhere.
- JWT in httpOnly cookies only. No localStorage for tokens.
- Resume files: Private S3 bucket, signed URLs for access (15-minute expiry).
- Rate limiting on all endpoints, stricter on AI and payment endpoints.
- Input validation and sanitisation on all user inputs.
- Razorpay webhook signature verification.

### 4.3 Scalability (MVP Target)

- Handle ~1,000 concurrent users.
- Company directory: 200-500 companies.
- Designed for horizontal scaling but single-instance deployment for MVP.

### 4.4 Data Privacy

- Users can delete their account (soft delete, 30-day grace, then hard delete including S3 objects).
- Resume data is personal — strict access control, no sharing between users.
- No analytics/tracking without consent. Privacy-first cookie policy.

---

## 5. AI Operations Summary

| Operation | Trigger | Async? | Credit Cost | Provider |
|-----------|---------|--------|-------------|----------|
| Resume parsing & extraction | Resume upload | Yes (job queue) | Included with upload | Claude (primary), OpenAI (fallback) |
| General ATS score | Resume upload | Yes (job queue) | Included with upload | Claude (primary), OpenAI (fallback) |
| Company ATS score | User-initiated | Yes (job queue) | 1 ATS credit | Claude (primary), OpenAI (fallback) |
| Job ATS score | User-initiated | Yes (job queue) | 1 ATS credit | Claude (primary), OpenAI (fallback) |
| AI-curated company list | User-initiated | Yes (job queue) | 1 curated list credit | Claude (primary), OpenAI (fallback) |
| CV generation | User-initiated | Yes (job queue) | 1 CV generation credit | Claude (primary), OpenAI (fallback) |
| Company profile enrichment | Admin-initiated | Yes (job queue) | N/A (platform cost) | Claude API |
| Company data refresh | Admin-initiated | Yes (job queue) | N/A (platform cost) | Claude API |

**AI Cost Management:**
- All AI results are cached aggressively (resume analysis is deterministic per file, ATS scores are deterministic per resume+company/JD pair).
- LLM provider is abstracted behind a Go interface — Claude is primary, OpenAI is fallback.
- Token budgets per operation will be estimated during Architecture phase.
- Admin dashboard shows AI cost tracking (tokens used, cost per operation, monthly spend).

---

## 6. Information Architecture

### 6.1 Public Pages (No Auth)

- `/` — Landing page with value proposition, CTA to browse directory.
- `/companies` — Company directory with search, filter, sort.
- `/companies/{slug}` — Company profile page (SSR, SEO).
- `/login` — Login page.
- `/register` — Registration page.
- `/forgot-password` — Password reset request.
- `/reset-password/{token}` — Password reset form.
- `/pricing` — Starter Pack and à la carte pricing overview.

### 6.2 Authenticated Pages (Free + Premium)

- `/dashboard` — User dashboard (free: lists overview, funnel; premium: + resume health, credits).
- `/lists` — List management.
- `/lists/{id}` — List detail with application tracking.
- `/settings` — Account settings, password change, payment history.
- `/settings/profile` — Edit user profile/preferences.

### 6.3 Premium Pages

- `/resumes` — Resume management (upload, view, set default).
- `/resumes/{id}` — Resume detail with extracted data and ATS score.
- `/ats` — ATS check interface (select type, company/JD, resume).
- `/ats/{id}` — ATS check result detail.
- `/curated-lists` — AI-curated company list results.

### 6.4 Admin Pages

- `/admin/dashboard` — Overview (users, revenue, system health).
- `/admin/users` — User management.
- `/admin/companies` — Company CRUD + moderation queue.
- `/admin/payments` — Transaction log, refund management.
- `/admin/features` — Feature flags.
- `/admin/ai` — AI cost tracking, provider management.

---

## 7. MVP Scope vs. Future Releases

### 7.1 MVP (v1) — Included

- Company directory with search, filter, sort (50-100 companies).
- SSR company profile pages with SEO.
- Client-side caching for offline browsing.
- Email + password authentication.
- Up to 3 free lists with application tracking (company + role, status history, round tracking).
- Free dashboard (funnel view, activity feed).
- Razorpay one-time payments (Starter Pack + à la carte).
- Credit system for action tracking.
- Resume upload (PDF, 5MB max, 3 active slots).
- AI resume parsing and extraction.
- ATS scoring (3 tiers: general, company, job).
- AI-curated company lists.
- Premium dashboard with credit tracking.
- Admin panel (user management, company CRUD, payments, feature flags, moderation queue, AI cost tracking).
- Moderator role (suggest edits to company profiles).
- Structured logging, basic monitoring, error tracking.

### 7.2 v2 — Deferred

- OAuth (Google) login.
- Magic link login / MFA.
- CV generation for specific JDs.
- Follow-up reminders / calendar integration.
- Job description URL auto-fetch (currently paste-only).
- Bulk import/export of company lists.
- Interview round-level analytics.
- Community contributions to company data (beyond moderator role).
- Company data auto-scraping.
- Mobile app (React Native or PWA enhancement).
- Advanced search (MeiliSearch/Typesense — may be MVP if Postgres FTS is insufficient).

---

## 8. Success Metrics

| Metric | Target (3 months post-launch) |
|--------|-------------------------------|
| Registered users | 500+ |
| Starter Pack purchases | 50+ (10% conversion) |
| Companies in directory | 200+ |
| Average ATS checks per premium user | 8+ |
| Resume uploads | 100+ |
| User retention (monthly active) | 40%+ |
| Platform uptime | 99.5%+ |
| AI operation success rate | 98%+ |

---

## 9. Seed Data Reference

The initial company seed list is based on `ai/bangalore_companies_prompt.md` — a curated list of ~76 Bangalore tech companies across 4 compensation tiers. This will be expanded to 100+ companies during Sprint 1 seeding, covering:

- **Tier 1 (5):** Google, Microsoft, Amazon, Apple, Meta
- **Tier 2 (~22):** Adobe, Salesforce, Cisco, Visa, Atlassian, Uber, LinkedIn, etc.
- **Tier 3 (~31):** MongoDB, Cloudflare, Databricks, Rubrik, Samsung R&D, etc.
- **Tier 4 (~13):** Flipkart, PhonePe, Razorpay, CRED, Freshworks, etc.

**Standard domain tags:** `AI/ML`, `Cloud`, `Infra`, `Distributed`, `Platform`, `SaaS`, `FinTech`, `Security`, `Networking`, `Dev Tools`, `Embedded`, `Database`, `Storage`, `Automotive`

Additional companies beyond Bangalore (Hyderabad, Pune, NCR, Chennai) will be added post-launch to reach the 200-500 target.

---

## 10. Open Items for Phase 2

The following items need resolution during the Architecture phase:

1. **AI token budget estimation** — Cost per operation for Claude API (resume parse, ATS score, curated list, CV generation).
2. **CV generation pricing** — Depends on AI cost estimation.
3. **Search engine decision** — Postgres FTS vs. MeiliSearch/Typesense for company search.
4. **GST handling** — Whether prices are GST-inclusive or displayed separately.
5. **Company profile AI enrichment flow** — Exact Claude API integration for admin-triggered research.
6. **Notification system** — In-app notifications for async AI results (WebSocket vs polling vs SSE).
7. **Compensation tier definitions** — Finalise tier boundaries and whether tiers are admin-assigned or algorithmically derived from compensation data.
