# ADR 001 — UUID v7 for Primary Keys

**Status:** Accepted
**Date:** 2026-03-12

## Context

We need a primary key strategy for all database tables. Options considered:
- Auto-increment integers (SERIAL/BIGSERIAL)
- UUID v4 (random)
- UUID v7 (time-sorted)
- ULID

## Decision

Use UUID v7 (RFC 9562) for all primary keys, generated in the application layer via `uuid.NewV7()` from the `google/uuid` Go library.

## Consequences

**Pros:**
- Time-sortable: natural ordering by creation time without a separate `created_at` index for listing queries.
- No coordination required: IDs generated client-side, safe for distributed systems and bulk inserts.
- Not enumerable: prevents IDOR attacks unlike sequential integers.
- Standard 128-bit format: native `uuid` type in PostgreSQL with excellent index support.

**Cons:**
- 36-character string representation is longer than integers in URLs and logs.
- Slightly larger storage (16 bytes vs 4/8 bytes for int/bigint), negligible at our scale.
- Requires Go 1.22+ for `uuid.NewV7()` support (we use Go 1.25).
