# Feedback Status Tracker

> Sources: `ai/feedback/session_01.md`, `ai/feedback/session_02.md`, `ai/feedback/session_03.md`, `ai/feedback/session_04.md`
> Branch: `feature/sprint-2-lists-tracking`
> Last updated: 2026-03-14 (Session 01 Batch 1–6 complete, Session 02 Batch 1–7 complete, Session 03 Batch 1–6 complete, Session 04 Batch 1–8 complete)

## Legend

| Status | Meaning |
|--------|---------|
| `[ ]`  | Pending |
| `[~]`  | In progress |
| `[x]`  | Complete |
| `[D]`  | Deferred (tracked for later sprint) |

---

## Critical Bugs

| ID | Task | Status | Notes |
|----|------|--------|-------|
| F-BUG-01 | API client: auto-refresh access token on 401 for all requests (not just session check) | `[x]` | Added `fetchWithAuth()` wrapper with 401 interceptor, concurrent refresh lock, and `setAuthFailureHandler` callback. |
| F-BUG-02 | Stale session UI: clear auth state when session expires | `[x]` | AuthProvider now re-checks session on `window.focus`. API client calls `onAuthFailure` to clear Zustand state when refresh fails. |

## UI/UX — High Priority

| ID | Task | Status | Notes |
|----|------|--------|-------|
| F-UX-01 | Dark mode theme (global) | `[x]` | All pages converted. Custom Tailwind tokens (bg-surface, bg-card, bg-overlay, border-edge etc.) in globals.css. 15+ files updated. |
| F-UX-02 | Sidebar: sticky user profile at bottom | `[x]` | Sidebar uses `flex flex-col`, nav has `flex-1 overflow-y-auto`, user card sits at bottom naturally (removed absolute positioning). |
| F-UX-03 | Remove Next.js dev toolbar ("N" button) | `[x]` | `devIndicators: false` in next.config.ts. |
| F-UX-04 | Homepage: different behaviour when logged in | `[x]` | Homepage redirects to `/dashboard` when authenticated. Logo links to `/dashboard` when logged in. |
| F-UX-05 | Navigation redesign: header + sidebar | `[x]` | Blue name button replaced with "Sign out" button (LogOut icon). Sign-out removed from dashboard page (now in header). |
| F-UX-06 | Add company to list by name search (not UUID) | `[x]` | Created `CompanyCombobox` with debounced search via `useCompanySearch` hook. List detail page now uses combobox instead of UUID input. Table shows company names via batch `GetNamesByIDs` resolution. |

## UI/UX — Medium Priority

| ID | Task | Status | Notes |
|----|------|--------|-------|
| F-UX-07 | Modern chip-style company filters | `[x]` | Replaced dropdown selects with inline toggleable chip-style filters. Multi-select arrays for sizes/compensation tiers. Color-coded tier chips, RSU toggle, clear button. |
| F-UX-08 | Company detail page: show user's applications | `[x]` | Added "Your Applications" section to company detail page (auth-aware). Uses `useEntriesByCompany` hook + `GET /api/entries?company_id=`. Inline status editing. Backend: `ListEntriesByCompanyID` repo method. |
| F-UX-09 | New Applications page | `[x]` | Created `/applications` page with status filter chips, cross-list entry table, inline status editing. Dashboard funnel cards now link here. Backend: `ListAllEntries` repo method + unified `/api/entries` handler. |
| F-UX-10 | Company filters: add office mode + locations | `[x]` | Office modes: implemented in Session 04 (migration 000019, `office_modes TEXT[]`). Locations: deferred — needs data sourcing. |

## Pre-Sprint-3 (Payments Prep)

| ID | Task | Status | Notes |
|----|------|--------|-------|
| F-PAY-01 | Feature flag system for payments | `[x]` | Full feature flag system: `FeatureFlagRepo` (GetByKey, List, Update), `FeatureFlagService` (IsEnabled, Toggle), admin handler (`GET/PUT /api/admin/feature-flags`). Seed migration inserts `payments_enabled` and `premium_bypass` flags. Wired into Services container and admin routes. |

