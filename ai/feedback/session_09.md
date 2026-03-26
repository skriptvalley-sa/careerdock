## Session 09 Plan

Session focused on credit shop UX for premium users, pricing strategy overhaul, support email update, and relevant doc updates.

---

## Scope

### 1. Support email update

Replace all occurrences of the current support email with `skriptvalley+careerdock@gmail.com`.

**Known locations to search:**
- `frontend/src/app/(public)/pricing/page.tsx` — FAQ section
- Any other `.tsx`, `.ts`, `.go`, `.md` files containing the old support email

Action: `grep -r "support@" frontend/ docs/` to find all occurrences before patching.

---

### 2. Credit Shop / Store for premium users

**Current behaviour:**
- All users see "Pricing" in the nav, which shows the Free / Starter Pack marketing page.
- Premium users have no meaningful destination here — the Starter Pack CTA doesn't apply to them.

**New behaviour:**
- **Free users:** nav item remains "Pricing", destination unchanged (public `/pricing` page).
- **Premium users:** nav item renamed to **"Credit Shop"**, destination becomes a new authenticated route `/shop` (or `/credits/shop`).

#### 2.1 Nav changes

File: `frontend/src/components/layout/sidebar.tsx`

- Detect `user.premium_since != null` (already available in `useAuth()`).
- Render `{ href: '/shop', label: 'Credit Shop', icon: ShoppingBag }` for premium users.
- Render `{ href: '/pricing', label: 'Pricing', icon: Tag }` for free users.

#### 2.2 New page: Credit Shop (`/shop`)

Route: `frontend/src/app/(dashboard)/shop/page.tsx`

Layout:
```
┌──────────────────────────────────────────────────────────┐
│  Credit Shop                                              │
│  Your current balance: 7 Resume · 18 ATS · 2 Lists · 5 CV│
├──────────────────────────────────────────────────────────┤
│                                                           │
│  [Starter Refill Pack]  [Resume Bundle]  [ATS Bundle]    │
│  [Curated Lists Bundle] [CV Generation Bundle]           │
│                                                           │
│  Cart (0 items) ────────────────────── [Checkout →]      │
└──────────────────────────────────────────────────────────┘
```

Each pack card shows:
- Pack name + icon
- Credit breakdown (e.g. "50 ATS Checks")
- Price in ₹
- `[+ Add to Cart]` button with quantity selector (max 5 per type)

Cart section (sticky footer or right panel):
- Lists added packs with quantities
- Running total in ₹
- `[Checkout]` → calls `POST /api/payments/orders` for each item (or a new multi-item order endpoint)

**Note on multi-item checkout:** For MVP simplicity, process each cart item as a separate Razorpay order, sequentially on the backend, within the same checkout session. A multi-order batch endpoint can be added in v2.

---

### 3. Updated pricing strategy

#### 3.1 New product catalogue

Replace current product IDs and credit allocations in `backend/internal/payment/` (or wherever `ProductCatalogue` and `CreditAllocation` maps live).

| Product ID | Display Name | Credits | Price (₹) | Notes |
|-----------|-------------|---------|----------:|-------|
| `starter_pack` | Starter Pack | Resume ×10, ATS ×50, Lists ×10, CV Gen ×50 | 799 | Activation pack — sets `premium_since`; free users only |
| `starter_refill` | Starter Refill | Resume ×10, ATS ×50, Lists ×10, CV Gen ×50 | 799 | Same credits as starter; premium users only (replaces `rebuy_pack`) |
| `resume_bundle` | Resume Bundle | Resume ×10 | 199 | |
| `ats_bundle` | ATS Bundle | ATS ×50 | 249 | (was 10 checks for ₹99) |
| `curated_list_bundle` | Curated Lists Bundle | Lists ×5 | 149 | |
| `cv_bundle` | CV Generation Bundle | CV Gen ×50 | 249 | Previously TBD/v2 — now launched |

**Pricing basis:** 30% gross margin on estimated per-unit AI + infra cost. Final prices should be validated against actual Claude API call costs before launch. Formula: `price = cost / 0.70`, rounded to nearest ₹9 or ₹49.

**GST:** All prices remain GST-inclusive (no change to billing approach).

