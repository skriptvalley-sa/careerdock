CREATE TABLE IF NOT EXISTS payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    razorpay_order_id   VARCHAR(255) NOT NULL,
    razorpay_payment_id VARCHAR(255),
    amount_paise        INTEGER      NOT NULL CHECK (amount_paise > 0),
    currency            VARCHAR(3)   NOT NULL DEFAULT 'INR',
    product_type        VARCHAR(30)  NOT NULL
                            CHECK (product_type IN (
                                'starter_pack', 'resume_upload', 'ats_bundle',
                                'cv_generation', 'rebuy_pack'
                            )),
    status              VARCHAR(20)  NOT NULL DEFAULT 'created'
                            CHECK (status IN ('created', 'captured', 'failed', 'refunded')),
    receipt_number      VARCHAR(100),

    -- Refund tracking
    refund_reason       TEXT,
    refunded_at         TIMESTAMPTZ,
    refunded_by         UUID         REFERENCES users(id) ON DELETE SET NULL,

    webhook_received_at TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_payments_razorpay_order  UNIQUE (razorpay_order_id),
    CONSTRAINT uq_payments_razorpay_payment UNIQUE (razorpay_payment_id),
    CONSTRAINT uq_payments_receipt         UNIQUE (receipt_number)
);

-- User's payment history
CREATE INDEX IF NOT EXISTS idx_payments_user ON payments (user_id, created_at DESC);

-- Admin: filter by status
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status, created_at DESC);