## Admin Panel Enhancements

| ID | Task | Status | Notes |
|----|------|--------|-------|
| F-ADMIN-01 | Admin: grant/revoke premium access per user | `[D]` | Admins can manually grant permanent premium, set custom action quotas. Bypass payment requirement for selected users. Defer to Sprint 3 alongside payment work. |
| F-ADMIN-02 | Admin: manage pricing & special offers | `[D]` | Manage pack pricing, user-category offers, sale periods. Defer to Sprint 3 alongside payment work. |

---

## Implementation Order

### Batch 1 — Bugs + Core Auth (do first, unblocks everything)
1. F-BUG-01 — API client 401 auto-refresh interceptor
2. F-BUG-02 — Stale session cleanup (window focus re-check, state clear)

### Batch 2 — Global Theme + Layout (touches everything, do early)
3. F-UX-01 — Dark mode theme (globals, tailwind config, all pages)
4. F-UX-02 — Sidebar sticky user profile
5. F-UX-03 — Disable Next.js dev toolbar

### Batch 3 — Navigation & Routing
6. F-UX-04 — Homepage logged-in behaviour
7. F-UX-05 — Header/sidebar navigation redesign

### Batch 4 — Company Interaction Improvements
8. F-UX-06 — Company search/add to list by name
9. F-UX-07 — Chip-style company filters
10. F-UX-08 — Company detail: show user applications

### Batch 5 — New Feature
11. F-UX-09 — Applications page with status filtering

### Batch 6 — Pre-Sprint-3
12. F-PAY-01 — Feature flag system

### Deferred to Sprint 3+
- F-UX-10 — Office mode + locations filters (needs migration + data)
- F-ADMIN-01 — Admin premium access control
- F-ADMIN-02 — Admin pricing management

---

## Session 02 Feedback

> Source: `ai/feedback/session_02.md`

