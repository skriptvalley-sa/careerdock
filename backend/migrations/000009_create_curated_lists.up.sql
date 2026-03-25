CREATE TABLE IF NOT EXISTS curated_lists (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resume_id         UUID         NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    preferences_hash  VARCHAR(64)  NOT NULL,
    result            JSONB        NOT NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- User's curated list history (most recent first)
CREATE INDEX IF NOT EXISTS idx_curated_lists_user ON curated_lists (user_id, created_at DESC);
