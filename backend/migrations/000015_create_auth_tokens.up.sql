CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ  NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_email_verification_token UNIQUE (token)
);

-- Cleanup expired tokens
CREATE INDEX IF NOT EXISTS idx_email_tokens_expires ON email_verification_tokens (expires_at)
    WHERE used_at IS NULL;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ  NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_password_reset_token UNIQUE (token)
);

-- Cleanup expired tokens
CREATE INDEX IF NOT EXISTS idx_password_tokens_expires ON password_reset_tokens (expires_at)
    WHERE used_at IS NULL;
