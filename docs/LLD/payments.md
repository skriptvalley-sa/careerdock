# CareerDock — Payment & Credit System Design (LLD)

> **Version:** 1.0
> **Status:** Draft (Phase 3)
> **Last updated:** 2026-03-11
> **Depends on:** [PRD.md](../PRD.md), [ARCHITECTURE.md](../ARCHITECTURE.md), [database.md](./database.md), [api.md](./api.md)

---

## 1. Overview

CareerDock uses a **one-time purchase + à la carte** model. No subscriptions. All payments go through **Razorpay** (India-first, native ₹ + UPI support). Purchases allocate **credits** to the user's account, which are consumed when performing premium actions.

**Key principles:**
1. **Idempotent** — duplicate webhook deliveries never double-allocate credits.
2. **Atomic** — credit allocation and payment status update happen in a single DB transaction.
3. **Auditable** — every credit change has a `credit_transactions` entry.
4. **Recoverable** — failed webhooks can be manually reconciled via admin dashboard.

---

## 2. Product Catalogue

All prices in INR. Stored in application config (not DB) — price changes require a code update and deployment. This is intentional for MVP: keeps pricing logic simple and auditable via git history.

| Product ID | Display Name | Price (₹) | Amount (paise) |
|-----------|-------------|----------:|---------------:|
| `starter_pack` | Starter Pack | 449 | 44900 |
| `starter_refill` | Starter Refill Pack | 399 | 39900 |
| `resume_bundle` | Resume Bundle | 89 | 8900 |
| `ats_bundle` | ATS Bundle | 229 | 22900 |
| `curated_list_bundle` | Curated Lists Bundle | 59 | 5900 |
| `cv_bundle` | Cover Letter Bundle | Coming soon | N/A |

### 2.1 Credit Allocation per Product

| Product | `resume_upload` | `ats_check` | `curated_list` | `cv_generation` | Sets `premium_since`? |
|---------|----------------:|------------:|---------------:|----------------:|:---------------------:|
| `starter_pack` | +10 | +50 | +10 | +50 | Yes (if not already set) |
| `starter_refill` | +10 | +50 | +10 | +50 | No |
| `resume_bundle` | +10 | 0 | 0 | 0 | No |
| `ats_bundle` | 0 | +50 | 0 | 0 | No |
| `curated_list_bundle` | 0 | 0 | +5 | 0 | No |
| `cv_bundle` | 0 | 0 | 0 | +50 | No |

**Business rules:** `starter_pack` is only available to non-premium users. All other bundles require `premium_since` to already be set. `cv_bundle` is intentionally unavailable until the cover letter feature ships.

---

## 3. Payment Flow

### 3.1 Sequence Diagram

```
    Frontend                    Backend                         Razorpay
       │                          │                                │
   ┌───┴───┐                      │                                │
   │ User  │                      │                                │
   │ clicks│                      │                                │
   │ "Buy" │                      │                                │
   └───┬───┘                      │                                │
       │                          │                                │
       │  POST /api/payments/     │                                │
       │       orders             │                                │
       │  { product_type:         │                                │
       │    "starter_pack" }      │                                │
       │─────────────────────────►│                                │
       │                          │                                │
       │                          │  1. Validate product_type      │
       │                          │  2. Look up price              │
       │                          │  3. Generate receipt number     │
       │                          │  4. Create Razorpay Order      │
       │                          │─────────────────────────────►  │
       │                          │◄─────────────────────────────  │
       │                          │  { id: "order_xxx",            │
       │                          │    amount: 44900,              │
       │                          │    status: "created" }         │
       │                          │                                │
       │                          │  5. Insert into payments       │
       │                          │     table (status: "created")  │
       │                          │                                │
       │  { razorpay_order_id,    │                                │
       │    amount_paise,         │                                │
       │    razorpay_key_id }     │                                │
       │◄─────────────────────────│                                │
       │                          │                                │
       │  6. Open Razorpay        │                                │
       │     Checkout widget      │                                │
       │─────────────────────────────────────────────────────────► │
       │                          │                                │
       │  (User completes UPI /   │                                │
       │   card / net banking)    │                                │
       │                          │                                │
       │◄─────────────────────────────────────────────────────────┤
       │  Checkout success        │                                │
       │  callback (client-side)  │                                │
       │                          │                                │
       │  7. Frontend shows       │  8. Razorpay sends webhook     │
       │     "Processing..."      │     payment.captured            │
       │     message              │◄─────────────────────────────  │
       │                          │                                │
       │                          │  9. Verify webhook signature   │
       │                          │  10. Find payment by order_id  │
       │                          │  11. Idempotency check         │
       │                          │  12. BEGIN TRANSACTION         │
       │                          │      - Update payment status   │
       │                          │      - Allocate credits        │
       │                          │      - Insert credit_txns      │
       │                          │      - Set premium_since       │
       │                          │  13. COMMIT                    │
       │                          │  14. Publish SSE event         │
       │                          │  15. Queue receipt email        │
       │                          │                                │
       │                          │  200 OK                        │
       │                          │─────────────────────────────►  │
       │                          │                                │
       │  SSE: credits_updated    │                                │
       │◄─────────────────────────│                                │
       │                          │                                │
       │  16. Frontend refetches  │                                │
       │      /api/credits        │                                │
       │      Shows confirmation  │                                │
```

