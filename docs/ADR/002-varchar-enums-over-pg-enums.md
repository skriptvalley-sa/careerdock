# ADR 002 — VARCHAR + CHECK over PostgreSQL Enums

**Status:** Accepted
**Date:** 2026-03-12

## Context

Database columns like `role`, `status`, and `credit_type` need constrained values. Options:
- PostgreSQL `CREATE TYPE ... AS ENUM` (native enums)
- `VARCHAR` with `CHECK` constraints
- Integer codes with application-level mapping

## Decision

Use `VARCHAR` columns with `CHECK` constraints for all enum-like fields. Enum values are defined as Go string constants in `domain/enums.go`.

Example: `status VARCHAR(20) NOT NULL CHECK (status IN ('applied', 'interview', 'offer', ...))`

## Consequences

**Pros:**
- Adding new values is a simple `ALTER TABLE ... DROP CONSTRAINT ... ADD CONSTRAINT` — no `ALTER TYPE` needed, which requires a full table rewrite in PostgreSQL.
- Values are human-readable in the database without joins or lookups.
- Simpler migration story: CHECK constraints are easier to manage in golang-migrate.
- No coupling between Go constants and PostgreSQL type definitions.

**Cons:**
- No compile-time database-level type safety (mitigated by Go constants + application validation).
- Slightly more storage than integer codes (negligible).
- CHECK constraints must be manually kept in sync with Go constants.
