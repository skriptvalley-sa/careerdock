CREATE TABLE resumes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot_number     SMALLINT     NOT NULL CHECK (slot_number BETWEEN 1 AND 3),
    file_name       VARCHAR(255) NOT NULL,
    file_size_bytes INTEGER      NOT NULL CHECK (file_size_bytes > 0 AND file_size_bytes <= 5242880),
    s3_key          VARCHAR(512) NOT NULL,

    -- Extracted and parsed content (Postgres-first strategy)
    extracted_text  TEXT,
    parsed_data     JSONB,
    ats_general     JSONB,

    -- Processing status
    status          VARCHAR(20)  NOT NULL DEFAULT 'uploading'
                        CHECK (status IN (
                            'uploading', 'extracting', 'parsing', 'ready', 'failed'
                        )),
    is_default      BOOLEAN      NOT NULL DEFAULT FALSE,

    -- Archival
    is_archived     BOOLEAN      NOT NULL DEFAULT FALSE,
    archived_at     TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User's resumes
CREATE INDEX idx_resumes_user ON resumes (user_id, created_at DESC);

-- Active slot uniqueness: only one non-archived resume per user per slot
CREATE UNIQUE INDEX idx_resumes_active_slot ON resumes (user_id, slot_number)
    WHERE NOT is_archived;

-- At most one default resume per user (among active resumes)
CREATE UNIQUE INDEX idx_resumes_default ON resumes (user_id)
    WHERE is_default = TRUE AND NOT is_archived;