### 3.2 Step-by-Step Details

#### Step 1-5: Order Creation (Backend)

```go
func (s *PaymentService) CreateOrder(ctx context.Context, userID uuid.UUID, productType string) (*PaymentOrder, error) {
    // 1. Validate product type
    product, ok := ProductCatalogue[productType]
    if !ok {
        return nil, ErrInvalidProduct
    }

    // 2. Business rule: starter_pack only for non-premium users
    if productType == "starter_pack" {
        user, _ := s.userRepo.GetByID(ctx, userID)
        if user.PremiumSince != nil {
            return nil, ErrAlreadyPremium
        }
    }

    // 3. Business rule: all refill / bundle products require premium
    if productType != "starter_pack" {
        user, _ := s.userRepo.GetByID(ctx, userID)
        if user.PremiumSince == nil {
            return nil, ErrNotPremium
        }
    }

    // 4. Generate receipt number: CDOCK-YYYYMMDD-XXXX
    receiptNumber := generateReceiptNumber()

    // 5. Create Razorpay order
    rzpOrder, err := s.razorpay.CreateOrder(razorpay.OrderParams{
        Amount:   product.AmountPaise,
        Currency: "INR",
        Receipt:  receiptNumber,
        Notes: map[string]string{
            "user_id":      userID.String(),
            "product_type": productType,
        },
    })

    // 6. Insert payment record
    payment := &Payment{
        ID:               uuid.New(),
        UserID:           userID,
        RazorpayOrderID:  rzpOrder.ID,
        AmountPaise:      product.AmountPaise,
        Currency:         "INR",
        ProductType:      productType,
        Status:           "created",
        ReceiptNumber:    receiptNumber,
    }
    s.paymentRepo.Create(ctx, payment)

    return &PaymentOrder{
        PaymentID:       payment.ID,
        RazorpayOrderID: rzpOrder.ID,
        AmountPaise:     product.AmountPaise,
        Currency:        "INR",
        RazorpayKeyID:   s.config.RazorpayKeyID,
    }, nil
}
```

#### Step 6: Razorpay Checkout (Frontend)

```typescript
const openCheckout = (order: PaymentOrder) => {
  const options = {
    key: order.razorpay_key_id,
    amount: order.amount_paise,
    currency: 'INR',
    name: 'CareerDock',
    description: 'Starter Pack',
    order_id: order.razorpay_order_id,
    handler: (response: RazorpayResponse) => {
      // Payment successful on client side
      // Don't allocate credits here — wait for webhook
      showProcessingMessage();
    },
    prefill: {
      email: user.email,
      name: user.name,
    },
    theme: {
      color: '#4F46E5', // brand colour
    },
  };

  const rzp = new Razorpay(options);
  rzp.open();
};
```

**Important:** The `handler` callback only confirms the payment was attempted. Credits are allocated server-side via webhook — never trust the client callback for credit allocation.

#### Steps 9-15: Webhook Processing (Backend)

