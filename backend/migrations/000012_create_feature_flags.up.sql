CREATE TABLE IF NOT EXISTS feature_flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(100) NOT NULL,
    enabled     BOOLEAN      NOT NULL DEFAULT FALSE,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_feature_flags_key UNIQUE (key)
);
