CREATE TABLE application_status_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_entry_id   UUID        NOT NULL REFERENCES list_entries(id) ON DELETE CASCADE,
    from_status     VARCHAR(20),
    to_status       VARCHAR(20) NOT NULL,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- History for a specific entry, chronological
CREATE INDEX idx_status_history_entry ON application_status_history (list_entry_id, changed_at ASC);

CREATE TABLE interview_rounds (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_entry_id   UUID         NOT NULL REFERENCES list_entries(id) ON DELETE CASCADE,
    round_number    SMALLINT     NOT NULL CHECK (round_number > 0),
    round_type      VARCHAR(100) NOT NULL,
    scheduled_date  DATE,
    outcome         VARCHAR(20)  NOT NULL DEFAULT 'pending'
                        CHECK (outcome IN ('passed', 'failed', 'pending')),
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Rounds for a specific entry, ordered by round number
CREATE INDEX idx_interview_rounds_entry ON interview_rounds (list_entry_id, round_number ASC);