```go
func (s *PaymentService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
    // 9. Verify signature
    if !razorpay.VerifyWebhookSignature(payload, signature, s.config.RazorpayWebhookSecret) {
        return ErrInvalidSignature
    }

    // Parse event
    event := parseWebhookEvent(payload)
    if event.Event != "payment.captured" {
        return nil // Only process captured payments
    }

    orderID := event.Payload.Payment.Entity.OrderID
    paymentID := event.Payload.Payment.Entity.ID

    // 10. Find payment by order ID
    payment, err := s.paymentRepo.GetByRazorpayOrderID(ctx, orderID)
    if err != nil {
        return ErrPaymentNotFound
    }

    // 11. Idempotency check
    if payment.Status == "captured" {
        return nil // Already processed — return success
    }

    // 12-13. Atomic transaction: update payment + allocate credits
    err = s.db.WithTransaction(ctx, func(tx pgx.Tx) error {
        // Update payment status
        payment.Status = "captured"
        payment.RazorpayPaymentID = &paymentID
        payment.WebhookReceivedAt = timeNow()
        if err := s.paymentRepo.UpdateTx(ctx, tx, payment); err != nil {
            return err
        }

        // Allocate credits
        allocation := CreditAllocation[payment.ProductType]
        for creditType, amount := range allocation {
            if amount == 0 {
                continue
            }
            // Upsert credit balance
            newBalance, err := s.creditRepo.AddBalanceTx(ctx, tx, payment.UserID, creditType, amount)
            if err != nil {
                return err
            }
            // Insert audit transaction
            s.creditTxnRepo.CreateTx(ctx, tx, &CreditTransaction{
                UserID:       payment.UserID,
                CreditType:   creditType,
                Amount:       amount,
                BalanceAfter: newBalance,
                Reason:       payment.ProductType + "_purchase",
                ReferenceID:  &payment.ID,
            })
        }

        // Set premium_since for starter_pack
        if payment.ProductType == "starter_pack" {
            s.userRepo.SetPremiumTx(ctx, tx, payment.UserID, timeNow())
        }

        return nil
    })

    // 14. Publish SSE notification (outside transaction — best effort)
    s.notifier.Send(payment.UserID, Notification{
        Type:  "credits_updated",
        Title: "Payment confirmed",
        Data:  map[string]any{"product_type": payment.ProductType},
    })

    // 15. Queue receipt email (async)
    s.queue.Enqueue(TaskSendEmail, EmailPayload{
        To:       payment.UserID,
        Template: "payment_receipt",
        Data:     payment,
    })

    return err
}
```

---

## 4. Credit System

### 4.1 Credit Types

| Type | Consumed By | Allocated By |
|------|-------------|-------------|
| `resume_upload` | Resume upload / re-upload | starter_pack, starter_refill, resume_bundle |
| `ats_check` | Company ATS check, Job ATS check | starter_pack, starter_refill, ats_bundle |
| `curated_list` | AI-curated list generation | starter_pack, starter_refill, curated_list_bundle |
| `cv_generation` | Tailored cover letter generation | starter_pack, starter_refill |

**ATS check credits are fungible** — the same credit is consumed whether the check is company-specific or job-specific.

### 4.2 Credit Consumption Flow

```go
func (s *ATSService) RequestCompanyCheck(ctx context.Context, userID, companyID uuid.UUID) (*ATSCheck, error) {
    // 1. Check cache first
    cacheKey := computeCacheKey(userID, companyID)
    if cached, ok := s.cache.Get(ctx, cacheKey); ok {
        return cached, nil
    }

    // 2. Verify credit balance
    balance, err := s.creditRepo.GetBalance(ctx, userID, "ats_check")
    if err != nil || balance < 1 {
        return nil, ErrInsufficientCredits
    }

    // 3. Deduct credit atomically
    err = s.db.WithTransaction(ctx, func(tx pgx.Tx) error {
        // Deduct 1 credit (uses SELECT FOR UPDATE to prevent race conditions)
        newBalance, err := s.creditRepo.DeductBalanceTx(ctx, tx, userID, "ats_check", 1)
        if err != nil {
            return err // CHECK constraint prevents negative balance
        }

        // Record transaction
        s.creditTxnRepo.CreateTx(ctx, tx, &CreditTransaction{
            UserID:       userID,
            CreditType:   "ats_check",
            Amount:       -1,
            BalanceAfter: newBalance,
            Reason:       "ats_check_consumed",
            ReferenceID:  &checkID,
        })

        return nil
    })

    // 4. Queue async job
    s.queue.Enqueue(TaskATSCompanyCheck, ATSPayload{
        CheckID:   checkID,
        UserID:    userID,
        CompanyID: companyID,
    })

    return &ATSCheck{ID: checkID, Status: "processing"}, nil
}
```

