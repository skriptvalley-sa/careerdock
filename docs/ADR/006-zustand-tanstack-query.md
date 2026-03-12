# ADR 006 — Zustand + TanStack Query over Redux

**Status:** Accepted
**Date:** 2026-03-12

## Context

The Next.js frontend needs client-side state management for:
1. **Server state** (API data: companies, lists, resumes, ATS results)
2. **Client state** (auth session, UI preferences, form state)

Options:
- Redux Toolkit + RTK Query
- Zustand + TanStack Query
- React Context + SWR
- Jotai/Recoil + TanStack Query

## Decision

Use **Zustand** for client state and **TanStack Query** for server state.

- `store/auth-store.ts` — Zustand store for auth session (user, isAuthenticated, isPremium, etc.)
- TanStack Query for all API data with hierarchical query key factories (`lib/query-keys.ts`)
- React Hook Form + Zod for form validation

## Consequences

**Pros:**
- Zustand is tiny (~1KB), has zero boilerplate (no providers, reducers, or actions), and works seamlessly with Next.js App Router.
- TanStack Query handles caching, deduplication, background refetch, and optimistic updates out of the box — no custom caching logic needed.
- Clear separation: Zustand for synchronous client state, TanStack Query for async server state.
- Both libraries have excellent TypeScript support.
- No provider nesting hell (Zustand stores are hooks, not context).

**Cons:**
- Two libraries to learn instead of one unified solution (Redux).
- Zustand stores are module-level singletons — need care with SSR hydration (mitigated by only using Zustand in client components).
- No built-in devtools integration like Redux DevTools (TanStack Query has its own devtools).
