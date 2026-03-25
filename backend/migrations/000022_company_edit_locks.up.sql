-- Company edit locks for moderator editing cooldown.
CREATE TABLE company_edit_locks (
    company_id UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    locked_by  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    locked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- Track moderator company edits for cooldown enforcement.
CREATE TABLE company_edits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    diff        JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_company_edits_company_user ON company_edits (company_id, user_id, created_at DESC);