### 4.3 Race Condition Prevention

**Problem:** Two concurrent requests could both read `balance = 1` and both deduct, resulting in `balance = -1`.

**Solution:** `SELECT ... FOR UPDATE` + `CHECK (balance >= 0)`:

```sql
-- In DeductBalanceTx:
UPDATE user_credits
SET balance = balance - $1, updated_at = NOW()
WHERE user_id = $2 AND credit_type = $3
RETURNING balance;
-- If balance would go negative, CHECK constraint causes the UPDATE to fail
```

The `CHECK (balance >= 0)` constraint on `user_credits.balance` acts as a database-level safeguard. Combined with `SELECT FOR UPDATE` (row-level lock), this prevents negative balances even under concurrent requests.

### 4.4 Credit Refund (Admin)

When an admin issues a refund:

```go
func (s *PaymentService) RefundPayment(ctx context.Context, adminID, paymentID uuid.UUID, reason string) error {
    payment, _ := s.paymentRepo.GetByID(ctx, paymentID)

    // Validation
    if payment.Status != "captured" {
        return ErrPaymentNotCaptured
    }
    if time.Since(payment.CreatedAt) > 7*24*time.Hour {
        return ErrRefundWindowExpired
    }

    // Check if any credits from this payment have been consumed
    allocation := CreditAllocation[payment.ProductType]
    for creditType, amount := range allocation {
        consumed, _ := s.creditTxnRepo.CountConsumedSince(ctx, payment.UserID, creditType, payment.CreatedAt)
        if consumed > 0 {
            return ErrCreditsAlreadyConsumed
        }
    }

    err := s.db.WithTransaction(ctx, func(tx pgx.Tx) error {
        // Update payment
        payment.Status = "refunded"
        payment.RefundReason = &reason
        payment.RefundedAt = timeNowPtr()
        payment.RefundedBy = &adminID
        s.paymentRepo.UpdateTx(ctx, tx, payment)

        // Deduct allocated credits
        for creditType, amount := range allocation {
            newBalance, _ := s.creditRepo.DeductBalanceTx(ctx, tx, payment.UserID, creditType, amount)
            s.creditTxnRepo.CreateTx(ctx, tx, &CreditTransaction{
                UserID:       payment.UserID,
                CreditType:   creditType,
                Amount:       -amount,
                BalanceAfter: newBalance,
                Reason:       "admin_refund",
                ReferenceID:  &payment.ID,
            })
        }

        // If starter_pack refund and no other payments, clear premium
        if payment.ProductType == "starter_pack" {
            otherPayments, _ := s.paymentRepo.CountCapturedExcluding(ctx, tx, payment.UserID, payment.ID)
            if otherPayments == 0 {
                s.userRepo.ClearPremiumTx(ctx, tx, payment.UserID)
            }
        }

        // Audit log
        s.auditRepo.CreateTx(ctx, tx, &AuditLog{
            AdminID:    adminID,
            Action:     "refund_issued",
            EntityType: "payment",
            EntityID:   &payment.ID,
            Details:    map[string]any{"reason": reason, "amount_paise": payment.AmountPaise},
        })

        return nil
    })

    // Trigger Razorpay refund API (async, outside transaction)
    s.queue.Enqueue(TaskRazorpayRefund, RefundPayload{
        PaymentID:        payment.RazorpayPaymentID,
        AmountPaise:      payment.AmountPaise,
        InternalPaymentID: payment.ID,
    })

    return err
}
```

---

## 5. Razorpay Integration Details

### 5.1 Configuration

```go
type RazorpayConfig struct {
    KeyID         string `mapstructure:"RAZORPAY_KEY_ID"`         // rzp_live_xxx or rzp_test_xxx
    KeySecret     string `mapstructure:"RAZORPAY_KEY_SECRET"`
    WebhookSecret string `mapstructure:"RAZORPAY_WEBHOOK_SECRET"` // For webhook signature verification
}
```

