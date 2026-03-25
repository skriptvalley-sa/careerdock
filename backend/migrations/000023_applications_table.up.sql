-- Create applications table: tracks job applications at the company level, not per list.
-- A user can have multiple applications per company (different roles).
-- IF NOT EXISTS guards make this safe to re-run after a partial failure.
CREATE TABLE IF NOT EXISTS applications (
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

CREATE INDEX IF NOT EXISTS idx_applications_user ON applications (user_id);
CREATE INDEX IF NOT EXISTS idx_applications_company ON applications (company_id);
CREATE INDEX IF NOT EXISTS idx_applications_user_company ON applications (user_id, company_id);

-- Migrate existing application data from list_entries (only if the source columns still exist).
-- For each unique (user_id, company_id), take the entry with the most recent update.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'list_entries' AND column_name = 'status'
  ) THEN
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
    ORDER BY ul.user_id, le.company_id, le.updated_at DESC
    ON CONFLICT DO NOTHING;
  END IF;
END $$;

-- Re-point status history: add application_id column, migrate references, then drop old FK.
ALTER TABLE application_status_history ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id) ON DELETE CASCADE;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'application_status_history' AND column_name = 'list_entry_id'
  ) THEN
    UPDATE application_status_history ash
    SET application_id = a.id
    FROM list_entries le
    JOIN user_lists ul ON le.list_id = ul.id
    JOIN applications a ON a.user_id = ul.user_id AND a.company_id = le.company_id
    WHERE ash.list_entry_id = le.id;

    -- Delete orphaned history rows that couldn't be migrated
    DELETE FROM application_status_history WHERE application_id IS NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'application_status_history' AND column_name = 'application_id'
  ) THEN
    ALTER TABLE application_status_history ALTER COLUMN application_id SET NOT NULL;
  END IF;
END $$;

ALTER TABLE application_status_history DROP COLUMN IF EXISTS list_entry_id;
DROP INDEX IF EXISTS idx_status_history_entry;
CREATE INDEX IF NOT EXISTS idx_status_history_application ON application_status_history (application_id, changed_at ASC);

-- Re-point interview rounds: add application_id, migrate, drop old FK.
ALTER TABLE interview_rounds ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id) ON DELETE CASCADE;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'interview_rounds' AND column_name = 'list_entry_id'
  ) THEN
    UPDATE interview_rounds ir
    SET application_id = a.id
    FROM list_entries le
    JOIN user_lists ul ON le.list_id = ul.id
    JOIN applications a ON a.user_id = ul.user_id AND a.company_id = le.company_id
    WHERE ir.list_entry_id = le.id;

    -- Delete orphaned rounds
    DELETE FROM interview_rounds WHERE application_id IS NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'interview_rounds' AND column_name = 'application_id'
  ) THEN
    ALTER TABLE interview_rounds ALTER COLUMN application_id SET NOT NULL;
  END IF;
END $$;

ALTER TABLE interview_rounds DROP COLUMN IF EXISTS list_entry_id;
DROP INDEX IF EXISTS idx_interview_rounds_entry;
CREATE INDEX IF NOT EXISTS idx_interview_rounds_application ON interview_rounds (application_id, round_number ASC);

-- Drop application-specific columns from list_entries (they now live on applications).
ALTER TABLE list_entries DROP COLUMN IF EXISTS role_title;
ALTER TABLE list_entries DROP COLUMN IF EXISTS status;
ALTER TABLE list_entries DROP COLUMN IF EXISTS date_applied;
ALTER TABLE list_entries DROP COLUMN IF EXISTS notes;
