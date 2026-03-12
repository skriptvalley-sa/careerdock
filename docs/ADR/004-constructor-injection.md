# ADR 004 — Constructor Injection over DI Framework

**Status:** Accepted
**Date:** 2026-03-12

## Context

The backend has a layered architecture (handler → service → repository) that requires dependency wiring. Options:
- DI frameworks (Wire, Dig, fx)
- Constructor injection (manual wiring in main.go)
- Global singletons

## Decision

Use manual constructor injection. All dependencies are wired explicitly in `cmd/api/main.go` and `cmd/worker/main.go` via constructor functions (`NewAuthService(...)`, `NewUserRepo(...)`, etc.).

A `service.Services` struct acts as the top-level container, holding all services for the handler layer.

## Consequences

**Pros:**
- Zero magic: the dependency graph is visible in a single file (`main.go`).
- Compile-time safety: missing dependencies cause build errors, not runtime panics.
- Easy to test: inject mocks directly via constructors.
- No framework lock-in or code generation step.
- IDE navigation works perfectly (Go to Definition on constructors).

**Cons:**
- `main.go` grows as services are added. Mitigated by the `Services` container pattern.
- Adding a new dependency requires updating the wiring in every entry point (API, worker). Acceptable for a solo project with 2 entry points.
