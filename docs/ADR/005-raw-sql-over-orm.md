# ADR 005 — Raw SQL (pgx) over ORM

**Status:** Accepted
**Date:** 2026-03-12

## Context

The repository layer needs a strategy for database access. Options:
- ORM (GORM, Ent)
- Query builder (Squirrel, goqu)
- Raw SQL with pgx

## Decision

Use raw SQL queries with `jackc/pgx/v5` as the PostgreSQL driver. No ORM or query builder.

Repositories use a shared `DBTX` interface that is satisfied by both `*pgxpool.Pool` (normal operations) and `pgx.Tx` (within transactions), enabling transparent transaction support.

## Consequences

**Pros:**
- Full control over queries: can optimise for performance (CTEs, window functions, partial indexes) without fighting an ORM abstraction.
- No hidden N+1 queries or unexpected SQL generation.
- pgx is the fastest Go PostgreSQL driver with native protocol support.
- SQL is readable and auditable — what you write is what runs.
- Easy to use PostgreSQL-specific features (JSONB, array types, LISTEN/NOTIFY).

**Cons:**
- More boilerplate: every query is hand-written SQL with manual `Scan()` calls.
- No automatic schema validation against structs (mitigated by tests).
- Migrations must be manually kept in sync with repository queries.
- No built-in migration generation from model changes.
