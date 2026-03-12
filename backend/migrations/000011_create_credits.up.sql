CREATE TABLE user_credits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credit_type VARCHAR(20) NOT NULL
                    CHECK (credit_type IN (
                        'resume_upload', 'ats_check', 'curated_list', 'cv_generation'
                    )),
    balance     INTEGER     NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_credits UNIQUE (user_id, credit_type)
);

CREATE TABLE credit_transactions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credit_type   VARCHAR(20)  NOT NULL
                      CHECK (credit_type IN (
                          'resume_upload', 'ats_check', 'curated_list', 'cv_generation'
                      )),
    amount        INTEGER      NOT NULL,
    balance_after INTEGER      NOT NULL,
    reason        VARCHAR(100) NOT NULL,
    reference_id  UUID,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- User's transaction history
CREATE INDEX idx_credit_txns_user ON credit_transactions (user_id, created_at DESC);

-- Transactions by type for analytics
CREATE INDEX idx_credit_txns_type ON credit_transactions (credit_type, created_at DESC);