- **Test mode:** Use `rzp_test_*` keys for local dev and staging. Razorpay test mode provides dummy payment methods.
- **Live mode:** Use `rzp_live_*` keys for production only.

### 5.2 Webhook Configuration (Razorpay Dashboard)

| Setting | Value |
|---------|-------|
| Webhook URL | `https://api.careerdock.skriptvalley.com/api/payments/webhook` |
| Events | `payment.captured`, `payment.failed`, `refund.created` |
| Secret | Stored in AWS Secrets Manager, loaded as `RAZORPAY_WEBHOOK_SECRET` |
| Active | Yes |

### 5.3 Webhook Signature Verification

Razorpay signs webhooks with HMAC-SHA256:

```go
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

The signature is sent in the `X-Razorpay-Signature` HTTP header.

### 5.4 Webhook Event Types

| Event | Action |
|-------|--------|
| `payment.captured` | Allocate credits, update payment status. **Primary flow.** |
| `payment.failed` | Update payment status to `failed`. No credit allocation. |
| `refund.created` | Confirmation that Razorpay processed the refund. Update refund metadata if needed. |

**We only act on `payment.captured`** for credit allocation. Other events are logged but don't trigger credit changes.

### 5.5 Razorpay Go SDK Usage

```go
import razorpay "github.com/razorpay/razorpay-go"

client := razorpay.NewClient(config.KeyID, config.KeySecret)

// Create order
orderParams := map[string]interface{}{
    "amount":   44900,
    "currency": "INR",
    "receipt":  "CDOCK-20260311-0001",
    "notes": map[string]interface{}{
        "user_id":      userID,
        "product_type": "starter_pack",
    },
}
order, err := client.Order.Create(orderParams, nil)

