# ADR 003 — Trunk-Based Development with Tag Releases

**Status:** Accepted
**Date:** 2026-03-12

## Context

As a solo founder project, we need a branching strategy that balances speed with safety. Options:
- Git Flow (develop, release, hotfix branches)
- GitHub Flow (feature branches + main)
- Trunk-based development (short-lived branches, deploy from main)

## Decision

Use trunk-based development:
- `main` is the single source of truth and always deployable.
- Feature work happens on short-lived branches (`feature/sprint-N-description`).
- PRs are squash-merged into main.
- Releases are tagged (`v1.0.0`, `v1.0.1`, etc.) from main.
- CI runs on every PR; CD deploys on tag push.

## Consequences

**Pros:**
- Simple mental model: one branch, one truth.
- Fast iteration: no merge conflicts between long-lived branches.
- CI/CD is straightforward: PR = test, tag = deploy.
- Works perfectly for solo developer workflow.

**Cons:**
- Feature flags needed for incomplete features that land on main (we have database-backed feature flags planned).
- No staging branch — staging environment deploys from main with a separate tag or manual trigger.
