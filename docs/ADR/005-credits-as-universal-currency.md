# ADR-005: Credits as Universal Currency

**Status:** Proposed
**Date:** 2026-03-14
**Context:** Session 02 feedback (S2-PAY-01)

## Decision

Credits will serve as the universal currency across the CareerDock platform. Instead of granting specific feature access (e.g., "3 resume uploads"), credits are a fungible unit that maps to individual actions.

## Rationale

- **Flexibility:** Users choose how to spend credits (resume analysis, ATS scoring, company matching, etc.)
- **Simplicity:** One balance to track instead of per-feature quotas
- **Monetization:** Easy to offer credit packs at different price points, promotional credits, referral bonuses
- **Cost mapping:** Each AI action has a known LLM cost; credit pricing can reflect actual cost per action

## Credit Action Mapping (Draft)

| Action | Credits | Notes |
|--------|---------|-------|
| Resume upload + AI analysis | 3 | Includes extraction, skill mapping, suggestions |
| General ATS score | 1 | One resume against general criteria |
| Company-specific ATS score | 2 | Tailored to company's known preferences |
| Job-specific ATS score | 2 | Tailored to a specific JD |
| AI company matching | 1 | Match resume profile to company directory |

## Starter Pack

The one-time Starter Pack (approx. ₹299-499) includes an initial credit balance (e.g., 10-15 credits). Additional credit packs available a la carte.

## Implementation Notes

- Sprint 3 will design the `user_credits` table, transaction log, and deduction flow
- Credits are non-expiring for MVP
- Admin panel will support granting/revoking credits per user
- Feature flag `payments_enabled` gates credit purchase UI

## Consequences

- Need to define credit costs carefully to balance user value vs LLM costs
- Must implement idempotent credit deduction (no double-spend on retries)
- Credit balance should be prominently displayed in the UI