### UI/UX

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S2-UX-01 | Sidebar: make navigation pane fixed (position: fixed) | `[x]` | Sidebar now uses `position: fixed` with `left-0 top-0 bottom-0`. Main content area offset with `ml-64`. |
| S2-UX-02 | Remove duplicate logo in sidebar | `[x]` | Removed CareerDock text from sidebar — header logo is sufficient. |
| S2-UX-03 | Company search: fix prefix/pattern matching ("goog" → "Google") | `[x]` | Backend `SearchCompanies` repo method updated to use `ILIKE '%' || $1 || '%'` pattern matching instead of exact match. |
| S2-UX-04 | Redesign "add companies to list" — company browser with multi-select | `[x]` | Created `CompanyBrowserPanel` with search, filters, highlighted card selection (no checkboxes), infinite scroll, and batch save. New `POST /api/lists/{id}/entries/batch` endpoint. |
| S2-UX-05 | Separate company status vs application status (data model change) | `[x]` | New `CompanyTrackingStatus` enum (marked, researching, applied, interviewing, offered, accepted, rejected). Migration adds `company_status` column, makes `role_title` nullable, adds UNIQUE(list_id, company_id). Frontend `CompanyStatusBadge` component. |
| S2-UX-06 | Company name in list links to company profile | `[x]` | Added `GetNameAndSlugsByIDs` across domain/repo/service/handler. Entry response now includes `company_slug`. Company names in list table are `<Link>` to `/companies/{slug}`. |
| S2-UX-07 | Company detail page: reorder sections, show list chips, overall status | `[x]` | Sections reordered: Description → Key Info → Tech Stack → Domains → Interview → Your Lists → Applications. Added clickable list pill chips. Overall status badge in header (highest priority across entries). |
| S2-UX-08 | Collapsible sidebar on ALL pages | `[x]` | Sidebar available on all dashboard pages. Collapse/expand toggle with animated width transition. Main content adjusts spacing dynamically. |
| S2-UX-09 | Multi-color neon/electro theme | `[x]` | Full neon theme: cyan (#00f0ff), magenta (#ff00e5), green (#39ff14), amber (#ffb800). CSS glow utilities, `.btn-neon` gradient button, `.card-neon-hover` effect. All blue references replaced across 20+ files. |

### Payments

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S2-PAY-01 | Credits as universal currency (design note) | `[x]` | Documented as design direction: credits map to actions like currency. To be detailed in Sprint 3 payments design. |

### Session 02 Implementation Order

| Batch | Items | Status |
|-------|-------|--------|
| 1 | S2-UX-01, S2-UX-02 — Layout/Navigation overhaul | `[x]` |
| 2 | S2-UX-03 — Company search prefix fix | `[x]` |
| 3 | S2-UX-05 — Data model: CompanyTrackingStatus | `[x]` |
| 4 | S2-UX-04, S2-UX-06 — Company browser + name links | `[x]` |
| 5 | S2-UX-07 — Company detail page improvements | `[x]` |
| 6 | S2-UX-09 — Multi-color neon theme | `[x]` |
| 7 | S2-PAY-01 — Bookkeeping + credits design note | `[x]` |

---

## Session 03 Feedback

> Source: `ai/feedback/session_03.md`

### Core Model

Lists are company curation tools, not application containers. Companies are the parent entity; applications are child records; lists are grouping/tracking buckets.

### Layout Fixes

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S3-UX-01 | Navigation pane hides the logo — sidebar overlaps header | `[x]` | Sidebar `top-0` → `top-14`, `h-screen` → `h-[calc(100vh-3.5rem)]`. Header `z-20` → `z-40`. |
| S3-UX-02 | Footer floats in middle on short pages | `[x]` | App-shell uses `flex flex-col` + `flex-1` on main. Fixed `4rem` → `3.5rem` (header is h-14). |

### List Editing & Core Model

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S3-UX-03 | List editing flow broken — "Add Companies" doesn't work properly | `[x]` | CompanyBrowserPanel reworked: existing companies pre-selected, toggle adds/removes, diff summary (+N added, -N removed), "Save Changes" calls sync endpoint. |
| S3-UX-04 | Remove action in wrong area (row-level) | `[x]` | Removed per-row "Remove" button. Removals now handled via edit flow (deselect in browser → "Save Changes"). |
| S3-UX-05 | Save action too narrow ("Add 1 Company") | `[x]` | Button now says "Save Changes". Footer shows diff: "+N added, -N removed". Disabled when no changes. |
| S3-MODEL | List detail page redesign — new columns: Company, Status, Applications count, +Add Application | `[x]` | New table: Company (linked), Status (inline editable), Applications (count badge), +Add Application (modal). Removed: Role, App Status, Notes, Actions columns. Button changed "Add companies" → "Edit List". Added `AddApplicationModal` component. |

### Applications Page

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S3-UX-06 | Applications page shows unapplied companies | `[x]` | Backend: `excludeNotApplied: true` in handler. Frontend: removed `not_applied` from filter chips and inline status dropdown. |

### Company Cards

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S3-UX-07 | Company cards: quick-add-to-list, remove description, promote domains/RSU | `[x]` | Description removed. Domains promoted above tech stack with magenta chips. RSU/Refresher now neon green/amber chips. Added `+` quick-add button (hover reveal, auth-only). `QuickAddToListModal` component with list toggle checkboxes. |

### Backend Endpoints

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S3-BE-01 | Sync list entries endpoint (`PUT /api/lists/{id}/entries/sync`) | `[x]` | Full desired set reconciliation via `SyncListEntries` service method with transaction. |
| S3-BE-02 | Filter `not_applied` from applications endpoint | `[x]` | Added `excludeNotApplied` bool to `ListAllEntries`. Handler passes `true` for all-entries mode. |
| S3-BE-03 | Lists containing a company endpoint (`GET /api/lists/by-company/{companyId}`) | `[x]` | `ListsWithCompanyFlag` repo → service → handler. Uses EXISTS subquery. |
| S3-BE-04 | Delete entry by company ID (`DELETE /api/lists/{id}/entries/by-company/{companyId}`) | `[x]` | `DeleteEntryByCompany` repo → service (with ownership check) → handler + route. |

### Deferred

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S3-D-01 | Office mode metadata (Remote/Hybrid/On-site) on company cards | `[D]` | Requires new DB column + migration + seed data |
| S3-D-02 | Company status chip on company cards (list membership indicator) | `[D]` | Requires batch-loading list membership for all visible companies |

### Session 03 Implementation Order

| Batch | Items | Status |
|-------|-------|--------|
| 1 | S3-UX-01, S3-UX-02 — Layout fixes (sidebar + footer) | `[x]` |
| 2 | S3-BE-01, S3-BE-02, S3-BE-03, S3-BE-04 — Backend endpoints | `[x]` |
| 3 | S3-UX-03, S3-UX-04, S3-UX-05, S3-MODEL — List page redesign + edit flow | `[x]` |
| 4 | S3-UX-06 — Applications page filter | `[x]` |
| 5 | S3-UX-07 — Company card enhancements + quick-add-to-list | `[x]` |
| 6 | Bookkeeping — Final FEEDBACK-STATUS update | `[x]` |

---

## Session 04 Feedback

> Source: `ai/feedback/session_04.md`

### Deferred Items Implemented (from Session 03)

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S3-D-01 | Office mode metadata on company cards | `[x]` | Migration 000019: added `office_modes TEXT[]` column with GIN index. Updated domain entity, repo (list/detail columns, scans, upsert), handler DTOs, seed struct. Seeded all 60 companies with office_modes. Frontend: added to `CompanyListItem` type, displayed as chip on company cards. |
| S3-D-02 | Company list membership indicator on cards | `[x]` | New `GET /api/lists/company-counts` endpoint. `CompanyListCounts` repo method (COUNT DISTINCT per company). `useCompanyListCounts()` hook. Company cards always show list indicator: count chip (cyan) if in lists, `+` button if not. Cache invalidation on add/remove. |

### UI/UX

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S4-UX-01 | Applications page: add company filter dropdown | `[x]` | Client-side company filter using `useMemo` to extract unique companies from entries. Dropdown styled as rounded-full select with ChevronDown icon. Works alongside status filter chips. Shows company count in footer summary. |
| S4-UX-02 | Company card chip layout improvements | `[x]` | RSU/Refresher anchored bottom-left. Office mode chip visible as metadata (bottom-right). List indicator always visible (not hover-only). Card uses `flex flex-col` with `flex-1` spacer. |
| S4-UX-03 | Add-to-list modal interaction refinement | `[x]` | Action button far-right, single-icon style. Not-in-list: empty Circle → green Plus (hover). In-list: green Check → red X (hover). Uses `group-hover/btn` Tailwind classes for transitions. |
| S4-UX-04 | Header fixed during fast upward scroll | `[x]` | Header changed from `sticky` to `fixed` positioning. Added `pt-14` wrapper in root layout. Added `overscroll-none` to body to prevent elastic bounce. |

### Implementation Issues

| ID | Task | Status | Notes |
|----|------|--------|-------|
| S4-BUG-01 | Fix duplicate React key errors on companies page | `[x]` | Root cause: `flatMap` across infinite query pages can produce duplicate company IDs. Fix: deduplication using `Set<string>` before rendering. API confirmed returning unique IDs. |

### Session 04 Implementation Order

| Batch | Items | Status |
|-------|-------|--------|
| 1 | S3-D-01 — DB migration + entity updates for office_modes | `[x]` |
| 2 | S3-D-02 — Backend list membership counts endpoint | `[x]` |
| 3 | S4-BUG-01 — Fix duplicate React key errors | `[x]` |
| 4 | S4-UX-02 — Company card improvements | `[x]` |
| 5 | S4-UX-03 — Add-to-list modal interaction refinement | `[x]` |
| 6 | S4-UX-01 — Applications page company filter | `[x]` |
| 7 | S4-UX-04 — Header fixed during fast scroll | `[x]` |
| 8 | Bookkeeping — Final FEEDBACK-STATUS update | `[x]` |