// Initiate refund
refundParams := map[string]interface{}{
    "amount": 44900,
    "notes": map[string]interface{}{
        "reason": "Customer requested, no credits consumed",
    },
}
refund, err := client.Payment.Refund(razorpayPaymentID, refundParams, nil)
```

---

## 6. Receipt Number Generation

Format: `CDOCK-YYYYMMDD-XXXX`

- `CDOCK` — static prefix.
- `YYYYMMDD` — date of order creation.
- `XXXX` — zero-padded sequential counter per day.

```go
// Simple approach: use a DB sequence or atomic Redis counter
func generateReceiptNumber(ctx context.Context, redis *redis.Client) string {
    date := time.Now().Format("20060102")
    key := fmt.Sprintf("receipt_counter:%s", date)
    counter := redis.Incr(ctx, key).Val()
    redis.Expire(ctx, key, 48*time.Hour) // TTL 2 days for safety
    return fmt.Sprintf("CDOCK-%s-%04d", date, counter)
}
```

The `UNIQUE` constraint on `payments.receipt_number` ensures no duplicates even if Redis is reset.

---

## 7. Edge Cases

### 7.1 Duplicate Webhook Delivery

**Scenario:** Razorpay sends the same `payment.captured` webhook twice.

**Handling:** Idempotency check — if `payment.status` is already `captured`, return 200 immediately without re-allocating credits.

### 7.2 Webhook Arrives Before Frontend Callback

**Scenario:** Webhook fires before the Razorpay Checkout widget returns to the frontend.

**Handling:** No issue — credits are allocated server-side via webhook regardless of frontend state. When the frontend callback fires, it shows "Processing..." and the SSE event or next `/api/credits` poll will show updated credits.

### 7.3 Webhook Never Arrives

**Scenario:** Network issue prevents webhook delivery.

**Handling:**
1. Razorpay retries webhooks for up to 24 hours.
2. If still not received: admin can manually reconcile via admin dashboard.
3. Admin dashboard shows payments with status `created` older than 30 minutes — these are candidates for manual verification against Razorpay dashboard.

### 7.4 Payment Fails

**Scenario:** User's payment method is declined.

**Handling:** No credits allocated. Payment record stays in `created` status. Razorpay sends `payment.failed` webhook — we update status to `failed`. User can retry by creating a new order.

### 7.5 User Closes Browser During Payment

**Scenario:** User starts payment flow, closes browser, and payment completes in the background (e.g., UPI payment already initiated).

**Handling:** Webhook still arrives — credits are allocated. When the user logs in next, they'll see their credits. No action needed.

### 7.6 Concurrent Purchases

**Scenario:** User opens two tabs and initiates two purchases simultaneously.

**Handling:** Each creates a separate Razorpay order. Both webhooks are processed independently. Credits from both are allocated. This is intentional — user can buy multiple products.

### 7.7 Refund After Partial Credit Consumption

**Scenario:** User buys Starter Pack (50 ATS credits), uses 5 ATS checks, then requests refund.

**Handling:** Refund denied — the `RefundPayment` flow checks if any credits from this payment period have been consumed. Any consumption blocks the refund. The 7-day window also applies.

### 7.8 Price Change Mid-Checkout

**Scenario:** Admin deploys a price change while a user has an open Razorpay Checkout.

**Handling:** The Razorpay order was created with the old price and is immutable. The user pays the old price. Credit allocation is based on `payment.product_type`, not the current catalogue price. No issue.

### 7.9 Razorpay API Down During Order Creation

**Scenario:** Razorpay API returns an error when trying to create an order.

**Handling:** Return `503 SERVICE_UNAVAILABLE` to the frontend. No payment record created. User can retry. No credits at risk.

---

## 8. GST Considerations

For MVP, prices are **GST-inclusive**. The displayed price is the final price the user pays.

| Product | Display Price | GST (18%) | Base Price |
|---------|-------------:|----------:|-----------:|
| Starter Pack | ₹449 | ₹68.49 | ₹380.51 |
| Starter Refill Pack | ₹399 | ₹60.86 | ₹338.14 |
| Resume Bundle | ₹89 | ₹13.58 | ₹75.42 |
| ATS Bundle | ₹229 | ₹34.93 | ₹194.07 |
| Curated Lists Bundle | ₹59 | ₹9.00 | ₹50.00 |

**Invoice/receipt** should show the GST breakdown. This is handled in the receipt email template.

**Note:** GST registration is required once revenue exceeds ₹20 lakh/year. For MVP launch, this is unlikely to be immediate. Monitor via admin dashboard revenue tracking.

---

## 9. Frontend Payment UX

### 9.1 Purchase Flow (User Perspective)

```
1. User visits /pricing page
   → Free users see Starter Pack (₹449)
   → Premium users are routed to /shop for refills

2. User clicks "Buy Now"
   → Frontend calls POST /api/payments/orders
   → Receives razorpay_order_id

3. Razorpay Checkout widget opens
   → User selects UPI / Card / Net Banking
   → Completes payment

4. On success callback:
   → Frontend shows "Payment received. Setting up your account..."
   → Opens SSE connection (if not already open)
   → Polls /api/credits every 3 seconds as fallback

5. On SSE event "credits_updated" OR poll shows credits:
   → Frontend shows success message with credit summary
   → Redirects to premium dashboard

6. On failure callback:
   → Frontend shows "Payment failed. Please try again."
   → No retry limit (user can click "Buy Now" again)
