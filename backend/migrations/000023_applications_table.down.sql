-- Restore application columns on list_entries.
ALTER TABLE list_entries ADD COLUMN role_title VARCHAR(255);
ALTER TABLE list_entries ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'not_applied'
    CHECK (status IN ('not_applied','applied','phone_screen','interview','offer','rejected','accepted','withdrawn'));
ALTER TABLE list_entries ADD COLUMN date_applied DATE;
ALTER TABLE list_entries ADD COLUMN notes TEXT;

-- Restore interview_rounds FK to list_entries.
ALTER TABLE interview_rounds ADD COLUMN list_entry_id UUID REFERENCES list_entries(id) ON DELETE CASCADE;
ALTER TABLE interview_rounds DROP COLUMN application_id;
DROP INDEX IF EXISTS idx_interview_rounds_application;

-- Restore status_history FK to list_entries.
ALTER TABLE application_status_history ADD COLUMN list_entry_id UUID REFERENCES list_entries(id) ON DELETE CASCADE;
ALTER TABLE application_status_history DROP COLUMN application_id;
DROP INDEX IF EXISTS idx_status_history_application;

DROP TABLE IF EXISTS applications;
