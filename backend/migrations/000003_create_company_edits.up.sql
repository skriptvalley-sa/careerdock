CREATE TABLE company_edits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    submitted_by    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewed_by     UUID         REFERENCES users(id) ON DELETE SET NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'approved', 'rejected')),
    changes         JSONB        NOT NULL,
    review_notes    TEXT,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Admin moderation queue: pending edits sorted by submission time
CREATE INDEX idx_company_edits_pending ON company_edits (created_at ASC)
    WHERE status = 'pending';

-- Edits for a specific company
CREATE INDEX idx_company_edits_company ON company_edits (company_id, created_at DESC);

-- Edits by a specific moderator
CREATE INDEX idx_company_edits_submitter ON company_edits (submitted_by, created_at DESC);
