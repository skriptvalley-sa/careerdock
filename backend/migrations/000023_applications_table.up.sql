-- Create applications table: tracks job applications at the company level, not per list.
-- A user can have multiple applications per company (different roles).
CREATE TABLE applications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id   UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    role_title   VARCHAR(255),
    status       VARCHAR(20)  NOT NULL DEFAULT 'not_applied'
                     CHECK (status IN (
                         'not_applied', 'applied', 'phone_screen', 'interview',
                         'offer', 'rejected', 'accepted', 'withdrawn'
                     )),
    date_applied DATE,
    notes        TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_user ON applications (user_id);
CREATE INDEX idx_applications_company ON applications (company_id);
CREATE INDEX idx_applications_user_company ON applications (user_id, company_id);

-- Migrate existing application data from list_entries.
-- For each unique (user_id, company_id), take the entry with the most recent update.
INSERT INTO applications (id, user_id, company_id, role_title, status, date_applied, notes, created_at, updated_at)
SELECT DISTINCT ON (ul.user_id, le.company_id)
    gen_random_uuid(),
    ul.user_id,
    le.company_id,
    le.role_title,
    le.status,
    le.date_applied,
    le.notes,
    le.created_at,
    le.updated_at
FROM list_entries le
JOIN user_lists ul ON le.list_id = ul.id
WHERE le.status != 'not_applied'
ORDER BY ul.user_id, le.company_id, le.updated_at DESC;

-- Re-point status history: add application_id column, migrate references, then drop old FK.
ALTER TABLE application_status_history ADD COLUMN application_id UUID REFERENCES applications(id) ON DELETE CASCADE;

UPDATE application_status_history ash
SET application_id = a.id
FROM list_entries le
JOIN user_lists ul ON le.list_id = ul.id
JOIN applications a ON a.user_id = ul.user_id AND a.company_id = le.company_id
WHERE ash.list_entry_id = le.id;

-- Delete orphaned history rows that couldn't be migrated
DELETE FROM application_status_history WHERE application_id IS NULL;

ALTER TABLE application_status_history ALTER COLUMN application_id SET NOT NULL;
ALTER TABLE application_status_history DROP COLUMN list_entry_id;
DROP INDEX IF EXISTS idx_status_history_entry;
CREATE INDEX idx_status_history_application ON application_status_history (application_id, changed_at ASC);

-- Re-point interview rounds: add application_id, migrate, drop old FK.
ALTER TABLE interview_rounds ADD COLUMN application_id UUID REFERENCES applications(id) ON DELETE CASCADE;

UPDATE interview_rounds ir
SET application_id = a.id
FROM list_entries le
JOIN user_lists ul ON le.list_id = ul.id
JOIN applications a ON a.user_id = ul.user_id AND a.company_id = le.company_id
WHERE ir.list_entry_id = le.id;

-- Delete orphaned rounds
DELETE FROM interview_rounds WHERE application_id IS NULL;

ALTER TABLE interview_rounds ALTER COLUMN application_id SET NOT NULL;
ALTER TABLE interview_rounds DROP COLUMN list_entry_id;
DROP INDEX IF EXISTS idx_interview_rounds_entry;
CREATE INDEX idx_interview_rounds_application ON interview_rounds (application_id, round_number ASC);

-- Drop application-specific columns from list_entries (they now live on applications).
ALTER TABLE list_entries DROP COLUMN role_title;
ALTER TABLE list_entries DROP COLUMN status;
ALTER TABLE list_entries DROP COLUMN date_applied;
ALTER TABLE list_entries DROP COLUMN notes;
