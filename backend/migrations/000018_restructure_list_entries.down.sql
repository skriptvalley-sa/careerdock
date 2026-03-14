DROP INDEX IF EXISTS uq_list_entries_list_company;

-- Restore role_title NOT NULL (set empty string for any nulls first)
UPDATE list_entries SET role_title = '' WHERE role_title IS NULL;
ALTER TABLE list_entries ALTER COLUMN role_title SET NOT NULL;

-- Remove company_status column
ALTER TABLE list_entries DROP COLUMN IF EXISTS company_status;
