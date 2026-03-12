CREATE TABLE ats_checks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resume_id       UUID         NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    check_type      VARCHAR(10)  NOT NULL
                        CHECK (check_type IN ('company', 'job')),
    company_id      UUID         REFERENCES companies(id) ON DELETE SET NULL,
    job_description TEXT,
    result          JSONB        NOT NULL,
    cache_key       VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Ensure company_id is set for company checks, job_description for job checks
    CONSTRAINT chk_ats_check_target CHECK (
        (check_type = 'company' AND company_id IS NOT NULL) OR
        (check_type = 'job' AND job_description IS NOT NULL)
    )
);

-- User's ATS check history
CREATE INDEX idx_ats_checks_user ON ats_checks (user_id, created_at DESC);

-- Cache key lookup
CREATE INDEX idx_ats_checks_cache_key ON ats_checks (cache_key);

-- Company ATS checks for analytics
CREATE INDEX idx_ats_checks_company ON ats_checks (company_id, created_at DESC)
    WHERE company_id IS NOT NULL;