```

**Credit Shop:** Premium users use `/shop` to add refill packs and bundles to a local cart. Checkout now creates a single bundled Razorpay order that matches the cart total, while the backend stores a cart snapshot to allocate credits correctly on capture.

### 9.2 Credit Display

**Premium dashboard header:**

```
┌─────────────────────────────────────────────────┐
│  Your Credits                                    │
│                                                  │
│  Resume Uploads: 7 remaining                     │
│  ATS Checks: 18 remaining                        │
│  Curated Lists: 2 remaining                      │
│  Cover Letters: 5 remaining                      │
│                                                  │
│  [Open Credit Shop]                              │
└─────────────────────────────────────────────────┘
```

### 9.3 Insufficient Credits UX

When a user attempts an action with insufficient credits:

```
┌─────────────────────────────────────────────────┐
│  ⚠ Insufficient Credits                         │
│                                                  │
│  You need 1 ATS check credit to run this check.  │
│  Current balance: 0                              │
│                                                  │
│  [Buy ATS Bundle (₹229 for 50 checks)]          │
└─────────────────────────────────────────────────┘
```

---

## 10. Admin Payment Dashboard

### 10.1 Revenue Overview

```
┌─────────────────────────────────────────────────┐
│  Revenue                           March 2026    │
│                                                  │
│  Total Revenue:     ₹58,000                      │
│  This Month:        ₹8,500                       │
│  Today:             ₹449                         │
│                                                  │
│  By Product:                                     │
│  ├── Starter Pack:      ₹22,450 (50 purchases)  │
│  ├── Starter Refill:    ₹5,985  (15 purchases)  │
│  ├── ATS Bundle:        ₹4,580  (20 purchases)  │
│  └── Resume Bundle:     ₹712    (8 purchases)   │
│                                                  │
│  Refunds:             ₹1,598 (2 refunds)        │
│  Net Revenue:         ₹56,402                    │
└─────────────────────────────────────────────────┘
```

### 10.2 Transaction Log

Table view with columns:
- Date | User | Product | Amount | Status | Receipt # | Actions

Filters: status, product_type, date range, user search.

Typical product values now include: `starter_pack`, `starter_refill`, `resume_bundle`, `ats_bundle`, `curated_list_bundle`, `cv_bundle`.

Action buttons: "View Details", "Issue Refund" (for eligible payments).

### 10.3 Reconciliation View

Shows payments in `created` status older than 30 minutes — these may have been paid but the webhook was missed. Admin can:
1. Check Razorpay dashboard for the order status.
2. If captured: manually trigger credit allocation.
3. If failed/expired: mark as failed.

---

## 11. Monitoring & Alerts

### 11.1 Payment Metrics

| Metric | Tracked Via | Alert Threshold |
|--------|-------------|-----------------|
| Webhook processing time | slog + Prometheus | p99 > 5 seconds |
| Webhook signature failures | Counter | > 3 in 1 hour |
| Payment conversion rate | `created` → `captured` | < 50% over 24h |
| Stale orders (`created` > 30 min) | Periodic check | > 5 stale orders |
| Credit allocation failures | Error counter | Any occurrence |
| Razorpay API errors | Error counter | > 5 in 15 min |

### 11.2 Logging

All payment operations log with structured fields:

```json
{
  "level": "INFO",
  "msg": "payment_webhook_processed",
  "user_id": "01912345-...",
  "payment_id": "01912400-...",
  "razorpay_order_id": "order_abc123",
  "product_type": "starter_pack",
  "amount_paise": 44900,
  "credits_allocated": {
    "resume_upload": 10,
    "ats_check": 50,
    "curated_list": 10,
    "cv_generation": 50
  },
  "processing_time_ms": 45,
  "request_id": "req_xyz789"
}
```

---

## 12. Testing Strategy

### 12.1 Unit Tests

- `TestCreateOrder` — validates product lookup, receipt generation, Razorpay mock.
- `TestHandleWebhook` — signature verification, idempotency, credit allocation.
- `TestHandleWebhookDuplicate` — second call is no-op.
- `TestHandleWebhookInvalidSignature` — returns error.
- `TestCreditDeduction` — concurrent deductions don't go negative.
- `TestRefundEligibility` — within 7 days, no consumed credits.
- `TestRefundWithConsumedCredits` — should fail.

### 12.2 Integration Tests

- Full payment flow with Razorpay test mode.
- Webhook delivery simulation.
- Credit balance verification end-to-end.

### 12.3 Local Development

- Use Razorpay test keys (`rzp_test_*`).
- Test webhooks via Razorpay's webhook testing tool or a local tunnel (ngrok).
- Seed a test user with credits for immediate premium feature testing.

---

## 13. Cross-Reference

| Architecture Decision | Payment Implementation |
|----------------------|----------------------|
| Razorpay integration (§3.8) | Full flow with order creation, checkout widget, webhook processing |
| Idempotent processing (§3.8) | UNIQUE on `razorpay_order_id` + status check before allocation |
| Credit system (PRD §3.4.1) | `user_credits` table with CHECK constraint + `credit_transactions` audit trail |
| SSE notifications (§3.10) | `credits_updated` event published after successful webhook processing |
| Async email (§3.9) | Receipt email queued via Asynq after webhook processing |
| Admin payment dashboard (Phase 7) | Revenue overview, transaction log, reconciliation view |