#### 3.2 Credit allocation changes

| Product | `resume_upload` | `ats_check` | `curated_list` | `cv_generation` | Sets `premium_since`? |
|---------|----------------:|------------:|---------------:|----------------:|:---------------------:|
| `starter_pack` | +10 | +50 | +10 | +50 | Yes |
| `starter_refill` | +10 | +50 | +10 | +50 | No |
| `resume_bundle` | +10 | 0 | 0 | 0 | No |
| `ats_bundle` | 0 | +50 | 0 | 0 | No |
| `curated_list_bundle` | 0 | 0 | +5 | 0 | No |
| `cv_bundle` | 0 | 0 | 0 | +50 | No |

**Removed products:** `resume_upload` (single upload for ₹49) — replaced by `resume_bundle`.
**Renamed:** `rebuy_pack` → `starter_refill` for clarity.

---

### 4. Backend changes

#### 4.1 Product catalogue update

File: wherever `ProductCatalogue` map is defined (likely `backend/internal/payment/catalogue.go` or similar).

- Update all product IDs, display names, prices, and credit amounts per table above.
- Remove `resume_upload` single-upload product.
- Add `starter_refill`, `resume_bundle`, `curated_list_bundle`, `cv_bundle`.
- `cv_generation` credit type is now purchasable (remove the "TBD/v2" gating).

#### 4.2 Business rule update

- `starter_pack` → only for users where `premium_since IS NULL`.
- `starter_refill` → only for users where `premium_since IS NOT NULL` (replaces `rebuy_pack` check).
- All other bundles → available to any premium user (no free-tier purchase of bundles).

#### 4.3 Payment confirm endpoint (already exists: `POST /api/payments/confirm`)

Review existing implementation for any hardcoded product IDs or amounts that reference the old catalogue. Update as needed.

---

### 5. Frontend changes

#### 5.1 Public pricing page update

File: `frontend/src/app/(public)/pricing/page.tsx`

- Update Starter Pack card: show new credit amounts (Resume ×10, ATS ×50, Lists ×10, CV Gen ×50) and new price ₹799.
- Update support email in FAQ section to `skriptvalley+careerdock@gmail.com`.
- Free tier card: no change needed.
- Remove any reference to old ₹99 ATS Bundle or ₹49 resume upload from the public page.

#### 5.2 Insufficient credits modal / upsell

File(s): wherever `ErrInsufficientCredits` is surfaced in UI.

- Update "Buy ATS Bundle (₹99 for 10 checks)" → "Buy ATS Bundle (₹249 for 50 checks)".
- Add CV generation bundle upsell if cv_generation credits run out.

#### 5.3 New Credit Shop page

New file: `frontend/src/app/(dashboard)/shop/page.tsx`

Components needed:
- `PackCard` — displays pack name, credits, price, add-to-cart button
- `ShopCart` — cart state (local React state, not persisted), checkout button
- Reuse existing `useAuth()` for credit balance display
- Reuse existing payment flow (`POST /api/payments/orders` + Razorpay checkout widget)

Cart state shape:
```typescript
type CartItem = {
  productId: string;
  label: string;
  price: number;
  qty: number;
};
```

---

### 6. Docs to update

| File | What to update |
|------|----------------|
| `docs/LLD/payments.md` | Section 2 product catalogue (new IDs, prices, credit amounts); Section 2.1 credit allocation table; Section 9.1 purchase flow description; Section 10.2 admin transaction log (new product names) |
| `docs/PRD.md` | Paid tier description — update credit amounts and add CV generation as active (not v2) |

---

## Implementation order

1. Support email grep + replace (quick win)
2. Backend: update product catalogue + business rules
3. Backend: run `make lint-backend` + `make build`
4. Frontend: update public pricing page (amounts + email)
5. Frontend: sidebar nav conditional (Pricing vs Credit Shop)
6. Frontend: Credit Shop page + cart
7. Frontend: run `make lint-frontend` + `make build`
8. Docs: update `payments.md` and `PRD.md`
9. Branch: `feat/credit-shop-session-09` → PR

---

## Branch

```bash
git checkout -b feat/credit-shop-session-09
```
